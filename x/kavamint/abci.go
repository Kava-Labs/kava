package kavamint

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/kavamint/keeper"
	"github.com/kava-labs/kava/x/kavamint/types"
)

// BeginBlocker mints & distributes new tokens for the previous block.
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	previousBlockTime, found := k.GetPreviousBlockTime(ctx)
	if !found {
		previousBlockTime = ctx.BlockTime()
	}
	secondsPassed := ctx.BlockTime().Sub(previousBlockTime).Seconds()
	// calculate totals before any minting is done to prevent new mints affecting the values
	totalSupply := k.TotalSupply(ctx)
	totalBonded := k.TotalBondedTokens(ctx)

	// ------------- Staking Rewards -------------
	stakingRewardCoins, err := k.AccumulateStakingRewards(ctx, totalBonded, previousBlockTime)
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
	communityPoolInflation, err := k.AccumulateCommunityPoolInflation(ctx, totalSupply, previousBlockTime)
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

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeMint,
			sdk.NewAttribute(types.AttributeKeyTotalSupply, totalSupply.String()),
			sdk.NewAttribute(types.AttributeKeyTotalBonded, totalBonded.String()),
			sdk.NewAttribute(types.AttributeSecondsPassed, fmt.Sprintf("%f", secondsPassed)),
			sdk.NewAttribute(types.AttributeKeyCommunityPoolMint, communityPoolInflation.String()),
			sdk.NewAttribute(types.AttributeKeyStakingRewardMint, stakingRewardCoins.String()),
		),
	)
}
