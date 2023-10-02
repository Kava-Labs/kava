package types

import (
	"errors"
	"time"

	sdkmath "cosmossdk.io/math"
)

var (
	// DefaultLastAccumulationTime is zero
	DefaultLastAccumulationTime = time.Time{}
	// DefaultLastTruncationError is zero
	DefaultLastTruncationError = sdkmath.LegacyZeroDec()
)

// NewStakingRewardsState returns a new staking rewards state object
func NewStakingRewardsState(
	lastAccumulationTime time.Time,
	lastTruncationError sdkmath.LegacyDec,
) StakingRewardsState {
	return StakingRewardsState{
		LastAccumulationTime: lastAccumulationTime,
		LastTruncationError:  lastTruncationError,
	}
}

// DefaultStakingRewardsState returns default params
func DefaultStakingRewardsState() StakingRewardsState {
	return NewStakingRewardsState(
		DefaultLastAccumulationTime,
		DefaultLastTruncationError,
	)
}

// Validate checks the params are valid
func (p StakingRewardsState) Validate() error {
	if err := validateDecNotNilNonNegative(p.LastTruncationError, "LastTruncationError"); err != nil {
		return err
	}

	if p.LastTruncationError.GTE(sdkmath.LegacyOneDec()) {
		return errors.New("LastTruncationError should not be greater or equal to 1")
	}

	if p.LastAccumulationTime.IsZero() && !p.LastTruncationError.IsZero() {
		return errors.New("LastTruncationError should be zero if last accumulation time is zero")
	}

	return nil
}
