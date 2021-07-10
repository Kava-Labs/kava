package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/incentive/types"
	validatorvesting "github.com/kava-labs/kava/x/validator-vesting"
)

// ClaimUSDXMintingReward sends the reward amount to the input address and zero's out the claim in the store
func (k Keeper) ClaimUSDXMintingReward(ctx sdk.Context, addr sdk.AccAddress, multiplierName types.MultiplierName) error {
	claim, found := k.GetUSDXMintingClaim(ctx, addr)
	if !found {
		return sdkerrors.Wrapf(types.ErrClaimNotFound, "address: %s", addr)
	}

	multiplier, found := k.GetMultiplier(ctx, multiplierName)
	if !found {
		return sdkerrors.Wrapf(types.ErrInvalidMultiplier, string(multiplierName))
	}

	claimEnd := k.GetClaimEnd(ctx)

	if ctx.BlockTime().After(claimEnd) {
		return sdkerrors.Wrapf(types.ErrClaimExpired, "block time %s > claim end time %s", ctx.BlockTime(), claimEnd)
	}

	claim, err := k.SynchronizeUSDXMintingClaim(ctx, claim)
	if err != nil {
		return err
	}

	rewardAmount := claim.Reward.Amount.ToDec().Mul(multiplier.Factor).RoundInt()
	if rewardAmount.IsZero() {
		return types.ErrZeroClaim
	}
	rewardCoin := sdk.NewCoin(claim.Reward.Denom, rewardAmount)
	length, err := k.GetPeriodLength(ctx, multiplier)
	if err != nil {
		return err
	}

	err = k.SendTimeLockedCoinsToAccount(ctx, types.IncentiveMacc, addr, sdk.NewCoins(rewardCoin), length)
	if err != nil {
		return err
	}

	k.ZeroUSDXMintingClaim(ctx, claim)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeClaim,
			sdk.NewAttribute(types.AttributeKeyClaimedBy, claim.GetOwner().String()),
			sdk.NewAttribute(types.AttributeKeyClaimAmount, claim.GetReward().String()),
			sdk.NewAttribute(types.AttributeKeyClaimAmount, claim.GetType()),
		),
	)
	return nil
}

// ClaimUSDXMintingReward sends the reward amount to the input receiver address and zero's out the claim in the store
func (k Keeper) ClaimUSDXMintingRewardVVesting(ctx sdk.Context, owner, receiver sdk.AccAddress, multiplierName types.MultiplierName) error {
	claim, found := k.GetUSDXMintingClaim(ctx, owner)
	if !found {
		return sdkerrors.Wrapf(types.ErrClaimNotFound, "address: %s", owner)
	}

	acc := k.accountKeeper.GetAccount(ctx, owner)
	if acc == nil {
		return sdkerrors.Wrapf(types.ErrAccountNotFound, "address not found: %s", owner)
	}
	_, ok := acc.(*validatorvesting.ValidatorVestingAccount)
	if !ok {
		return sdkerrors.Wrapf(types.ErrInvalidAccountType, "owner account must be validator vesting account %s", owner)
	}

	multiplier, found := k.GetMultiplier(ctx, multiplierName)
	if !found {
		return sdkerrors.Wrapf(types.ErrInvalidMultiplier, string(multiplierName))
	}

	claimEnd := k.GetClaimEnd(ctx)

	if ctx.BlockTime().After(claimEnd) {
		return sdkerrors.Wrapf(types.ErrClaimExpired, "block time %s > claim end time %s", ctx.BlockTime(), claimEnd)
	}

	claim, err := k.SynchronizeUSDXMintingClaim(ctx, claim)
	if err != nil {
		return err
	}

	rewardAmount := claim.Reward.Amount.ToDec().Mul(multiplier.Factor).RoundInt()
	if rewardAmount.IsZero() {
		return types.ErrZeroClaim
	}
	rewardCoin := sdk.NewCoin(claim.Reward.Denom, rewardAmount)
	length, err := k.GetPeriodLength(ctx, multiplier)
	if err != nil {
		return err
	}

	err = k.SendTimeLockedCoinsToAccount(ctx, types.IncentiveMacc, receiver, sdk.NewCoins(rewardCoin), length)
	if err != nil {
		return err
	}

	k.ZeroUSDXMintingClaim(ctx, claim)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeClaim,
			sdk.NewAttribute(types.AttributeKeyClaimedBy, claim.GetOwner().String()),
			sdk.NewAttribute(types.AttributeKeyClaimAmount, claim.GetReward().String()),
			sdk.NewAttribute(types.AttributeKeyClaimAmount, claim.GetType()),
		),
	)
	return nil
}

