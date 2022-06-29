package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/kava-labs/kava/x/liquidstaking/types"
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

	delegationHolder, found := k.GetDelegationHolder(ctx, valAddr)
	if !found {
		delegationHolder = types.NewDelegationHolder(valAddr)
	}

	err := k.bankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.Coins{liquidCoin})
	if err != nil {
		return err
	}

	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, delegatorAddr, sdk.Coins{liquidCoin})
	if err != nil {
		return err
	}

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

	k.SetDelegationHolder(ctx, delegationHolder)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeMintDerivative,
			sdk.NewAttribute(sdk.AttributeKeyAmount, liquidCoin.String()),
			sdk.NewAttribute(types.AttributeKeyValidator, delegationHolder.Validator.String()),
			sdk.NewAttribute(types.AttributeKeyModuleAccount, moduleAccAddress.String()),
		),
	)

	return nil
}

// BurnDerivative burns an existing derivative
func (k Keeper) BurnDerivative(ctx sdk.Context, delegatorAddr sdk.AccAddress, valAddr sdk.ValAddress, amount sdk.Coin) error {

	// User must have enough tokens to fulfill redeem request
	balance := k.bankKeeper.GetBalance(ctx, delegatorAddr, amount.Denom)
	if balance.Amount.LT(amount.Amount) {
		return types.ErrNotEnoughBalance
	}

	validator, found := k.stakingKeeper.GetValidator(ctx, valAddr)
	if !found {
		return types.ErrNoValidatorFound
	}

	delegationHolder, found := k.GetDelegationHolder(ctx, valAddr)
	if !found {
		return types.ErrNoDerivativeFound
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
	shareDenomSupply := k.bankKeeper.GetSupply(ctx, amount.Denom)
	shares := delegation.Shares.Mul(amount.Amount.ToDec()).QuoInt(shareDenomSupply.Amount)

	returnAmount, err := k.stakingKeeper.Unbond(ctx, maccAddr, valAddr, shares)
	if err != nil {
		return err
	}

	if validator.IsBonded() {
		k.bondedTokensToNotBonded(ctx, returnAmount)
	}

	// Burn share amount from user's address
	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, delegatorAddr, types.ModuleAccountName, sdk.NewCoins(amount))
	if err != nil {
		return err
	}
	err = k.bankKeeper.BurnCoins(ctx, types.ModuleAccountName, sdk.NewCoins(amount))
	if err != nil {
		return err
	}

	// Create a delegation for an equivalent amount of KAVA tokens from the user
	returnCoin := sdk.NewCoin(k.stakingKeeper.BondDenom(ctx), returnAmount)
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleAccountName, delegatorAddr, sdk.NewCoins(returnCoin))
	if err != nil {
		return err
	}
	_, err = k.stakingKeeper.Delegate(ctx, delegatorAddr, returnAmount, stakingtypes.Unbonded, validator, true)
	if err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeBurnDerivative,
			sdk.NewAttribute(types.AttributeKeyAmountBurned, amount.String()),
			sdk.NewAttribute(types.AttributeKeyAmountReturned, returnCoin.String()),
			sdk.NewAttribute(types.AttributeKeyValidator, delegationHolder.Validator.String()),
			sdk.NewAttribute(types.AttributeKeyModuleAccount, maccAddr.String()),
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
