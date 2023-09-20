package types

import errorsmod "cosmossdk.io/errors"

var (
	ErrStakingRewardsPayout = errorsmod.Register(ModuleName, 1, "staking rewards payout error")
)
