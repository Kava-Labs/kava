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
	communityPoolInflation, err := k.AccumulateCommunityPoolInflation(ctx, previousBlockTime)
	if err != nil {
		panic(err)
	}

	// mint community pool inflation
	if err := k.MintCoins(ctx, communityPoolInflation); err != nil {
		panic(err)
	}

	// send inflation coins to the community pool (x/community module account)
	if err := k.FundCommunityPool(ctx, communityPoolInflation); err != nil {
		panic(err)
	}

	// ------------- Bookkeeping -------------
	// bookkeep the previous block time
	k.SetPreviousBlockTime(ctx, ctx.BlockTime())

	// TODO: emit event
}
