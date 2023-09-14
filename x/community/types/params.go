package types

import (
	"errors"
	"time"

	sdkmath "cosmossdk.io/math"
)

var (
	DefaultUpgradeTimeDisableInflation = time.Time{}
	DefaultRewardsPerSecond            = sdkmath.ZeroInt()
)

// NewParams returns a new params object
func NewParams(upgradeTime time.Time, rewardsPerSecond sdkmath.Int) Params {
	return Params{
		UpgradeTimeDisableInflation: upgradeTime,
		RewardsPerSecond:            rewardsPerSecond,
	}
}

// DefaultParams returns default params
func DefaultParams() Params {
	return NewParams(
		DefaultUpgradeTimeDisableInflation,
		DefaultRewardsPerSecond,
	)
}

// Validate checks the params are valid
func (p Params) Validate() error {
	if p.RewardsPerSecond.IsNil() {
		return errors.New("rewards per second should not be nil")
	}

	if p.RewardsPerSecond.IsNegative() {
		return errors.New("rewards per second should not be negative")
	}

	return nil
}
