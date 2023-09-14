package types

import (
	"errors"
	fmt "fmt"
	"time"

	sdkmath "cosmossdk.io/math"
)

var (
	DefaultUpgradeTimeDisableInflation = time.Time{}
	// DefaultRewardsPerSecond is ~4.6 KAVA per block, 6.3s block time
	DefaultRewardsPerSecond = sdkmath.NewInt(744191)
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
		return fmt.Errorf("rewards per second should not be negative: %s", p.RewardsPerSecond)
	}

	return nil
}
