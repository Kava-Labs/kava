package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/earn/types"
)

// Strategy is the interface that must be implemented by a strategy.
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
	GetEstimatedTotalAssets(ctx sdk.Context, denom string) (sdk.Coin, error)

	// Deposit the specified amount of coins into this strategy. The amount
	// must be denominated in GetDenom().
	Deposit(ctx sdk.Context, amount sdk.Coin) error

	// Withdraw the specified amount of coins from this strategy. The amount
	// must be denominated in GetDenom().
	Withdraw(ctx sdk.Context, amount sdk.Coin) error
}

func (k *Keeper) GetStrategy(strategyType types.StrategyType) (Strategy, error) {
	switch strategyType {
	case types.STRATEGY_TYPE_HARD:
		return (*HardStrategy)(k), nil
	case types.STRATEGY_TYPE_SAVINGS:
		panic("unimplemented")
	default:
		return nil, fmt.Errorf("unknown strategy type: %s", strategyType)
	}
}
