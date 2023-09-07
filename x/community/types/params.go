package types

import "time"

var (
	DefaultUpgradeTimeDisableInflation = time.Time{}
)

// NewParams returns a new params object
func NewParams(upgradeTime time.Time) Params {
	return Params{
		UpgradeTimeDisableInflation: upgradeTime,
	}
}

// DefaultParams returns default params
func DefaultParams() Params {
	return NewParams(
		DefaultUpgradeTimeDisableInflation,
	)
}

// Validate checks the params are valid
func (p Params) Validate() error {
	return nil
}
