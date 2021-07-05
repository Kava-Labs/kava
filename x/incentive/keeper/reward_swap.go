package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/incentive/types"
)

// AccumulateSwapRewards calculates new rewards to distribute this block and updates the global index to reflect this.
func (k Keeper) AccumulateSwapRewards(ctx sdk.Context, rewardPeriod types.MultiRewardPeriod) {

	previousAccrualTime, found := k.GetSwapRewardAccrualTime(ctx, rewardPeriod.CollateralType)
	if !found {
		previousAccrualTime = ctx.BlockTime()
	}

	indexes, found := k.GetSwapRewardIndexes(ctx, rewardPeriod.CollateralType)
	if !found {
		indexes = types.RewardIndexes{}
	}

	acc := types.NewAccumulator(previousAccrualTime, indexes)

	totalShares, found := k.swapKeeper.GetPoolShares(ctx, rewardPeriod.CollateralType)
	if !found {
		totalShares = sdk.ZeroDec()
	}

	acc.Accumulate(rewardPeriod, totalShares, ctx.BlockTime())

	k.SetSwapRewardAccrualTime(ctx, rewardPeriod.CollateralType, acc.PreviousAccumulationTime)
	k.SetSwapRewardIndexes(ctx, rewardPeriod.CollateralType, acc.Indexes)
}
