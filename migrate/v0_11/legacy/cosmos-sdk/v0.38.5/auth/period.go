package v38_5

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Period defines a length of time and amount of coins that will vest
type Period struct {
	Length int64     `json:"length" yaml:"length"` // length of the period, in seconds
	Amount sdk.Coins `json:"amount" yaml:"amount"` // amount of coins vesting during this period
}

// Periods stores all vesting periods passed as part of a PeriodicVestingAccount
type Periods []Period
