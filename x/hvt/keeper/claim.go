package keeper

import (
	"time"

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
	case types.Stake:
		err = k.claimDelegatorReward(ctx, claim, multiplier)
	default:
		return sdkerrors.Wrap(types.ErrInvalidDepositType, string(depositType))
	}
	if err != nil {
		return err
	}
	k.DeleteClaim(ctx, claim)
	return nil
}

// GetPeriodLength returns the length of the period based on the input blocktime and multiplier
// note that pay dates are always the 1st or 15th of the month at 14:00UTC.
func (k Keeper) GetPeriodLength(ctx sdk.Context, multiplier types.Multiplier) (int64, error) {

	switch multiplier.Name {
	case types.Small:
		return 0, nil
	case types.Medium, types.Large:
		currentDay := ctx.BlockTime().Day()
		payDay := 1
		monthOffset := 1
		if currentDay < 15 || (currentDay == 15 && ctx.BlockTime().Hour() < 14) {
			payDay = 15
			monthOffset = 0
		}
		periodEndDate := time.Date(ctx.BlockTime().Year(), ctx.BlockTime().Month(), payDay, 14, 0, 0, 0, time.UTC).AddDate(0, multiplier.MonthsLockup+monthOffset, 0)
		return periodEndDate.Unix() - ctx.BlockTime().Unix(), nil
	}
	return 0, types.ErrInvalidMultiplier
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
	if ctx.BlockTime().After(lps.ClaimEnd) {
		return sdkerrors.Wrapf(types.ErrClaimExpired, "block time %s > claim end time %s", ctx.BlockTime(), lps.ClaimEnd)
	}
	rewardAmount := sdk.NewDecFromInt(claim.Amount.Amount).Mul(multiplier.Factor).RoundInt()
	if rewardAmount.IsZero() {
		return types.ErrZeroClaim
	}
	rewardCoin := sdk.NewCoin(claim.Amount.Denom, rewardAmount)
	length, err := k.GetPeriodLength(ctx, multiplier)
	if err != nil {
		return err
	}

	return k.SendTimeLockedCoinsToAccount(ctx, types.LPAccount, claim.Owner, sdk.NewCoins(rewardCoin), length)
}

func (k Keeper) claimDelegatorReward(ctx sdk.Context, claim types.Claim, rewardMultiplier types.RewardMultiplier) error {
	dss, found := k.GetDelegatorSchedule(ctx, claim.DepositDenom)
	if !found {
		return sdkerrors.Wrapf(types.ErrLPScheduleNotFound, claim.DepositDenom)
	}
	multiplier, found := k.GetMultiplier(dss.DistributionSchedule, rewardMultiplier)
	if !found {
		return sdkerrors.Wrapf(types.ErrInvalidMultiplier, string(rewardMultiplier))
	}
	if ctx.BlockTime().After(dss.DistributionSchedule.ClaimEnd) {
		return sdkerrors.Wrapf(types.ErrClaimExpired, "block time %s > claim end time %s", ctx.BlockTime(), dss.DistributionSchedule.ClaimEnd)
	}
	rewardAmount := sdk.NewDecFromInt(claim.Amount.Amount).Mul(multiplier.Factor).RoundInt()
	if rewardAmount.IsZero() {
		return types.ErrZeroClaim
	}
	rewardCoin := sdk.NewCoin(claim.Amount.Denom, rewardAmount)

	length, err := k.GetPeriodLength(ctx, multiplier)
	if err != nil {
		return err
	}

	return k.SendTimeLockedCoinsToAccount(ctx, types.DelegatorAccount, claim.Owner, sdk.NewCoins(rewardCoin), length)
}
