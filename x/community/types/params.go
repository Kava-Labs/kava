package types

import (
	"errors"
	fmt "fmt"
	"time"

	sdkmath "cosmossdk.io/math"
)

var (
	DefaultUpgradeTimeDisableInflation = time.Time{}
	// DefaultStakingRewardsPerSecond is ~4.6 KAVA per block, 6.3s block time
	DefaultStakingRewardsPerSecond = sdkmath.LegacyNewDec(744191)
)

// NewParams returns a new params object
func NewParams(
	upgradeTime time.Time,
	stakingRewardsPerSecond sdkmath.LegacyDec,
) Params {
	return Params{
		UpgradeTimeDisableInflation: upgradeTime,
		StakingRewardsPerSecond:     stakingRewardsPerSecond,
	}
}

// DefaultParams returns default params
func DefaultParams() Params {
	return NewParams(
		DefaultUpgradeTimeDisableInflation,
		DefaultStakingRewardsPerSecond,
	)
}

// Validate checks the params are valid
func (p Params) Validate() error {
	// p.UpgradeTimeDisableInflation.IsZero() is a valid state. It's taken to mean inflation will be disabled on the block 1.

	if p.StakingRewardsPerSecond.IsNil() {
		return errors.New("StakingRewardsPerSecond should not be nil")
	}

	if p.StakingRewardsPerSecond.IsNegative() {
		return fmt.Errorf("StakingRewardsPerSecond should not be negative: %s", p.StakingRewardsPerSecond)
	}

	return nil
}
