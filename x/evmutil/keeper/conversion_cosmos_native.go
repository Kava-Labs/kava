package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/evmutil/types"
)

// ConvertCosmosCoinToERC20 locks the initiator's sdk.Coin in the module account
// and mints the receiver a corresponding amount of an ERC20 representing the Coin.
// If a conversion has never been made before and no contract exists, one will be deployed.
// Only denoms registered to the AllowedCosmosDenoms param may be converted.
func (k *Keeper) ConvertCosmosCoinToERC20(
	ctx sdk.Context,
	initiator sdk.AccAddress,
	receiver types.InternalEVMAddress,
	amount sdk.Coin,
) error {
	// check that the conversion is allowed
	tokenInfo, allowed := k.GetAllowedTokenMetadata(ctx, amount.Denom)
	if !allowed {
		return errorsmod.Wrapf(types.ErrSDKConversionNotEnabled, amount.Denom)
	}

	// send coins from initiator to the module account
	// do this before possible contract deploy to prevent unnecessary store interactions
	err := k.bankKeeper.SendCoinsFromAccountToModule(
		ctx, initiator, types.ModuleName, sdk.NewCoins(amount),
	)
	if err != nil {
		return err
	}

	// find deployed contract if it exits
	contractAddress, err := k.GetOrDeployCosmosCoinERC20Contract(ctx, tokenInfo)
	if err != nil {
		return err
	}

	// mint erc20 tokens for the user
	err = k.MintERC20(ctx, contractAddress, receiver, amount.Amount.BigInt())
	if err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeConvertCosmosCoinToERC20,
		sdk.NewAttribute(types.AttributeKeyInitiator, initiator.String()),
		sdk.NewAttribute(types.AttributeKeyReceiver, receiver.String()),
		sdk.NewAttribute(types.AttributeKeyERC20Address, contractAddress.Hex()),
		sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
	))

	return nil
}

// ConvertCosmosCoinFromERC20 burns the ERC20 wrapper of the cosmos coin and
// sends the underlying sdk coin form the module account to the receiver.
func (k *Keeper) ConvertCosmosCoinFromERC20(
	ctx sdk.Context,
	initiator types.InternalEVMAddress,
	receiver sdk.AccAddress,
	coin sdk.Coin,
) error {
	amount := coin.Amount.BigInt()
	// get deployed contract
	contractAddress, found := k.GetDeployedCosmosCoinContract(ctx, coin.Denom)
	if !found {
		// no contract deployed
		return errorsmod.Wrapf(types.ErrInvalidCosmosDenom, fmt.Sprintf("no erc20 contract found for %s", coin.Denom))
	}

	// verify sufficient balance
	balance, err := k.QueryERC20BalanceOf(ctx, contractAddress, initiator)
	if err != nil {
		return errorsmod.Wrapf(types.ErrEVMCall, "failed to retrieve balance %s", err.Error())
	}
	if balance.Cmp(amount) == -1 {
		return errorsmod.Wrapf(sdkerrors.ErrInsufficientFunds, "failed to convert to cosmos coins")
	}

	// burn initiator's ERC20 tokens
	err = k.BurnERC20(ctx, contractAddress, initiator, amount)
	if err != nil {
		return err
	}

	// send sdk coins to receiver, unlocking them from the module account
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, receiver, sdk.NewCoins(coin))
	if err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeConvertCosmosCoinFromERC20,
		sdk.NewAttribute(types.AttributeKeyInitiator, initiator.String()),
		sdk.NewAttribute(types.AttributeKeyReceiver, receiver.String()),
		sdk.NewAttribute(types.AttributeKeyERC20Address, contractAddress.Hex()),
		sdk.NewAttribute(types.AttributeKeyAmount, coin.String()),
	))

	return nil
}
