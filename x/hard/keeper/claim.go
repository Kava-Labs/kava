package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/hard/types"
	validatorvesting "github.com/kava-labs/kava/x/validator-vesting"
)

const (
	// BeginningOfMonth hard rewards that are claimed after the 15th at 14:00UTC of the month always vest on the first of the month
	BeginningOfMonth = 1
	// MidMonth hard rewards that are claimed before the 15th at 14:00UTC of the month always vest on the 15 of the month
	MidMonth = 15
	// PaymentHour hard rewards always vest at 14:00UTC
	PaymentHour = 14
)

// ClaimReward sends the reward amount to the reward owner and deletes the claim from the store
func (k Keeper) ClaimReward(ctx sdk.Context, claimHolder sdk.AccAddress, receiver sdk.AccAddress, depositDenom string, claimType types.ClaimType, multiplier types.MultiplierName) error {

	claim, found := k.GetClaim(ctx, claimHolder, depositDenom, claimType)
	if !found {
		return sdkerrors.Wrapf(types.ErrClaimNotFound, "no %s %s claim found for %s", depositDenom, claimType, claimHolder)
	}

	err := k.validateSenderReceiver(ctx, claimHolder, receiver)
	if err != nil {
		return err
	}

	switch claimType {
	case types.LP:
		err = k.claimLPReward(ctx, claim, receiver, multiplier)
	case types.Stake:
		err = k.claimDelegatorReward(ctx, claim, receiver, multiplier)
	default:
		return sdkerrors.Wrap(types.ErrInvalidClaimType, string(claimType))
	}
	if err != nil {
		return err
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeClaimHardReward,
			sdk.NewAttribute(sdk.AttributeKeyAmount, claim.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyClaimHolder, claimHolder.String()),
			sdk.NewAttribute(types.AttributeKeyDepositDenom, depositDenom),
			sdk.NewAttribute(types.AttributeKeyClaimType, string(claimType)),
			sdk.NewAttribute(types.AttributeKeyClaimMultiplier, string(multiplier)),
		),
	)
	k.DeleteClaim(ctx, claim)
	return nil
}

// GetPeriodLength returns the length of the period based on the input blocktime and multiplier
// note that pay dates are always the 1st or 15th of the month at 14:00UTC.
func (k Keeper) GetPeriodLength(ctx sdk.Context, multiplier types.Multiplier) (int64, error) {

	if multiplier.MonthsLockup == 0 {
		return 0, nil
	}
	switch multiplier.Name {
	case types.Small, types.Medium, types.Large:
		currentDay := ctx.BlockTime().Day()
		payDay := BeginningOfMonth
		monthOffset := int64(1)
		if currentDay < MidMonth || (currentDay == MidMonth && ctx.BlockTime().Hour() < PaymentHour) {
			payDay = MidMonth
			monthOffset = int64(0)
		}
		periodEndDate := time.Date(ctx.BlockTime().Year(), ctx.BlockTime().Month(), payDay, PaymentHour, 0, 0, 0, time.UTC).AddDate(0, int(multiplier.MonthsLockup+monthOffset), 0)
		return periodEndDate.Unix() - ctx.BlockTime().Unix(), nil
	}
	return 0, types.ErrInvalidMultiplier
}

func (k Keeper) claimLPReward(ctx sdk.Context, claim types.Claim, receiver sdk.AccAddress, multiplierName types.MultiplierName) error {
	lps, found := k.GetLPSchedule(ctx, claim.DepositDenom)
	if !found {
		return sdkerrors.Wrapf(types.ErrLPScheduleNotFound, claim.DepositDenom)
	}
	multiplier, found := lps.GetMultiplier(multiplierName)
	if !found {
		return sdkerrors.Wrapf(types.ErrInvalidMultiplier, string(multiplierName))
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

	return k.SendTimeLockedCoinsToAccount(ctx, types.LPAccount, receiver, sdk.NewCoins(rewardCoin), length)
}

func (k Keeper) claimDelegatorReward(ctx sdk.Context, claim types.Claim, receiver sdk.AccAddress, multiplierName types.MultiplierName) error {
	dss, found := k.GetDelegatorSchedule(ctx, claim.DepositDenom)
	if !found {
		return sdkerrors.Wrapf(types.ErrLPScheduleNotFound, claim.DepositDenom)
	}
	multiplier, found := dss.DistributionSchedule.GetMultiplier(multiplierName)
	if !found {
		return sdkerrors.Wrapf(types.ErrInvalidMultiplier, string(multiplierName))
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

	return k.SendTimeLockedCoinsToAccount(ctx, types.DelegatorAccount, receiver, sdk.NewCoins(rewardCoin), length)
}

func (k Keeper) validateSenderReceiver(ctx sdk.Context, sender, receiver sdk.AccAddress) error {
	senderAcc := k.accountKeeper.GetAccount(ctx, sender)
	if senderAcc == nil {
		return sdkerrors.Wrapf(types.ErrAccountNotFound, sender.String())
	}
	switch senderAcc.(type) {
	case *validatorvesting.ValidatorVestingAccount:
		if sender.Equals(receiver) {
			return sdkerrors.Wrapf(types.ErrInvalidAccountType, "%T", senderAcc)
		}
	default:
		if !sender.Equals(receiver) {
			return sdkerrors.Wrapf(types.ErrInvalidReceiver, "%s", sender)
		}
	}
	return nil
}
