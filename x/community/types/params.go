package types

import (
	fmt "fmt"
	"time"

	sdkmath "cosmossdk.io/math"
)

var (
	DefaultUpgradeTimeDisableInflation = time.Time{}
	// DefaultStakingRewardsPerSecond is ~4.6 KAVA per block, 6.3s block time
	DefaultStakingRewardsPerSecond = sdkmath.LegacyNewDec(744191)
	// DefaultStakingRewardsPerSecond is ~4.6 KAVA per block, 6.3s block time
	DefaultUpgradeTimeSetStakingRewardsPerSecond = sdkmath.LegacyNewDec(744191)
)

// NewParams returns a new params object
func NewParams(
	upgradeTime time.Time,
	stakingRewardsPerSecond sdkmath.LegacyDec,
	upgradeTimeSetstakingRewardsPerSecond sdkmath.LegacyDec,
) Params {
	return Params{
		UpgradeTimeDisableInflation:           upgradeTime,
		StakingRewardsPerSecond:               stakingRewardsPerSecond,
		UpgradeTimeSetStakingRewardsPerSecond: upgradeTimeSetstakingRewardsPerSecond,
	}
}

// DefaultParams returns default params
func DefaultParams() Params {
	return NewParams(
		DefaultUpgradeTimeDisableInflation,
		DefaultStakingRewardsPerSecond,
		DefaultUpgradeTimeSetStakingRewardsPerSecond,
	)
}

// Validate checks the params are valid
func (p Params) Validate() error {
	// p.UpgradeTimeDisableInflation.IsZero() is a valid state. It's taken to mean inflation will be disabled on the block 1.

	if err := validateDecNotNilNonNegative(p.StakingRewardsPerSecond, "StakingRewardsPerSecond"); err != nil {
		return err
	}

	if err := validateDecNotNilNonNegative(p.UpgradeTimeSetStakingRewardsPerSecond, "UpgradeTimeSetStakingRewardsPerSecond"); err != nil {
		return err
	}

	return nil
}

func validateDecNotNilNonNegative(value sdkmath.LegacyDec, name string) error {
	if value.IsNil() {
		return fmt.Errorf("%s should not be nil", name)
	}

	if value.IsNegative() {
		return fmt.Errorf("%s should not be negative: %s", name, value)
	}

	return nil
}
