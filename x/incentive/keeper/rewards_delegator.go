package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/incentive/types"
)

// AccumulateDelegatorRewards calculates new rewards to distribute this block and updates the global indexes to reflect this.
// The provided rewardPeriod must be valid to avoid panics in calculating time durations.
func (k Keeper) AccumulateDelegatorRewards(ctx sdk.Context, rewardPeriod types.MultiRewardPeriod) {
	previousAccrualTime, found := k.GetPreviousDelegatorRewardAccrualTime(ctx, rewardPeriod.CollateralType)
	if !found {
		previousAccrualTime = ctx.BlockTime()
	}

	rewards, upTo := types.CalculatePerSecondRewards(
		rewardPeriod.Start,
		rewardPeriod.End,
		sdk.NewDecCoinsFromCoins(rewardPeriod.RewardsPerSecond...),
		previousAccrualTime,
		ctx.BlockTime(),
	)

	k.distributors[types.RewardTypeSwap].Distribute(ctx, rewardPeriod.CollateralType, rewards)

	k.SetPreviousDelegatorRewardAccrualTime(ctx, rewardPeriod.CollateralType, upTo)
}
