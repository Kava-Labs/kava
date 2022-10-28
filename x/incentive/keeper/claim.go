package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/incentive/types"
)

// ClaimUSDXMintingReward pays out funds from a claim to a receiver account.
// Rewards are removed from a claim and paid out according to the multiplier, which reduces the reward amount in exchange for shorter vesting times.
func (k Keeper) ClaimUSDXMintingReward(ctx sdk.Context, owner, receiver sdk.AccAddress, multiplierName string) error {
	claim, found := k.GetClaim(ctx, types.CLAIM_TYPE_USDX_MINTING, owner)
	if !found {
		return sdkerrors.Wrapf(types.ErrClaimNotFound, "address: %s", owner)
	}

	multiplier, found := k.GetMultiplierByDenom(ctx, types.USDXMintingRewardDenom, multiplierName)
	if !found {
		return sdkerrors.Wrapf(types.ErrInvalidMultiplier, "denom '%s' has no multiplier '%s'", types.USDXMintingRewardDenom, multiplierName)
	}

	claimEnd := k.GetClaimEnd(ctx)

	if ctx.BlockTime().After(claimEnd) {
		return sdkerrors.Wrapf(types.ErrClaimExpired, "block time %s > claim end time %s", ctx.BlockTime(), claimEnd)
	}

	claim, err := k.SynchronizeUSDXMintingClaim(ctx, claim)
	if err != nil {
		return err
	}

	// Claim all instead of 1 to not break MsgClaimUSDXMintingReward
	rewardCoins := sdk.NewCoins()
	for _, coin := range claim.Reward {
		rewardAmount := coin.Amount.ToDec().Mul(multiplier.Factor).RoundInt()
		rewardCoins = rewardCoins.Add(sdk.NewCoin(coin.Denom, rewardAmount))
	}

	if rewardCoins.IsZero() {
		return types.ErrZeroClaim
	}
	length := k.GetPeriodLength(ctx.BlockTime(), multiplier.MonthsLockup)

	err = k.SendTimeLockedCoinsToAccount(ctx, types.IncentiveMacc, receiver, rewardCoins, length)
	if err != nil {
		return err
	}

	// remove claimed coins (NOT reward coins)
	claim.Reward = sdk.NewCoins()
	k.SetClaim(ctx, claim)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeClaim,
			sdk.NewAttribute(types.AttributeKeyClaimedBy, owner.String()),
			sdk.NewAttribute(types.AttributeKeyClaimAmount, claim.Reward.String()),
			sdk.NewAttribute(types.AttributeKeyClaimType, claim.Type.String()),
		),
	)
	return nil
}

// ClaimHardReward pays out funds from a claim to a receiver account.
// Rewards are removed from a claim and paid out according to the multiplier, which reduces the reward amount in exchange for shorter vesting times.
func (k Keeper) ClaimHardReward(ctx sdk.Context, owner, receiver sdk.AccAddress, denom string, multiplierName string) error {
	multiplier, found := k.GetMultiplierByDenom(ctx, denom, multiplierName)
	if !found {
		return sdkerrors.Wrapf(types.ErrInvalidMultiplier, "denom '%s' has no multiplier '%s'", denom, multiplierName)
	}

	claimEnd := k.GetClaimEnd(ctx)

	if ctx.BlockTime().After(claimEnd) {
		return sdkerrors.Wrapf(types.ErrClaimExpired, "block time %s > claim end time %s", ctx.BlockTime(), claimEnd)
	}

	k.SynchronizeHardLiquidityProviderClaim(ctx, owner)

	syncedClaim, found := k.GetHardLiquidityProviderClaim(ctx, owner)
	if !found {
		return sdkerrors.Wrapf(types.ErrClaimNotFound, "address: %s", owner)
	}

	amt := syncedClaim.Reward.AmountOf(denom)

	claimingCoins := sdk.NewCoins(sdk.NewCoin(denom, amt))
	rewardCoins := sdk.NewCoins(sdk.NewCoin(denom, amt.ToDec().Mul(multiplier.Factor).RoundInt()))
	if rewardCoins.IsZero() {
		return types.ErrZeroClaim
	}
	length := k.GetPeriodLength(ctx.BlockTime(), multiplier.MonthsLockup)

	err := k.SendTimeLockedCoinsToAccount(ctx, types.IncentiveMacc, receiver, rewardCoins, length)
	if err != nil {
		return err
	}

	// remove claimed coins (NOT reward coins)
	syncedClaim.Reward = syncedClaim.Reward.Sub(claimingCoins)
	k.SetHardLiquidityProviderClaim(ctx, syncedClaim)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeClaim,
			sdk.NewAttribute(types.AttributeKeyClaimedBy, owner.String()),
			sdk.NewAttribute(types.AttributeKeyClaimAmount, claimingCoins.String()),
			sdk.NewAttribute(types.AttributeKeyClaimType, syncedClaim.GetType()),
		),
	)
	return nil
}

