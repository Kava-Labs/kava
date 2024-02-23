package keeper

import (
	"math/big"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/evmutil/types"
)

// MintConversionPairCoin mints the given amount of a ConversionPair denom and
// sends it to the provided address.
func (k Keeper) MintConversionPairCoin(
	ctx sdk.Context,
	pair types.ConversionPair,
	amount *big.Int,
	recipient sdk.AccAddress,
) (sdk.Coin, error) {
	coin := sdk.NewCoin(pair.Denom, sdkmath.NewIntFromBigInt(amount))
	coins := sdk.NewCoins(coin)

	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, coins); err != nil {
		return sdk.Coin{}, err
	}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipient, coins); err != nil {
		return sdk.Coin{}, err
	}

	return coin, nil
}

// BurnConversionPairCoin transfers the provided amount to the module account
// then burns it.
func (k Keeper) BurnConversionPairCoin(
	ctx sdk.Context,
	pair types.ConversionPair,
	coin sdk.Coin,
	account sdk.AccAddress,
) error {
	coins := sdk.NewCoins(coin)

	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, account, types.ModuleName, coins); err != nil {
		return err
	}

	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, coins); err != nil {
		return err
	}

	return nil
}

// ConvertCoinToERC20 converts an sdk.Coin from the originating account to an
// ERC20 to the receiver account.
func (k Keeper) ConvertCoinToERC20(
	ctx sdk.Context,
	initiatorAccount sdk.AccAddress,
	receiverAccount types.InternalEVMAddress,
	coin sdk.Coin,
) error {
	pair, err := k.GetEnabledConversionPairFromDenom(ctx, coin.Denom)
	if err != nil {
		// Coin not in enabled conversion pair list
		return err
	}

	// compute coin.Amount * 10^-pair.DecimalConversion
	evmAmount, remainder := MulByPowersTen(coin.Amount.BigInt(), int64(-1*pair.DecimalConversion))
	if remainder.Cmp(new(big.Int)) != 0 {
		return errorsmod.Wrapf(types.ErrInvalidConversionAmount, "%s not divisble by 10^%d", coin.Amount, pair.DecimalConversion)
	}

	if err := k.BurnConversionPairCoin(ctx, pair, coin, initiatorAccount); err != nil {
		return err
	}

	if err := k.UnlockERC20Tokens(ctx, pair, evmAmount, receiverAccount); err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeConvertCoinToERC20,
		sdk.NewAttribute(types.AttributeKeyInitiator, initiatorAccount.String()),
		sdk.NewAttribute(types.AttributeKeyReceiver, receiverAccount.String()),
		sdk.NewAttribute(types.AttributeKeyERC20Address, pair.GetAddress().String()),
		sdk.NewAttribute(types.AttributeKeyOutputAmount, evmAmount.String()),
		sdk.NewAttribute(types.AttributeKeyInputAmount, coin.Amount.String()),
	))

	return nil
}

// ConvertERC20ToCoin converts an ERC20 coin from the originating account to an
// sdk.Coin to the receiver account.
func (k Keeper) ConvertERC20ToCoin(
	ctx sdk.Context,
	initiator types.InternalEVMAddress,
	receiver sdk.AccAddress,
	contractAddr types.InternalEVMAddress,
	amount sdkmath.Int,
) error {
	// Check that the contract is enabled to convert to coin
	pair, err := k.GetEnabledConversionPairFromERC20Address(ctx, contractAddr)
	if err != nil {
		// contract not in enabled conversion pair list
		return err
	}

	// compute amount * 10^pair.DecimalConversion
	// If DecimalConversion is negative the output will be smaller than the input, with a potential decimal part.
	// This just errors if there is a decimal part, so users must pad their conversion amounts with zeros.
	// Alternatively we could only transfer the non decimal part, leaving the rest on the evm side.
	sdkamount, remainder := MulByPowersTen(amount.BigInt(), int64(pair.DecimalConversion))
	if remainder.Cmp(new(big.Int)) != 0 {
		return errorsmod.Wrapf(types.ErrInvalidConversionAmount, "%s not divisble by 10^%d", amount, pair.DecimalConversion)
	}

	// lock erc20 tokens
	if err := k.LockERC20Tokens(ctx, pair, amount.BigInt(), initiator); err != nil {
		return err
	}

	// mint conversion pair coin
	_, err = k.MintConversionPairCoin(ctx, pair, sdkamount, receiver)
	if err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeConvertERC20ToCoin,
		sdk.NewAttribute(types.AttributeKeyERC20Address, contractAddr.String()),
		sdk.NewAttribute(types.AttributeKeyInitiator, initiator.String()),
		sdk.NewAttribute(types.AttributeKeyReceiver, receiver.String()),
		sdk.NewAttribute(types.AttributeKeyOutputAmount, sdkamount.String()),
		sdk.NewAttribute(types.AttributeKeyInputAmount, amount.String()),
	))

	return nil
}

