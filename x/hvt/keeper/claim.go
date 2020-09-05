package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/hvt/types"
)

// ClaimReward sends the reward amount to the reward owner and deletes the claim from the store
func (k Keeper) ClaimReward(ctx sdk.Context, claimHolder sdk.AccAddress, depositDenom string, depositType types.DepositType, multiplier types.RewardMultiplier) error {

	claim, found := k.GetClaim(ctx, claimHolder, depositDenom, depositType)
	if !found {
		return sdkerrors.Wrapf(types.ErrClaimNotFound, "no %s %s claim found for %s", depositDenom, depositType, claimHolder)
	}

	var err error
	switch depositType {
	case types.LP:
		err = k.claimLPReward(ctx, claim, multiplier)
	case types.Gov:
		err = k.claimGovReward(ctx, claim, multiplier)
	}
	if err != nil {
		return err
	}
	k.DeleteClaim(ctx, claim)
	return nil
}

func (k Keeper) claimLPReward(ctx sdk.Context, claim types.Claim, rewardMultiplier types.RewardMultiplier) error {
	lps, found := k.GetLPSchedule(ctx, claim.DepositDenom)
	if !found {
		return sdkerrors.Wrapf(types.ErrLPScheduleNotFound, claim.DepositDenom)
	}
	multiplier, found := k.GetMultiplier(lps, rewardMultiplier)
	if !found {
		return sdkerrors.Wrapf(types.ErrInvalidMultiplier, string(rewardMultiplier))
	}
	rewardAmount := sdk.NewDecFromInt(claim.Amount.Amount).Mul(multiplier.Factor).RoundInt()
	if rewardAmount.IsZero() {
		return types.ErrZeroClaim
	}
	rewardCoin := sdk.NewCoin(claim.Amount.Denom, rewardAmount)

	return k.SendTimeLockedCoinsToAccount(ctx, types.LPAccount, claim.Owner, sdk.NewCoins(rewardCoin), int64(multiplier.LockDuration.Seconds()))
}

func (k Keeper) claimGovReward(ctx sdk.Context, claim types.Claim, rewardMultiplier types.RewardMultiplier) error {
	gds, found := k.GetGovSchedule(ctx, claim.DepositDenom)
	if !found {
		return sdkerrors.Wrapf(types.ErrLPScheduleNotFound, claim.DepositDenom)
	}
	multiplier, found := k.GetMultiplier(gds, rewardMultiplier)
	if !found {
		return sdkerrors.Wrapf(types.ErrInvalidMultiplier, string(rewardMultiplier))
	}
	rewardAmount := sdk.NewDecFromInt(claim.Amount.Amount).Mul(multiplier.Factor).RoundInt()
	if rewardAmount.IsZero() {
		return types.ErrZeroClaim
	}
	rewardCoin := sdk.NewCoin(claim.Amount.Denom, rewardAmount)

	return k.SendTimeLockedCoinsToAccount(ctx, types.GovAccount, claim.Owner, sdk.NewCoins(rewardCoin), int64(multiplier.LockDuration.Seconds()))
}
