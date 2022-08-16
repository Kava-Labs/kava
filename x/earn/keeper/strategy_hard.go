package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/earn/types"
)

// HardStrategy defines the strategy that deposits assets to Hard
type HardStrategy Keeper

var _ Strategy = (*HardStrategy)(nil)

// GetStrategyType returns the strategy type
func (s *HardStrategy) GetStrategyType() types.StrategyType {
	return types.STRATEGY_TYPE_HARD
}

// GetEstimatedTotalAssets returns the current value of all assets deposited
// in hard.
func (s *HardStrategy) GetEstimatedTotalAssets(ctx sdk.Context, denom string) (sdk.Coin, error) {
	macc := s.accountKeeper.GetModuleAccount(ctx, types.ModuleName)
	deposit, found := s.hardKeeper.GetSyncedDeposit(ctx, macc.GetAddress())
	if !found {
		// Return 0 if no deposit exists for module account
		return sdk.NewCoin(denom, sdk.ZeroInt()), nil
	}

	// Only return the deposit for the vault denom.
	for _, coin := range deposit.Amount {
		if coin.Denom == denom {
			return coin, nil
		}
	}

	// Return 0 if no deposit exists for the vault denom
	return sdk.NewCoin(denom, sdk.ZeroInt()), nil
}

// Deposit deposits the specified amount of coins into hard.
func (s *HardStrategy) Deposit(ctx sdk.Context, amount sdk.Coin) error {
	macc := s.accountKeeper.GetModuleAccount(ctx, types.ModuleName)
	return s.hardKeeper.Deposit(ctx, macc.GetAddress(), sdk.NewCoins(amount))
}

// Withdraw withdraws the specified amount of coins from hard.
func (s *HardStrategy) Withdraw(ctx sdk.Context, amount sdk.Coin) error {
	macc := s.accountKeeper.GetModuleAccount(ctx, types.ModuleName)
	return s.hardKeeper.Withdraw(ctx, macc.GetAddress(), sdk.NewCoins(amount))
}