// UnlockERC20Tokens transfers the given amount of a conversion pair ERC20 token
// to the provided account.
func (k Keeper) UnlockERC20Tokens(
	ctx sdk.Context,
	pair types.ConversionPair,
	amount *big.Int,
	receiver types.InternalEVMAddress,
) error {
	contractAddr := pair.GetAddress()
	startBal, err := k.QueryERC20BalanceOf(ctx, contractAddr, receiver)
	if err != nil {
		return errorsmod.Wrapf(types.ErrEVMCall, "failed to retrieve balance %s", err.Error())
	}
	res, err := k.CallEVM(
		ctx,
		types.ERC20MintableBurnableContract.ABI, // abi
		types.ModuleEVMAddress,                  // from addr
		pair.GetAddress(),                       // contract addr
		"transfer",                              // method
		// Transfer ERC20 args
		receiver.Address,
		amount,
	)
	if err != nil {
		return err
	}

	// validate end bal
	endBal, err := k.QueryERC20BalanceOf(ctx, contractAddr, receiver)
	if err != nil {
		return errorsmod.Wrapf(types.ErrEVMCall, "failed to retrieve balance %s", err.Error())
	}
	expectedEndBal := big.NewInt(0).Add(startBal, amount)
	if expectedEndBal.Cmp(endBal) != 0 {
		return errorsmod.Wrapf(
			types.ErrBalanceInvariance,
			"invalid token balance - expected: %v, actual: %v",
			expectedEndBal, endBal,
		)
	}

	// Check for unexpected `Approval` event in logs
	if err := k.monitorApprovalEvent(res); err != nil {
		return err
	}

	return err
}

// LockERC20Tokens transfers the given amount of a conversion pair ERC20 token
// from the initiator account to the module account.
func (k Keeper) LockERC20Tokens(
	ctx sdk.Context,
	pair types.ConversionPair,
	amount *big.Int,
	initiator types.InternalEVMAddress,
) error {
	contractAddr := pair.GetAddress()
	initiatorStartBal, err := k.QueryERC20BalanceOf(ctx, contractAddr, initiator)
	if err != nil {
		return errorsmod.Wrapf(types.ErrEVMCall, "failed to retrieve balance: %s", err.Error())
	}

	res, err := k.CallEVM(
		ctx,
		types.ERC20MintableBurnableContract.ABI, // abi
		initiator.Address,                       // from addr
		contractAddr,                            // contract addr
		"transfer",                              // method
		// Transfer ERC20 args
		types.ModuleEVMAddress,
		amount,
	)
	if err != nil {
		return err
	}

	// validate end bal
	initiatorEndBal, err := k.QueryERC20BalanceOf(ctx, contractAddr, initiator)
	if err != nil {
		return errorsmod.Wrapf(types.ErrEVMCall, "failed to retrieve balance %s", err.Error())
	}
	expectedEndBal := big.NewInt(0).Sub(initiatorStartBal, amount)
	if expectedEndBal.Cmp(initiatorEndBal) != 0 {
		return errorsmod.Wrapf(
			types.ErrBalanceInvariance,
			"invalid token balance - expected: %v, actual: %v",
			expectedEndBal, initiatorEndBal,
		)
	}

	// Check for unexpected `Approval` event in logs
	if err := k.monitorApprovalEvent(res); err != nil {
		return err
	}

	return err
}

// MulByPowersTen computes i*10^power and returns integers for the part before and after the decimal point.
func MulByPowersTen(i *big.Int, power int64) (*big.Int, *big.Int) {
	positivePower := big.NewInt(power)
	positivePower.Abs(positivePower)

	multiple := new(big.Int).Exp(big.NewInt(10), positivePower, nil)

	if power < 0 {
		return new(big.Int).Mul(i, multiple), big.NewInt(0)
	} else {
		intPart := new(big.Int).Div(i, multiple)
		decimalPart := new(big.Int).Mod(i, multiple)
		return intPart, decimalPart
	}
}
