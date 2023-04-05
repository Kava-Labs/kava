package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/liquid/types"
)

// MintDerivative removes a user's staking delegation and mints them equivalent staking derivative coins.
//
// The input staking token amount is used to calculate shares in the user's delegation, which are transferred to a delegation owned by the module.
// Derivative coins are them minted and transferred to the user.
func (k Keeper) MintDerivative(ctx sdk.Context, delegatorAddr sdk.AccAddress, valAddr sdk.ValAddress, amount sdk.Coin) (sdk.Coin, error) {
	bondDenom := k.stakingKeeper.BondDenom(ctx)
	if amount.Denom != bondDenom {
		return sdk.Coin{}, errorsmod.Wrapf(types.ErrInvalidDenom, "expected %s", bondDenom)
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
func (k Keeper) CalculateDerivativeSharesFromTokens(ctx sdk.Context, delegator sdk.AccAddress, validator sdk.ValAddress, tokens sdkmath.Int) (sdkmath.Int, sdk.Dec, error) {
	if !tokens.IsPositive() {
		return sdkmath.Int{}, sdk.Dec{}, errorsmod.Wrap(types.ErrUntransferableShares, "token amount must be positive")
	}
	shares, err := k.stakingKeeper.ValidateUnbondAmount(ctx, delegator, validator, tokens)
	if err != nil {
		return sdkmath.Int{}, sdk.Dec{}, err
	}
	return shares.TruncateInt(), shares, nil
}

// BurnDerivative burns an user's staking derivative coins and returns them an equivalent staking delegation.
//
// The derivative coins are burned, and an equivalent number of shares in the module's staking delegation are transferred back to the user.
func (k Keeper) BurnDerivative(ctx sdk.Context, delegatorAddr sdk.AccAddress, valAddr sdk.ValAddress, amount sdk.Coin) (sdk.Dec, error) {

	if amount.Denom != k.GetLiquidStakingTokenDenom(valAddr) {
		return sdk.Dec{}, errorsmod.Wrap(types.ErrInvalidDenom, "derivative denom does not match validator")
	}

	if err := k.burnCoins(ctx, delegatorAddr, sdk.NewCoins(amount)); err != nil {
		return sdk.Dec{}, err
	}

	modAcc := k.accountKeeper.GetModuleAccount(ctx, types.ModuleAccountName)
	shares := sdk.NewDecFromInt(amount.Amount)
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

// GetStakedTokensForDerivatives returns the total value of the provided derivatives
// in staked tokens, accounting for the specific share prices.
func (k Keeper) GetStakedTokensForDerivatives(ctx sdk.Context, coins sdk.Coins) (sdk.Coin, error) {
	total := sdk.ZeroInt()

	for _, coin := range coins {
		valAddr, err := types.ParseLiquidStakingTokenDenom(coin.Denom)
		if err != nil {
			return sdk.Coin{}, fmt.Errorf("invalid derivative denom: %w", err)
		}

		validator, found := k.stakingKeeper.GetValidator(ctx, valAddr)
		if !found {
			return sdk.Coin{}, fmt.Errorf("invalid derivative denom %s: validator not found", coin.Denom)
		}

		// bkava is 1:1 to delegation shares
		valTokens := validator.TokensFromSharesTruncated(sdk.NewDecFromInt(coin.Amount))
		total = total.Add(valTokens.TruncateInt())
	}

	totalCoin := sdk.NewCoin(k.stakingKeeper.BondDenom(ctx), total)
	return totalCoin, nil
}

// GetTotalDerivativeValue returns the total sum value of all derivative coins
// for all validators denominated by the bond token (ukava).
func (k Keeper) GetTotalDerivativeValue(ctx sdk.Context) (sdk.Coin, error) {
	bkavaCoins := sdk.NewCoins()

	k.bankKeeper.IterateTotalSupply(ctx, func(c sdk.Coin) bool {
		if k.IsDerivativeDenom(ctx, c.Denom) {
			bkavaCoins = bkavaCoins.Add(c)
		}

		return false
	})

	return k.GetStakedTokensForDerivatives(ctx, bkavaCoins)
}

// GetDerivativeValue returns the total underlying value of the provided
// derivative denominated by the bond token (ukava).
func (k Keeper) GetDerivativeValue(ctx sdk.Context, denom string) (sdk.Coin, error) {
	return k.GetStakedTokensForDerivatives(ctx, sdk.NewCoins(k.bankKeeper.GetSupply(ctx, denom)))
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

// DerivativeFromTokens calculates the approximate amount of derivative coins that would be minted for a given amount of staking tokens.
func (k Keeper) DerivativeFromTokens(ctx sdk.Context, valAddr sdk.ValAddress, tokens sdk.Coin) (sdk.Coin, error) {
	bondDenom := k.stakingKeeper.BondDenom(ctx)
	if tokens.Denom != bondDenom {
		return sdk.Coin{}, errorsmod.Wrapf(types.ErrInvalidDenom, "'%s' does not match staking denom '%s'", tokens.Denom, bondDenom)
	}

	// Use GetModuleAddress instead of GetModuleAccount to avoid creating a module account if it doesn't exist.
	modAddress := k.accountKeeper.GetModuleAddress(types.ModuleAccountName)
	derivative, _, err := k.CalculateDerivativeSharesFromTokens(ctx, modAddress, valAddr, tokens.Amount)
	if err != nil {
		return sdk.Coin{}, err
	}
	liquidTokenDenom := k.GetLiquidStakingTokenDenom(valAddr)
	liquidToken := sdk.NewCoin(liquidTokenDenom, derivative)
	return liquidToken, nil
}
