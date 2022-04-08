package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/incentive/types"
	savingstypes "github.com/kava-labs/kava/x/savings/types"
)

// AccumulateSavingsRewards calculates new rewards to distribute this block and updates the global indexes
func (k Keeper) AccumulateSavingsRewards(ctx sdk.Context, rewardPeriod types.MultiRewardPeriod) {

	previousAccrualTime, found := k.GetSavingsRewardAccrualTime(ctx, rewardPeriod.CollateralType)
	if !found {
		previousAccrualTime = ctx.BlockTime()
	}

	indexes, found := k.GetSavingsRewardIndexes(ctx, rewardPeriod.CollateralType)
	if !found {
		indexes = types.RewardIndexes{}
	}

	acc := types.NewAccumulator(previousAccrualTime, indexes)

	savingsMacc := k.accountKeeper.GetModuleAccount(ctx, savingstypes.ModuleName)
	maccCoins := k.bankKeeper.GetAllBalances(ctx, savingsMacc.GetAddress())
	denomBalance := maccCoins.AmountOf(rewardPeriod.CollateralType)

	acc.Accumulate(rewardPeriod, denomBalance.ToDec(), ctx.BlockTime())

	k.SetSavingsRewardAccrualTime(ctx, rewardPeriod.CollateralType, acc.PreviousAccumulationTime)

	if len(acc.Indexes) > 0 {
		// the store panics when setting empty or nil indexes
		k.SetSavingsRewardIndexes(ctx, rewardPeriod.CollateralType, acc.Indexes)
	}
}
