package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/kava-labs/kava/x/liquidstaking/types"
)

// MintDerivative mints a new derivative
func (k Keeper) MintDerivative(ctx sdk.Context, delegatorAddr sdk.AccAddress, valAddr sdk.ValAddress, coin sdk.Coin) error {

	_, found := k.stakingKeeper.GetValidator(ctx, valAddr)
	if !found {
		return types.ErrNoValidatorFound
	}

	validator, found := k.stakingKeeper.GetValidator(ctx, valAddr)
	if !found {
		return types.ErrNoValidatorFound
	}

	delegation, found := k.stakingKeeper.GetDelegation(ctx, delegatorAddr, valAddr)
	if !found {
		return types.ErrNoDelegatorForAddress
	}

	if coin.Denom != k.stakingKeeper.BondDenom(ctx) {
		return types.ErrOnlyBondDenomAllowdForTokenize
	}

	delegationAmount := validator.Tokens.ToDec().Mul(delegation.GetShares()).Quo(validator.DelegatorShares)
	if coin.Amount.GT(sdk.Int(delegationAmount)) {
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
			delFree := acc.GetDelegatedFree().AmountOf(coin.Denom)
			if delFree.LT(coin.Amount) {
				return types.ErrExceedingFreeVestingDelegations
			}
		}
	}

	liquidCoinDenom := k.GetLiquidStakingTokenDenom(ctx, valAddr)
	liquidCoin := sdk.NewCoin(liquidCoinDenom, coin.Amount)

	delegationHolder, found := k.GetDelegationHolder(ctx, valAddr)
	if !found {
		// Generate specific module account address for this valAddr
		modAccName := authtypes.NewModuleAddress(fmt.Sprintf("%s:%s", types.ModuleName, valAddr.String()))
		modBaseAcc := authtypes.NewBaseAccount(modAccName, nil, 0, 0)
		modAcc := authtypes.NewModuleAccount(modBaseAcc, types.ModuleName, []string{}...)
		k.accountKeeper.SetModuleAccount(ctx, modAcc)

		delegationHolder = types.NewDelegationHolder(modAcc.GetAddress(), valAddr, liquidCoin)
	} else {
		// Check that the existing delegation liquid staking token denom matches
		if delegationHolder.Delegations.Denom != liquidCoinDenom {
			return types.ErrInvalidLiquidCoinDenom
		}
		// Add the liquid staking coin to the delegations
		delegationHolder.Delegations = delegationHolder.Delegations.Add(liquidCoin)
	}

	err := k.bankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.Coins{liquidCoin})
	if err != nil {
		return err
	}

	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, delegatorAddr, sdk.Coins{liquidCoin})
	if err != nil {
		return err
	}

	shares, err := k.stakingKeeper.ValidateUnbondAmount(
		ctx, delegatorAddr, valAddr, coin.Amount,
	)

	returnAmount, err := k.stakingKeeper.Unbond(ctx, delegatorAddr, valAddr, shares)
	if err != nil {
		return err
	}

	if validator.IsBonded() {
		k.bondedTokensToNotBonded(ctx, returnAmount)
	}

	// Note: UndelegateCoinsFromModuleToAccount is internally calling TrackUndelegation for vesting account
	// TODO: sort out types.BondedPoolName vs. types.NotBondedPoolName
	err = k.bankKeeper.UndelegateCoinsFromModuleToAccount(ctx, types.NotBondedPoolName, delegatorAddr, sdk.Coins{liquidCoin})
	if err != nil {
		return err
	}

	// send coins to module account
	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, delegatorAddr, delegationHolder.ModuleAccount.String(), sdk.Coins{liquidCoin})
	if err != nil {
		return err
	}

	validator, found = k.stakingKeeper.GetValidator(ctx, valAddr)
	if !found {
		return types.ErrNoValidatorFound
	}

	// delegate from module account
	_, err = k.stakingKeeper.Delegate(ctx, delegationHolder.ModuleAccount, coin.Amount, stakingtypes.Unbonded, validator, true)
	if err != nil {
		return err
	}

	k.SetDelegationHolder(ctx, delegationHolder)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeMintDerivative,
			sdk.NewAttribute(sdk.AttributeKeyAmount, coin.String()),
			sdk.NewAttribute(types.AttributeKeyValidator, delegationHolder.Validator.String()),
		),
	)

	return nil
}

// BurnDerivative burns an existing derivative
func (k Keeper) BurnDerivative(ctx sdk.Context, validator sdk.ValAddress, coin sdk.Coin) error {

	// ctx.EventManager().EmitEvent(
	// 	sdk.NewEvent(
	// 		types.EventTypeMintDerivative,
	// 		sdk.NewAttribute(sdk.AttributeKeyAmount, coin.String()),
	// 		sdk.NewAttribute(types.AttributeKeyValidator, derivative.Validator.String()),
	// 	),
	// )

	return nil
}

func (k Keeper) GetLiquidStakingTokenDenom(ctx sdk.Context, valAddr sdk.ValAddress) string {
	return fmt.Sprintf("%s-%s", k.stakingKeeper.BondDenom(ctx), valAddr.String())
}

// bondedTokensToNotBonded transfers coins from the bonded to the not bonded pool within staking
func (k Keeper) bondedTokensToNotBonded(ctx sdk.Context, tokens sdk.Int) {
	coins := sdk.NewCoins(sdk.NewCoin(k.stakingKeeper.BondDenom(ctx), tokens))
	// TODO: sort out types.BondedPoolName vs. types.NotBondedPoolName
	if err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.BondedPoolName, types.NotBondedPoolName, coins); err != nil {
		panic(err)
	}
}
