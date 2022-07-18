package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/earn/types"
)

// HardStrategy defines the strategy that deposits assets to Hard
type HardStrategy Keeper

var _ Strategy = (*HardStrategy)(nil)

func (s *HardStrategy) GetName() string {
	return "Lend"
}

func (s *HardStrategy) GetDescription() string {
	return "Supplies assets to Lend"
}

func (s *HardStrategy) GetSupportedDenoms() []string {
	return []string{"usdx"}
}

func (s *HardStrategy) GetEstimatedTotalAssets(ctx sdk.Context, denom string) (sdk.Coin, error) {
	macc := s.accountKeeper.GetModuleAccount(ctx, types.ModuleName)
	borrow, found := s.hardKeeper.GetSyncedDeposit(ctx, macc.GetAddress())
	if !found {
		// Return 0 if no borrow exists for module account
		return sdk.NewCoin(denom, sdk.ZeroInt()), nil
	}

	// Only return the borrow for the provided denom.
	for _, coin := range borrow.Amount {
		if coin.Denom == denom {
			return coin, nil
		}
	}

	// Return 0 if no borrow exists for the denom
	return sdk.NewCoin(denom, sdk.ZeroInt()), nil
}

func (s *HardStrategy) Deposit(ctx sdk.Context, amount sdk.Coin) error {
	macc := s.accountKeeper.GetModuleAccount(ctx, types.ModuleName)
	return s.hardKeeper.Deposit(ctx, macc.GetAddress(), sdk.NewCoins(amount))
}

func (s *HardStrategy) Withdraw(ctx sdk.Context, amount sdk.Coin) error {
	macc := s.accountKeeper.GetModuleAccount(ctx, types.ModuleName)
	return s.hardKeeper.Withdraw(ctx, macc.GetAddress(), sdk.NewCoins(amount))
}

// LiquidateAll liquidates all assets in the strategy, this should be called
// only in case of emergency or when all assets should be moved to a new
// strategy.
func (s *HardStrategy) LiquidateAll(ctx sdk.Context, denom string) (amount sdk.Coin, err error) {
	totalAssets, err := s.GetEstimatedTotalAssets(ctx, denom)
	if err != nil {
		return sdk.Coin{}, err
	}

	macc := s.accountKeeper.GetModuleAccount(ctx, types.ModuleName)
	if err := s.hardKeeper.Withdraw(ctx, macc.GetAddress(), sdk.NewCoins(totalAssets)); err != nil {
		return sdk.Coin{}, err
	}

	return totalAssets, nil
}
