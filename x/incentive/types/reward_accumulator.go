package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type RewardAccumulator interface {
	AccumulateRewards(
		ctx sdk.Context,
		claimType ClaimType,
		rewardPeriod MultiRewardPeriod,
	) error
}
