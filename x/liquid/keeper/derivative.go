package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"

	"github.com/kava-labs/kava/x/liquid/types"
)

// MintDerivative mints a new derivative
func (k Keeper) MintDerivative(ctx sdk.Context, delegatorAddr sdk.AccAddress, valAddr sdk.ValAddress, amount sdk.Coin) error {

	if amount.Denom != k.stakingKeeper.BondDenom(ctx) {
		return types.ErrOnlyBondDenomAllowedForTokenize
	}

	acc := k.accountKeeper.GetAccount(ctx, delegatorAddr)
	if acc != nil {
		acc, ok := acc.(vesting.VestingAccount)
		if ok {
			// if account is a vesting account, it checks if free delegation (non-vesting delegation) is not exceeding
			// the tokenize share amount and execute further tokenize share process
			// tokenize share is reducing unlocked tokens delegation from the vesting account and further process
			// is not causing issues
			delFree := acc.GetDelegatedFree().AmountOf(amount.Denom)
			if delFree.LT(amount.Amount) {
				return types.ErrExceedingFreeVestingDelegations
			}
		}
	}

	derivativeAmount, shares, err := k.CalculateDerivativeSharesFromTokens(ctx, delegatorAddr, valAddr, amount.Amount)
	if err != nil {
		return err
	}

	moduleAccAddress := authtypes.NewModuleAddress(types.ModuleAccountName)
	if err := k.TransferDelegation(ctx, valAddr, delegatorAddr, moduleAccAddress, shares); err != nil {
		return err
	}

	liquidTokenDenom := k.GetLiquidStakingTokenDenom(ctx, valAddr)
	liquidToken := sdk.NewCoins(sdk.NewCoin(liquidTokenDenom, derivativeAmount))
	if err = k.mintCoins(ctx, delegatorAddr, liquidToken); err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeMintDerivative,
			sdk.NewAttribute(types.AttributeKeyDelegator, delegatorAddr.String()),
			sdk.NewAttribute(types.AttributeKeyValidator, valAddr.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, liquidToken.String()),
			sdk.NewAttribute(types.AttributeKeySharesTransferred, shares.String()),
		),
	)

	return nil
}

func (k Keeper) CalculateDerivativeSharesFromTokens(ctx sdk.Context, delegator sdk.AccAddress, validator sdk.ValAddress, tokens sdk.Int) (sdk.Int, sdk.Dec, error) {
	shares, err := k.stakingKeeper.ValidateUnbondAmount(ctx, delegator, validator, tokens)
	if err != nil {
		// TODO wrap staking errors
		return sdk.Int{}, sdk.Dec{}, sdkerrors.Wrap(types.ErrInvalidMint, err.Error())
	}
	return shares.TruncateInt(), shares, nil
}

func (k Keeper) mintCoins(ctx sdk.Context, receiver sdk.AccAddress, amount sdk.Coins) error {
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleAccountName, amount); err != nil {
		return err
	}
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleAccountName, receiver, amount); err != nil {
		return err
	}
	return nil
}

// BurnDerivative burns an existing derivative
func (k Keeper) BurnDerivative(ctx sdk.Context, delegatorAddr sdk.AccAddress, valAddr sdk.ValAddress, amount sdk.Coin) error {

	// User must have enough tokens to fulfill redeem request
	balance := k.bankKeeper.GetBalance(ctx, delegatorAddr, amount.Denom)
	if balance.Amount.LT(amount.Amount) {
		return types.ErrNotEnoughBalance
	}

	// Confirm that the coin's denom matches the validator's specific liquidate staking coin denom
	coinDenom := k.GetLiquidStakingTokenDenom(ctx, valAddr)
	if coinDenom != amount.Denom {
		return types.ErrInvalidDerivativeDenom
	}

	// Calculate the ratio between shares and redeem amount:
	// (moduleAccountTotalDelegation * redeemAmount) / totalIssue
	maccAddr := k.accountKeeper.GetModuleAddress(types.ModuleAccountName)
	delegation, found := k.stakingKeeper.GetDelegation(ctx, maccAddr, valAddr)
	if !found {
		return types.ErrNotEnoughDelegationShares
	}
	shareDenomSupply := k.bankKeeper.GetSupply(ctx, amount.Denom)
	shares := delegation.Shares.Mul(amount.Amount.ToDec()).QuoInt(shareDenomSupply.Amount)

	// Burn share amount from user's address
	err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, delegatorAddr, types.ModuleAccountName, sdk.NewCoins(amount))
	if err != nil {
		return err
	}
	err = k.bankKeeper.BurnCoins(ctx, types.ModuleAccountName, sdk.NewCoins(amount))
	if err != nil {
		return err
	}

	if err := k.TransferDelegation(ctx, valAddr, maccAddr, delegatorAddr, shares); err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeBurnDerivative,
			sdk.NewAttribute(types.AttributeKeyAmountBurned, amount.String()),
			// sdk.NewAttribute(types.AttributeKeyAmountReturned, returnCoin.String()), // TODO
			sdk.NewAttribute(types.AttributeKeyModuleAccount, maccAddr.String()),
		),
	)

	return nil
}

func (k Keeper) GetLiquidStakingTokenDenom(ctx sdk.Context, valAddr sdk.ValAddress) string {
	return types.GetLiquidStakingTokenDenom(k.derivativeDenom, valAddr)
}
