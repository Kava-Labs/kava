package keeper

import sdk "github.com/cosmos/cosmos-sdk/types"

// StableCoinStrategy defines the stablecoin strategy:
// 1. Supply USDX to Lend
type StableCoinStrategy Keeper

var _ Strategy = (*StableCoinStrategy)(nil)

func (s *StableCoinStrategy) GetName() string {
	return "USDX"
}

func (s *StableCoinStrategy) GetDescription() string {
	return "Supplies the USDX to Lend"
}

func (s *StableCoinStrategy) GetSupportedDenoms() []string {
	return []string{"usdx"}
}

func (s *StableCoinStrategy) GetEstimatedTotalAssets(denom string) (sdk.Coin, error) {
	// 1. Get amount of USDX in Lend

	return sdk.Coin{}, nil
}

func (s *StableCoinStrategy) Deposit(amount sdk.Coin) error {
	return nil
}

func (s *StableCoinStrategy) Withdraw(amount sdk.Coin) error {
	return nil
}

// LiquidateAll liquidates all assets in the strategy, this should be called
// only in case of emergency or when all assets should be moved to a new
// strategy.
func (s *StableCoinStrategy) LiquidateAll() (amount sdk.Coin, err error) {
	return sdk.Coin{}, nil
}
