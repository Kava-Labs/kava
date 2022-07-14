package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/kava-labs/kava/x/liquid/types"
)

// MintDerivative mints a new derivative
func (k Keeper) MintDerivative(ctx sdk.Context, delegatorAddr sdk.AccAddress, valAddr sdk.ValAddress, amount sdk.Coin) error {

	validator, found := k.stakingKeeper.GetValidator(ctx, valAddr)
	if !found {
		return types.ErrNoValidatorFound
	}

	delegation, found := k.stakingKeeper.GetDelegation(ctx, delegatorAddr, valAddr)
	if !found {
		return types.ErrNoDelegatorForAddress
	}

	if amount.Denom != k.stakingKeeper.BondDenom(ctx) {
		return types.ErrOnlyBondDenomAllowedForTokenize
	}

	delegationAmount := validator.Tokens.ToDec().Mul(delegation.GetShares()).Quo(validator.DelegatorShares)
	if amount.Amount.GT(sdk.Int(delegationAmount)) {
		return types.ErrNotEnoughDelegationShares
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

	liquidTokenDenom := k.GetLiquidStakingTokenDenom(ctx, valAddr)
	liquidToken := sdk.NewCoin(liquidTokenDenom, amount.Amount)
	err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.Coins{liquidToken})
	if err != nil {
		return err
	}

	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, delegatorAddr, sdk.Coins{liquidToken})
	if err != nil {
		return err
	}

	// Validate unbond share amount
	shares, err := validator.SharesFromTokens(amount.Amount)
	if err != nil {
		return err
	}

	sharesTruncated, err := validator.SharesFromTokensTruncated(amount.Amount)
	if err != nil {
		return err
	}

	delShares := delegation.GetShares()
	if sharesTruncated.GT(delShares) {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "invalid shares amount")
	}

	// Cap the shares at the delegation's shares. Shares could be greater due to rounding,
	// however we don't truncate shares because we want to allow for full delegation withdraw.
	if shares.GT(delShares) {
		shares = delShares
	}

	returnAmount, err := k.stakingKeeper.Unbond(ctx, delegatorAddr, valAddr, shares)
	if err != nil {
		return err
	}

	if validator.IsBonded() {
		k.bondedTokensToNotBonded(ctx, returnAmount)
	}

	// Note: UndelegateCoinsFromModuleToAccount is internally calling TrackUndelegation for vesting account
	err = k.bankKeeper.UndelegateCoinsFromModuleToAccount(ctx, stakingtypes.NotBondedPoolName, delegatorAddr, sdk.Coins{amount})
	if err != nil {
		return err
	}

	// send coins to module account
	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, delegatorAddr, types.ModuleAccountName, sdk.Coins{amount})
	if err != nil {
		return err
	}

	// delegate from module account
	moduleAccAddress := authtypes.NewModuleAddress(types.ModuleAccountName)
	_, err = k.stakingKeeper.Delegate(ctx, moduleAccAddress, amount.Amount, stakingtypes.Unbonded, validator, true)
	if err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeMintDerivative,
			sdk.NewAttribute(sdk.AttributeKeyAmount, liquidToken.String()),
			sdk.NewAttribute(types.AttributeKeyValidator, validator.String()),
			sdk.NewAttribute(types.AttributeKeyModuleAccount, moduleAccAddress.String()),
		),
	)

	return nil
}

func (k Keeper) GetLiquidStakingTokenDenom(ctx sdk.Context, valAddr sdk.ValAddress) string {
	return types.GetLiquidStakingTokenDenom(k.stakingKeeper.BondDenom(ctx), valAddr)
}

// bondedTokensToNotBonded transfers coins from the bonded to the not bonded pool within staking
func (k Keeper) bondedTokensToNotBonded(ctx sdk.Context, tokens sdk.Int) {
	coins := sdk.NewCoins(sdk.NewCoin(k.stakingKeeper.BondDenom(ctx), tokens))
	if err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, stakingtypes.BondedPoolName, stakingtypes.NotBondedPoolName, coins); err != nil {
		panic(err)
	}
}
