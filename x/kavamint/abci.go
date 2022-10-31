package kavamint

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/kavamint/keeper"
)

// BeginBlocker mints & distributes new tokens for the previous block.
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	previousBlockTime, found := k.GetPreviousBlockTime(ctx)
	if !found {
		previousBlockTime = ctx.BlockTime()
	}

	// ------------- Staking Rewards -------------
	stakingRewardCoins, err := k.AccumulateStakingRewards(ctx, previousBlockTime)
	if err != nil {
		panic(err)
	}

	// mint staking rewards
	if err := k.MintCoins(ctx, stakingRewardCoins); err != nil {
		panic(err)
	}

	// send staking rewards to auth fee collector for distribution to validators
	if err := k.AddCollectedFees(ctx, stakingRewardCoins); err != nil {
		panic(err)
	}

	// ------------- Community Pool -------------
	// TODO: mint tokens for community pool
	// TODO: send tokens to community pool

	// ------------- Bookkeeping -------------
	// bookkeep the previous block time
	k.SetPreviousBlockTime(ctx, ctx.BlockTime())

	// TODO: emit event
}