// ClaimHardReward sends the reward amount to the input address and zero's out the claim in the store
func (k Keeper) ClaimHardReward(ctx sdk.Context, addr sdk.AccAddress, multiplierName types.MultiplierName) error {
	_, found := k.GetHardLiquidityProviderClaim(ctx, addr)
	if !found {
		return sdkerrors.Wrapf(types.ErrClaimNotFound, "address: %s", addr)
	}

	multiplier, found := k.GetMultiplier(ctx, multiplierName)
	if !found {
		return sdkerrors.Wrapf(types.ErrInvalidMultiplier, string(multiplierName))
	}

	claimEnd := k.GetClaimEnd(ctx)

	if ctx.BlockTime().After(claimEnd) {
		return sdkerrors.Wrapf(types.ErrClaimExpired, "block time %s > claim end time %s", ctx.BlockTime(), claimEnd)
	}

	k.SynchronizeHardLiquidityProviderClaim(ctx, addr)

	claim, found := k.GetHardLiquidityProviderClaim(ctx, addr)
	if !found {
		return sdkerrors.Wrapf(types.ErrClaimNotFound, "address: %s", addr)
	}

	var rewardCoins sdk.Coins
	for _, coin := range claim.Reward {
		rewardAmount := coin.Amount.ToDec().Mul(multiplier.Factor).RoundInt()
		if rewardAmount.IsZero() {
			continue
		}
		rewardCoins = append(rewardCoins, sdk.NewCoin(coin.Denom, rewardAmount))
	}
	if rewardCoins.IsZero() {
		return types.ErrZeroClaim
	}
	length, err := k.GetPeriodLength(ctx, multiplier)
	if err != nil {
		return err
	}

	err = k.SendTimeLockedCoinsToAccount(ctx, types.IncentiveMacc, addr, rewardCoins, length)
	if err != nil {
		return err
	}

	k.ZeroHardLiquidityProviderClaim(ctx, claim)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeClaim,
			sdk.NewAttribute(types.AttributeKeyClaimedBy, claim.GetOwner().String()),
			sdk.NewAttribute(types.AttributeKeyClaimAmount, claim.GetReward().String()),
			sdk.NewAttribute(types.AttributeKeyClaimType, claim.GetType()),
		),
	)
	return nil
}

// ClaimHardRewardVVesting sends the reward amount to the input address and zero's out the claim in the store
func (k Keeper) ClaimHardRewardVVesting(ctx sdk.Context, owner, receiver sdk.AccAddress, multiplierName types.MultiplierName) error {
	_, found := k.GetHardLiquidityProviderClaim(ctx, owner)
	if !found {
		return sdkerrors.Wrapf(types.ErrClaimNotFound, "address: %s", owner)
	}

	acc := k.accountKeeper.GetAccount(ctx, owner)
	if acc == nil {
		return sdkerrors.Wrapf(types.ErrAccountNotFound, "address not found: %s", owner)
	}
	_, ok := acc.(*validatorvesting.ValidatorVestingAccount)
	if !ok {
		return sdkerrors.Wrapf(types.ErrInvalidAccountType, "owner account must be validator vesting account %s", owner)
	}

	multiplier, found := k.GetMultiplier(ctx, multiplierName)
	if !found {
		return sdkerrors.Wrapf(types.ErrInvalidMultiplier, string(multiplierName))
	}

	claimEnd := k.GetClaimEnd(ctx)

	if ctx.BlockTime().After(claimEnd) {
		return sdkerrors.Wrapf(types.ErrClaimExpired, "block time %s > claim end time %s", ctx.BlockTime(), claimEnd)
	}

	k.SynchronizeHardLiquidityProviderClaim(ctx, owner)

	claim, found := k.GetHardLiquidityProviderClaim(ctx, owner)
	if !found {
		return sdkerrors.Wrapf(types.ErrClaimNotFound, "address: %s", owner)
	}

	var rewardCoins sdk.Coins
	for _, coin := range claim.Reward {
		rewardAmount := coin.Amount.ToDec().Mul(multiplier.Factor).RoundInt()
		if rewardAmount.IsZero() {
			continue
		}
		rewardCoins = append(rewardCoins, sdk.NewCoin(coin.Denom, rewardAmount))
	}
	if rewardCoins.IsZero() {
		return types.ErrZeroClaim
	}
	length, err := k.GetPeriodLength(ctx, multiplier)
	if err != nil {
		return err
	}

	err = k.SendTimeLockedCoinsToAccount(ctx, types.IncentiveMacc, receiver, rewardCoins, length)
	if err != nil {
		return err
	}

	k.ZeroHardLiquidityProviderClaim(ctx, claim)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeClaim,
			sdk.NewAttribute(types.AttributeKeyClaimedBy, claim.GetOwner().String()),
			sdk.NewAttribute(types.AttributeKeyClaimAmount, claim.GetReward().String()),
			sdk.NewAttribute(types.AttributeKeyClaimType, claim.GetType()),
		),
	)
	return nil
}

