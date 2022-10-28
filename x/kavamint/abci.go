package kavamint

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/kavamint/keeper"
)

// BeginBlocker mints & distributes new tokens for the previous block.
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	params := k.GetParams(ctx)

	// ------------- Staking Rewards -------------
	// number of tokens minted for staking rewards is total_bonded_tokens * apy
	totalBonded := k.TotalBondedTokens(ctx)
	stakingRewardsAmount := params.StakingRewardsApy.MulInt(totalBonded).TruncateInt()
	stakingRewardCoins := sdk.NewCoins(sdk.NewCoin(k.BondDenom(ctx), stakingRewardsAmount))

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

	// TODO: emit event
}