// ClaimDelegatorReward pays out funds from a claim to a receiver account.
// Rewards are removed from a claim and paid out according to the multiplier, which reduces the reward amount in exchange for shorter vesting times.
func (k Keeper) ClaimDelegatorReward(ctx sdk.Context, owner, receiver sdk.AccAddress, denom string, multiplierName string) error {
	claim, found := k.GetDelegatorClaim(ctx, owner)
	if !found {
		return sdkerrors.Wrapf(types.ErrClaimNotFound, "address: %s", owner)
	}

	multiplier, found := k.GetMultiplierByDenom(ctx, denom, multiplierName)
	if !found {
		return sdkerrors.Wrapf(types.ErrInvalidMultiplier, "denom '%s' has no multiplier '%s'", denom, multiplierName)
	}

	claimEnd := k.GetClaimEnd(ctx)

	if ctx.BlockTime().After(claimEnd) {
		return sdkerrors.Wrapf(types.ErrClaimExpired, "block time %s > claim end time %s", ctx.BlockTime(), claimEnd)
	}

	syncedClaim, err := k.SynchronizeDelegatorClaim(ctx, claim)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrClaimNotFound, "address: %s", owner)
	}

	amt := syncedClaim.Reward.AmountOf(denom)

	claimingCoins := sdk.NewCoins(sdk.NewCoin(denom, amt))
	rewardCoins := sdk.NewCoins(sdk.NewCoin(denom, amt.ToDec().Mul(multiplier.Factor).RoundInt()))
	if rewardCoins.IsZero() {
		return types.ErrZeroClaim
	}

	length := k.GetPeriodLength(ctx.BlockTime(), multiplier.MonthsLockup)

	err = k.SendTimeLockedCoinsToAccount(ctx, types.IncentiveMacc, receiver, rewardCoins, length)
	if err != nil {
		return err
	}

	// remove claimed coins (NOT reward coins)
	syncedClaim.Reward = syncedClaim.Reward.Sub(claimingCoins)
	k.SetDelegatorClaim(ctx, syncedClaim)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeClaim,
			sdk.NewAttribute(types.AttributeKeyClaimedBy, owner.String()),
			sdk.NewAttribute(types.AttributeKeyClaimAmount, claimingCoins.String()),
			sdk.NewAttribute(types.AttributeKeyClaimType, syncedClaim.GetType()),
		),
	)
	return nil
}

// ClaimSwapReward pays out funds from a claim to a receiver account.
// Rewards are removed from a claim and paid out according to the multiplier, which reduces the reward amount in exchange for shorter vesting times.
func (k Keeper) ClaimSwapReward(ctx sdk.Context, owner, receiver sdk.AccAddress, denom string, multiplierName string) error {
	multiplier, found := k.GetMultiplierByDenom(ctx, denom, multiplierName)
	if !found {
		return sdkerrors.Wrapf(types.ErrInvalidMultiplier, "denom '%s' has no multiplier '%s'", denom, multiplierName)
	}

	claimEnd := k.GetClaimEnd(ctx)

	if ctx.BlockTime().After(claimEnd) {
		return sdkerrors.Wrapf(types.ErrClaimExpired, "block time %s > claim end time %s", ctx.BlockTime(), claimEnd)
	}

	syncedClaim, found := k.GetSynchronizedSwapClaim(ctx, owner)
	if !found {
		return sdkerrors.Wrapf(types.ErrClaimNotFound, "address: %s", owner)
	}

	amt := syncedClaim.Reward.AmountOf(denom)

	claimingCoins := sdk.NewCoins(sdk.NewCoin(denom, amt))
	rewardCoins := sdk.NewCoins(sdk.NewCoin(denom, amt.ToDec().Mul(multiplier.Factor).RoundInt()))
	if rewardCoins.IsZero() {
		return types.ErrZeroClaim
	}
	length := k.GetPeriodLength(ctx.BlockTime(), multiplier.MonthsLockup)

	err := k.SendTimeLockedCoinsToAccount(ctx, types.IncentiveMacc, receiver, rewardCoins, length)
	if err != nil {
		return err
	}

	// remove claimed coins (NOT reward coins)
	syncedClaim.Reward = syncedClaim.Reward.Sub(claimingCoins)
	k.SetSwapClaim(ctx, syncedClaim)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeClaim,
			sdk.NewAttribute(types.AttributeKeyClaimedBy, owner.String()),
			sdk.NewAttribute(types.AttributeKeyClaimAmount, claimingCoins.String()),
			sdk.NewAttribute(types.AttributeKeyClaimType, syncedClaim.GetType()),
		),
	)
	return nil
}

