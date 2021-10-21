package v18de63

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// historical rewards for a validator
// height is implicit within the store key
// cumulative reward ratio is the sum from the zeroeth period
// until this period of rewards / tokens, per the spec
// The reference count indicates the number of objects
// which might need to reference this historical entry
// at any point.
// ReferenceCount =
//    number of outstanding delegations which ended the associated period (and might need to read that record)
//  + number of slashes which ended the associated period (and might need to read that record)
//  + one per validator for the zeroeth period, set on initialization
type ValidatorHistoricalRewards struct {
	CumulativeRewardRatio sdk.DecCoins `json:"cumulative_reward_ratio" yaml:"cumulative_reward_ratio"`
	ReferenceCount        uint16       `json:"reference_count" yaml:"reference_count"`
}

// current rewards and current period for a validator
// kept as a running counter and incremented each block
// as long as the validator's tokens remain constant
type ValidatorCurrentRewards struct {
	Rewards sdk.DecCoins `json:"rewards" yaml:"rewards"` // current rewards
	Period  uint64       `json:"period" yaml:"period"`   // current period
}

// accumulated commission for a validator
// kept as a running counter, can be withdrawn at any time
type ValidatorAccumulatedCommission = sdk.DecCoins

// validator slash event
// height is implicit within the store key
// needed to calculate appropriate amounts of staking token
// for delegations which withdraw after a slash has occurred
type ValidatorSlashEvent struct {
	ValidatorPeriod uint64  `json:"validator_period" yaml:"validator_period"` // period when the slash occurred
	Fraction        sdk.Dec `json:"fraction" yaml:"fraction"`                 // slash fraction
}

// ValidatorSlashEvents is a collection of ValidatorSlashEvent
type ValidatorSlashEvents []ValidatorSlashEvent

// outstanding (un-withdrawn) rewards for a validator
// inexpensive to track, allows simple sanity checks
type ValidatorOutstandingRewards = sdk.DecCoins
