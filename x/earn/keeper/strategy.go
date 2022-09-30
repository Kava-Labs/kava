package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/earn/types"
)

// Strategy is the interface that must be implemented by a strategy.
type Strategy interface {
	// GetStrategyType returns the strategy type
	GetStrategyType() types.StrategyType

	// GetEstimatedTotalAssets returns the estimated total assets of the
	// strategy with the specified denom. This is the value if the strategy were
	// to liquidate all assets.
	//
	// **Note:** This may not reflect the true value as it may become outdated
	// from market changes.
	GetEstimatedTotalAssets(ctx sdk.Context, denom string) (sdk.Coin, error)

	// Deposit the specified amount of coins into this strategy.
	Deposit(ctx sdk.Context, amount sdk.Coin) error

	// Withdraw the specified amount of coins from this strategy.
	Withdraw(ctx sdk.Context, amount sdk.Coin) error
}

// GetStrategy returns the strategy for the given strategy type.
func (k *Keeper) GetStrategy(strategyType types.StrategyType) (Strategy, error) {
	switch strategyType {
	case types.STRATEGY_TYPE_HARD:
		return (*HardStrategy)(k), nil
	case types.STRATEGY_TYPE_SAVINGS:
		return (*SavingsStrategy)(k), nil
	default:
		return nil, fmt.Errorf("unknown strategy type: %s", strategyType)
	}
}
