package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RewardPeriod stores the state of an ongoing reward
type RewardPeriod struct {
	Denom         string        `json:"denom" yaml:"denom"`
	Start         time.Time     `json:"start" yaml:"start"`
	End           time.Time     `json:"end" yaml:"end"`
	Reward        sdk.Coin      `json:"reward" yaml:"reward"` // per second reward payouts. For example, if we know from params that 10000KAVA is being paid out over 1 week (604800 rewards periods), then the value of reward would be (10000 * 1000000)/604800 = 16534ukava per second
	ClaimEnd      time.Time     `json:"claim_end" yaml:"claim_end"`
	ClaimTimeLock time.Duration `json:"claim_time_lock" yaml:"claim_time_lock"` // the amount of time rewards are timelocked once they are sent to users
}

// ClaimPeriod stores the state of an ongoing claim period
type ClaimPeriod struct {
	Denom string    `json:"denom" yaml:"denom"`
	ID    uint64    `json:"id" yaml:"id"`
	End   time.Time `json:"end" yaml:"end"`
}

// Claim stores the rewards that can be claimed by owner
type Claim struct {
	Owner  sdk.AccAddress `json:"owner" yaml:"owner"`
	Reward sdk.Coin       `json:"reward" yaml:"reward"`
	ID     uint64         `json:"id" yaml:"id"`
}