// ClaimSavingsReward is a stub method for MsgServer interface compliance
func (k Keeper) ClaimSavingsReward(ctx sdk.Context, owner, receiver sdk.AccAddress, denom string, multiplierName string) error {
	multiplier, found := k.GetMultiplierByDenom(ctx, denom, multiplierName)
	if !found {
		return sdkerrors.Wrapf(types.ErrInvalidMultiplier, "denom '%s' has no multiplier '%s'", denom, multiplierName)
	}

	claimEnd := k.GetClaimEnd(ctx)

	if ctx.BlockTime().After(claimEnd) {
		return sdkerrors.Wrapf(types.ErrClaimExpired, "block time %s > claim end time %s", ctx.BlockTime(), claimEnd)
	}

	k.SynchronizeSavingsClaim(ctx, owner)

	syncedClaim, found := k.GetSavingsClaim(ctx, owner)
	if !found {
		return sdkerrors.Wrapf(types.ErrClaimNotFound, "address: %s", owner)
	}

	amt := syncedClaim.Reward.AmountOf(denom)

	claimingCoins := sdk.NewCoins(sdk.NewCoin(denom, amt))
	rewardCoins := sdk.NewCoins(sdk.NewCoin(denom, amt.ToDec().Mul(multiplier.Factor).RoundInt()))
	if rewardCoins.IsZero() {
		return types.ErrZeroClaim
	}
	length := k.GetPeriodLength(ctx.BlockTime(), multiplier.MonthsLockup)

	err := k.SendTimeLockedCoinsToAccount(ctx, types.IncentiveMacc, receiver, rewardCoins, length)
	if err != nil {
		return err
	}

	// remove claimed coins (NOT reward coins)
	syncedClaim.Reward = syncedClaim.Reward.Sub(claimingCoins)
	k.SetSavingsClaim(ctx, syncedClaim)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeClaim,
			sdk.NewAttribute(types.AttributeKeyClaimedBy, owner.String()),
			sdk.NewAttribute(types.AttributeKeyClaimAmount, claimingCoins.String()),
			sdk.NewAttribute(types.AttributeKeyClaimType, syncedClaim.GetType()),
		),
	)
	return nil
}

// ClaimEarnReward pays out funds from a claim to a receiver account.
// Rewards are removed from a claim and paid out according to the multiplier, which reduces the reward amount in exchange for shorter vesting times.
func (k Keeper) ClaimEarnReward(ctx sdk.Context, owner, receiver sdk.AccAddress, denom string, multiplierName string) error {
	multiplier, found := k.GetMultiplierByDenom(ctx, denom, multiplierName)
	if !found {
		return sdkerrors.Wrapf(types.ErrInvalidMultiplier, "denom '%s' has no multiplier '%s'", denom, multiplierName)
	}

	claimEnd := k.GetClaimEnd(ctx)

	if ctx.BlockTime().After(claimEnd) {
		return sdkerrors.Wrapf(types.ErrClaimExpired, "block time %s > claim end time %s", ctx.BlockTime(), claimEnd)
	}

	syncedClaim, found := k.GetSynchronizedEarnClaim(ctx, owner)
	if !found {
		return sdkerrors.Wrapf(types.ErrClaimNotFound, "address: %s", owner)
	}

	amt := syncedClaim.Reward.AmountOf(denom)

	claimingCoins := sdk.NewCoins(sdk.NewCoin(denom, amt))
	rewardCoins := sdk.NewCoins(sdk.NewCoin(denom, amt.ToDec().Mul(multiplier.Factor).RoundInt()))
	if rewardCoins.IsZero() {
		return types.ErrZeroClaim
	}
	length := k.GetPeriodLength(ctx.BlockTime(), multiplier.MonthsLockup)

	err := k.SendTimeLockedCoinsToAccount(ctx, types.IncentiveMacc, receiver, rewardCoins, length)
	if err != nil {
		return err
	}

	// remove claimed coins (NOT reward coins)
	syncedClaim.Reward = syncedClaim.Reward.Sub(claimingCoins)
	k.SetEarnClaim(ctx, syncedClaim)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeClaim,
			sdk.NewAttribute(types.AttributeKeyClaimedBy, owner.String()),
			sdk.NewAttribute(types.AttributeKeyClaimAmount, claimingCoins.String()),
			sdk.NewAttribute(types.AttributeKeyClaimType, syncedClaim.GetType()),
		),
	)
	return nil
}