// ClaimDelegatorReward sends the reward amount to the input address and zero's out the delegator claim in the store
func (k Keeper) ClaimDelegatorReward(ctx sdk.Context, addr sdk.AccAddress, multiplierName types.MultiplierName) error {
	claim, found := k.GetDelegatorClaim(ctx, addr)
	if !found {
		return sdkerrors.Wrapf(types.ErrClaimNotFound, "address: %s", addr)
	}

	multiplier, found := k.GetMultiplier(ctx, multiplierName)
	if !found {
		return sdkerrors.Wrapf(types.ErrInvalidMultiplier, string(multiplierName))
	}

	claimEnd := k.GetClaimEnd(ctx)

	if ctx.BlockTime().After(claimEnd) {
		return sdkerrors.Wrapf(types.ErrClaimExpired, "block time %s > claim end time %s", ctx.BlockTime(), claimEnd)
	}

	syncedClaim, err := k.SynchronizeDelegatorClaim(ctx, claim)
	if !found {
		return sdkerrors.Wrapf(types.ErrClaimNotFound, "address: %s", addr)
	}

	var rewardCoins sdk.Coins
	for _, coin := range syncedClaim.Reward {
		rewardAmount := coin.Amount.ToDec().Mul(multiplier.Factor).RoundInt()
		if rewardAmount.IsZero() {
			continue
		}
		rewardCoins = append(rewardCoins, sdk.NewCoin(coin.Denom, rewardAmount))
	}
	if rewardCoins.IsZero() {
		return types.ErrZeroClaim
	}
	length, err := k.GetPeriodLength(ctx, multiplier)
	if err != nil {
		return err
	}

	err = k.SendTimeLockedCoinsToAccount(ctx, types.IncentiveMacc, addr, rewardCoins, length)
	if err != nil {
		return err
	}

	k.ZeroDelegatorClaim(ctx, syncedClaim)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeClaim,
			sdk.NewAttribute(types.AttributeKeyClaimedBy, syncedClaim.GetOwner().String()),
			sdk.NewAttribute(types.AttributeKeyClaimAmount, syncedClaim.GetReward().String()),
			sdk.NewAttribute(types.AttributeKeyClaimType, syncedClaim.GetType()),
		),
	)
	return nil
}

// ClaimDelegatorRewardVVesting sends the reward amount to the input address and zero's out the claim in the store
func (k Keeper) ClaimDelegatorRewardVVesting(ctx sdk.Context, owner, receiver sdk.AccAddress, multiplierName types.MultiplierName) error {
	claim, found := k.GetDelegatorClaim(ctx, owner)
	if !found {
		return sdkerrors.Wrapf(types.ErrClaimNotFound, "address: %s", owner)
	}

	acc := k.accountKeeper.GetAccount(ctx, owner)
	if acc == nil {
		return sdkerrors.Wrapf(types.ErrAccountNotFound, "address not found: %s", owner)
	}
	_, ok := acc.(*validatorvesting.ValidatorVestingAccount)
	if !ok {
		return sdkerrors.Wrapf(types.ErrInvalidAccountType, "owner account must be validator vesting account %s", owner)
	}

	multiplier, found := k.GetMultiplier(ctx, multiplierName)
	if !found {
		return sdkerrors.Wrapf(types.ErrInvalidMultiplier, string(multiplierName))
	}

	claimEnd := k.GetClaimEnd(ctx)

	if ctx.BlockTime().After(claimEnd) {
		return sdkerrors.Wrapf(types.ErrClaimExpired, "block time %s > claim end time %s", ctx.BlockTime(), claimEnd)
	}

	syncedClaim, err := k.SynchronizeDelegatorClaim(ctx, claim)
	if !found {
		return sdkerrors.Wrapf(types.ErrClaimNotFound, "address: %s", owner)
	}

	var rewardCoins sdk.Coins
	for _, coin := range syncedClaim.Reward {
		rewardAmount := coin.Amount.ToDec().Mul(multiplier.Factor).RoundInt()
		if rewardAmount.IsZero() {
			continue
		}
		rewardCoins = append(rewardCoins, sdk.NewCoin(coin.Denom, rewardAmount))
	}
	if rewardCoins.IsZero() {
		return types.ErrZeroClaim
	}
	length, err := k.GetPeriodLength(ctx, multiplier)
	if err != nil {
		return err
	}

	err = k.SendTimeLockedCoinsToAccount(ctx, types.IncentiveMacc, receiver, rewardCoins, length)
	if err != nil {
		return err
	}

	k.ZeroDelegatorClaim(ctx, syncedClaim)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeClaim,
			sdk.NewAttribute(types.AttributeKeyClaimedBy, syncedClaim.GetOwner().String()),
			sdk.NewAttribute(types.AttributeKeyClaimAmount, syncedClaim.GetReward().String()),
			sdk.NewAttribute(types.AttributeKeyClaimType, syncedClaim.GetType()),
		),
	)
	return nil
}

