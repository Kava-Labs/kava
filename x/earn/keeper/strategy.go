package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Strategy interface {
	// GetName returns the name of the strategy.
	GetName() string

	// GetDescription returns the description of the strategy.
	GetDescription() string

	// GetSupportedDenoms returns a slice of supported denom for this strategy.
	// For example, stablecoin stakers strategy supports both "busd" and "usdc".
	GetSupportedDenoms() []string

	// GetEstimatedTotalAssets returns the estimated total assets denominated in
	// GetDenom() of this strategy. This is the value if the strategy were to
	// liquidate all assets.
	//
	// **Note:** This may not reflect the true value as it may become outdated
	// from market changes.
	GetEstimatedTotalAssets(denom string) (sdk.Coin, error)

	// Deposit the specified amount of coins into this strategy. The amount
	// must be denominated in GetDenom().
	Deposit(amount sdk.Coin) error

	// Withdraw the specified amount of coins from this strategy. The amount
	// must be denominated in GetDenom().
	Withdraw(amount sdk.Coin) error

	// LiquidateAll liquidates all of the entire strategy's positions, returning
	// the amount of liquidated denominated in GetDenom(). This should be only
	// called during use of emergency via governance.
	LiquidateAll() (amount sdk.Coin, err error)
}
