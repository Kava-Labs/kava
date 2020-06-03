package v18de63

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BaseVestingAccount implements the VestingAccount interface. It contains all
// the necessary fields needed for any vesting account implementation.
type BaseVestingAccount struct {
	*BaseAccount

	OriginalVesting  sdk.Coins `json:"original_vesting" yaml:"original_vesting"`   // coins in account upon initialization
	DelegatedFree    sdk.Coins `json:"delegated_free" yaml:"delegated_free"`       // coins that are vested and delegated
	DelegatedVesting sdk.Coins `json:"delegated_vesting" yaml:"delegated_vesting"` // coins that vesting and delegated
	EndTime          int64     `json:"end_time" yaml:"end_time"`                   // when the coins become unlocked
}

// ContinuousVestingAccount implements the VestingAccount interface. It
// continuously vests by unlocking coins linearly with respect to time.
type ContinuousVestingAccount struct {
	*BaseVestingAccount

	StartTime int64 `json:"start_time" yaml:"start_time"` // when the coins start to vest
}

// PeriodicVestingAccount implements the VestingAccount interface. It
// periodically vests by unlocking coins during each specified period
type PeriodicVestingAccount struct {
	*BaseVestingAccount
	StartTime      int64   `json:"start_time" yaml:"start_time"`           // when the coins start to vest
	VestingPeriods Periods `json:"vesting_periods" yaml:"vesting_periods"` // the vesting schedule
}

// DelayedVestingAccount implements the VestingAccount interface. It vests all
// coins after a specific time, but non prior. In other words, it keeps them
// locked until a specified time.
type DelayedVestingAccount struct {
	*BaseVestingAccount
}
