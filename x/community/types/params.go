package types

import (
	fmt "fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	DefaultUpgradeTimeDisableInflation = time.Time{}
	DefaultRewardsPerSecond            = sdk.NewCoin("ukava", sdkmath.ZeroInt())
)

// NewParams returns a new params object
func NewParams(upgradeTime time.Time, rewardsPerSecond sdk.Coin) Params {
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
	if err := p.RewardsPerSecond.Validate(); err != nil {
		return fmt.Errorf("invalid rewards per second: %w", err)
	}

	return nil
}
