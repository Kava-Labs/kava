package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/liquid/types"
)

// MintDerivative removes a user's staking delegation and mints them equivalent staking derivative coins.
//
// The input staking token amount is used to calculate shares in the user's delegation, which are transferred to a delegation owned by the module.
// Derivative coins are them minted and transferred to the user.
func (k Keeper) MintDerivative(ctx sdk.Context, delegatorAddr sdk.AccAddress, valAddr sdk.ValAddress, amount sdk.Coin) (sdk.Coin, error) {
	bondDenom := k.stakingKeeper.BondDenom(ctx)
	if amount.Denom != bondDenom {
		return sdk.Coin{}, sdkerrors.Wrapf(types.ErrInvalidDenom, "expected %s", bondDenom)
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

// CalculateDerivativeSharesFromTokens converts a staking token amount into its equivalent delegation shares, and staking derivative amount.
// This combines the code for calculating the shares to be transferred, and the derivative coins to be minted.
func (k Keeper) CalculateDerivativeSharesFromTokens(ctx sdk.Context, delegator sdk.AccAddress, validator sdk.ValAddress, tokens sdk.Int) (sdk.Int, sdk.Dec, error) {
	if !tokens.IsPositive() {
		return sdk.Int{}, sdk.Dec{}, sdkerrors.Wrap(types.ErrUntransferableShares, "token amount must be positive")
	}
	shares, err := k.stakingKeeper.ValidateUnbondAmount(ctx, delegator, validator, tokens)
	if err != nil {
		return sdk.Int{}, sdk.Dec{}, err
	}
	return shares.TruncateInt(), shares, nil
}

// BurnDerivative burns an user's staking derivative coins and returns them an equivalent staking delegation.
//
// The derivative coins are burned, and an equivalent number of shares in the module's staking delegation are transferred back to the user.
func (k Keeper) BurnDerivative(ctx sdk.Context, delegatorAddr sdk.AccAddress, valAddr sdk.ValAddress, amount sdk.Coin) (sdk.Dec, error) {

	if amount.Denom != k.GetLiquidStakingTokenDenom(valAddr) {
		return sdk.Dec{}, sdkerrors.Wrap(types.ErrInvalidDenom, "derivative denom does not match validator")
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

// IsDerivativeDenom returns true if the denom is a valid derivative denom and
// corresponds to a valid validator.
func (k Keeper) IsDerivativeDenom(ctx sdk.Context, denom string) bool {
	valAddr, err := types.ParseLiquidStakingTokenDenom(denom)
	if err != nil {
		return false
	}

	_, found := k.stakingKeeper.GetValidator(ctx, valAddr)
	return found
}

// GetKavaForDerivatives returns the total amount of the provided derivatives
// in Kava accounting for the specific share prices.
func (k Keeper) GetKavaForDerivatives(ctx sdk.Context, coins sdk.Coins) (sdk.Int, error) {
	totalKava := sdk.ZeroInt()

	for _, coin := range coins {
		valAddr, err := types.ParseLiquidStakingTokenDenom(coin.Denom)
		if err != nil {
			return sdk.Int{}, fmt.Errorf("invalid derivative denom: %w", err)
		}

		validator, found := k.stakingKeeper.GetValidator(ctx, valAddr)
		if !found {
			return sdk.Int{}, fmt.Errorf("invalid derivative denom %s: validator not found", coin.Denom)
		}

		// bkava is 1:1 to delegation shares
		valTokens := validator.TokensFromSharesTruncated(coin.Amount.ToDec())
		totalKava = totalKava.Add(valTokens.TruncateInt())
	}

	return totalKava, nil
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
