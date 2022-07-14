package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/kava-labs/kava/x/liquid/types"
)

// MintDerivative mints a new derivative
func (k Keeper) MintDerivative(ctx sdk.Context, delegatorAddr sdk.AccAddress, valAddr sdk.ValAddress, shares sdk.Dec) error {

	validator, found := k.stakingKeeper.GetValidator(ctx, valAddr)
	if !found {
		return types.ErrNoValidatorFound
	}

	delegation, found := k.stakingKeeper.GetDelegation(ctx, delegatorAddr, valAddr)
	if !found {
		return types.ErrNoDelegatorForAddress
	}

	delegationAmount := validator.Tokens.ToDec().Mul(delegation.GetShares()).Quo(validator.DelegatorShares)
	if shares.GT(delegationAmount) {
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
			delFree := acc.GetDelegatedFree().AmountOf(k.stakingKeeper.BondDenom(ctx))
			if delFree.LT(shares.RoundInt()) {
				return types.ErrExceedingFreeVestingDelegations
			}
		}
	}

	liquidCoinDenom := k.GetLiquidStakingTokenDenom(ctx, valAddr)
	liquidCoin := sdk.NewCoin(liquidCoinDenom, shares.RoundInt())

	returnAmount, err := k.stakingKeeper.Unbond(ctx, delegatorAddr, valAddr, shares)
	if err != nil {
		return err
	}
	returnCoins := sdk.NewCoins(sdk.NewCoin(k.stakingKeeper.BondDenom(ctx), returnAmount))

	if validator.IsBonded() {
		k.bondedTokensToNotBonded(ctx, returnAmount)
	}

	// Note: UndelegateCoinsFromModuleToAccount is internally calling TrackUndelegation for vesting account
	err = k.bankKeeper.UndelegateCoinsFromModuleToAccount(ctx, stakingtypes.NotBondedPoolName, delegatorAddr, returnCoins)
	if err != nil {
		return err
	}

	// send coins to module account
	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, delegatorAddr, types.ModuleAccountName, returnCoins)
	if err != nil {
		return err
	}

	validator, found = k.stakingKeeper.GetValidator(ctx, valAddr)
	if !found {
		return types.ErrNoValidatorFound
	}

	// delegate from module account
	moduleAccAddress := authtypes.NewModuleAddress(types.ModuleAccountName)
	_, err = k.stakingKeeper.Delegate(ctx, moduleAccAddress, returnAmount, stakingtypes.Unbonded, validator, true)
	if err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeMintDerivative,
			sdk.NewAttribute(sdk.AttributeKeyAmount, liquidCoin.String()),
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
