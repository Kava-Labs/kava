package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/earn/types"
)

// LendStrategy defines the strategy that deposits assets to Lend
type LendStrategy Keeper

var _ Strategy = (*LendStrategy)(nil)

func (s *LendStrategy) GetName() string {
	return "Lend"
}

func (s *LendStrategy) GetDescription() string {
	return "Supplies assets to Lend"
}

func (s *LendStrategy) GetSupportedDenoms() []string {
	return []string{"usdx"}
}

func (s *LendStrategy) GetEstimatedTotalAssets(ctx sdk.Context, denom string) (sdk.Coin, error) {
	macc := s.accountKeeper.GetModuleAccount(ctx, types.ModuleName)
	borrow, found := s.hardKeeper.GetSyncedBorrow(ctx, macc.GetAddress())
	if !found {
		// Return 0 if no borrow exists for module account
		return sdk.NewCoin(denom, sdk.ZeroInt()), nil
	}

	for _, coin := range borrow.Amount {
		if coin.Denom == denom {
			return coin, nil
		}
	}

	// Return 0 if no borrow exists for the denom
	return sdk.NewCoin(denom, sdk.ZeroInt()), nil
}

func (s *LendStrategy) Deposit(ctx sdk.Context, amount sdk.Coin) error {
	macc := s.accountKeeper.GetModuleAccount(ctx, types.ModuleName)
	return s.hardKeeper.Deposit(ctx, macc.GetAddress(), sdk.NewCoins(amount))
}

func (s *LendStrategy) Withdraw(ctx sdk.Context, amount sdk.Coin) error {
	macc := s.accountKeeper.GetModuleAccount(ctx, types.ModuleName)
	return s.hardKeeper.Withdraw(ctx, macc.GetAddress(), sdk.NewCoins(amount))
}

// LiquidateAll liquidates all assets in the strategy, this should be called
// only in case of emergency or when all assets should be moved to a new
// strategy.
func (s *LendStrategy) LiquidateAll(ctx sdk.Context, denom string) (amount sdk.Coin, err error) {
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
