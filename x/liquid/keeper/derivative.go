package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/liquid/types"
)

// MintDerivative mints a new derivative
func (k Keeper) MintDerivative(ctx sdk.Context, delegatorAddr sdk.AccAddress, valAddr sdk.ValAddress, amount sdk.Coin) (sdk.Coin, error) {

	if amount.Denom != k.stakingKeeper.BondDenom(ctx) {
		return sdk.Coin{}, types.ErrOnlyBondDenomAllowedForTokenize
	}

	derivativeAmount, shares, err := k.CalculateDerivativeSharesFromTokens(ctx, delegatorAddr, valAddr, amount.Amount)
	if err != nil {
		return sdk.Coin{}, err
	}

	// Fetching the module account will create it if it doesn't exist.
	// This is necessary as otherwise TransferDelegation will create a normal account.
	modAcc := k.accountKeeper.GetModuleAccount(ctx, types.ModuleAccountName)
	if _, err := k.TransferDelegation(ctx, valAddr, delegatorAddr, modAcc.GetAddress(), shares); err != nil {
		return sdk.Coin{}, err
	}

	liquidTokenDenom := k.GetLiquidStakingTokenDenom(valAddr)
	liquidToken := sdk.NewCoin(liquidTokenDenom, derivativeAmount)
	if err = k.mintCoins(ctx, delegatorAddr, sdk.NewCoins(liquidToken)); err != nil {
		return sdk.Coin{}, err
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

	return liquidToken, nil
}

func (k Keeper) CalculateDerivativeSharesFromTokens(ctx sdk.Context, delegator sdk.AccAddress, validator sdk.ValAddress, tokens sdk.Int) (sdk.Int, sdk.Dec, error) {
	shares, err := k.stakingKeeper.ValidateUnbondAmount(ctx, delegator, validator, tokens)
	if err != nil {
		// TODO wrap staking errors
		return sdk.Int{}, sdk.Dec{}, err
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

func (k Keeper) burnCoins(ctx sdk.Context, sender sdk.AccAddress, amount sdk.Coins) error {
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleAccountName, amount); err != nil {
		return err
	}
	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleAccountName, amount); err != nil {
		return err
	}
	return nil
}

// BurnDerivative burns an user's derivative coins and returns the original delegation.
func (k Keeper) BurnDerivative(ctx sdk.Context, delegatorAddr sdk.AccAddress, valAddr sdk.ValAddress, amount sdk.Coin) (sdk.Dec, error) {

	if amount.Denom != k.GetLiquidStakingTokenDenom(valAddr) {
		return sdk.Dec{}, types.ErrInvalidDerivativeDenom
	}

	if err := k.burnCoins(ctx, delegatorAddr, sdk.NewCoins(amount)); err != nil {
		return sdk.Dec{}, err
	}

	modAcc := k.accountKeeper.GetModuleAccount(ctx, types.ModuleAccountName)
	shares := amount.Amount.ToDec()
	receivedShares, err := k.TransferDelegation(ctx, valAddr, modAcc.GetAddress(), delegatorAddr, shares)
	if err != nil {
		return sdk.Dec{}, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeBurnDerivative,
			sdk.NewAttribute(types.AttributeKeyDelegator, delegatorAddr.String()),
			sdk.NewAttribute(types.AttributeKeyValidator, valAddr.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, amount.String()),
			sdk.NewAttribute(types.AttributeKeySharesTransferred, shares.String()),
		),
	)
	return receivedShares, nil
}

func (k Keeper) GetLiquidStakingTokenDenom(valAddr sdk.ValAddress) string {
	return types.GetLiquidStakingTokenDenom(k.derivativeDenom, valAddr)
}

func (k Keeper) TokenToDerivative(ctx sdk.Context, valAddr sdk.ValAddress, amount sdk.Int) (sdk.Coin, error) {
	modAcc := k.accountKeeper.GetModuleAccount(ctx, types.ModuleAccountName) // TODO don't create account
	derivative, _, err := k.CalculateDerivativeSharesFromTokens(ctx, modAcc.GetAddress(), valAddr, amount)
	if err != nil {
		return sdk.Coin{}, nil
	}
	liquidTokenDenom := k.GetLiquidStakingTokenDenom(valAddr)
	liquidToken := sdk.NewCoin(liquidTokenDenom, derivative)
	return liquidToken, nil
}
