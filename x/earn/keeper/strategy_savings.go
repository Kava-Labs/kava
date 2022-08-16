package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/earn/types"
)

// SavingsStrategy defines the strategy that deposits assets to x/savings
type SavingsStrategy Keeper

var _ Strategy = (*SavingsStrategy)(nil)

func (s *SavingsStrategy) GetStrategyType() types.StrategyType {
	return types.STRATEGY_TYPE_SAVINGS
}

func (s *SavingsStrategy) GetEstimatedTotalAssets(ctx sdk.Context, denom string) (sdk.Coin, error) {
	macc := s.accountKeeper.GetModuleAccount(ctx, types.ModuleName)
	deposit, found := s.savingsKeeper.GetDeposit(ctx, macc.GetAddress())
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

func (s *SavingsStrategy) Deposit(ctx sdk.Context, amount sdk.Coin) error {
	macc := s.accountKeeper.GetModuleAccount(ctx, types.ModuleName)
	return s.savingsKeeper.Deposit(ctx, macc.GetAddress(), sdk.NewCoins(amount))
}

func (s *SavingsStrategy) Withdraw(ctx sdk.Context, amount sdk.Coin) error {
	macc := s.accountKeeper.GetModuleAccount(ctx, types.ModuleName)
	return s.savingsKeeper.Withdraw(ctx, macc.GetAddress(), sdk.NewCoins(amount))
}
