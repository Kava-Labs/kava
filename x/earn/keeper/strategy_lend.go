package keeper

import sdk "github.com/cosmos/cosmos-sdk/types"

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

func (s *LendStrategy) GetEstimatedTotalAssets(denom string) (sdk.Coin, error) {
	// 1. Get amount in Lend

	return sdk.Coin{}, nil
}

func (s *LendStrategy) Deposit(amount sdk.Coin) error {
	return nil
}

func (s *LendStrategy) Withdraw(amount sdk.Coin) error {
	return nil
}

// LiquidateAll liquidates all assets in the strategy, this should be called
// only in case of emergency or when all assets should be moved to a new
// strategy.
func (s *LendStrategy) LiquidateAll() (amount sdk.Coin, err error) {
	return sdk.Coin{}, nil
}
