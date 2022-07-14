package keeper

import sdk "github.com/cosmos/cosmos-sdk/types"

// StableCoinStrategy defines the stablecoin strategy:
// 1. Mint USDX from stablecoin (USDC, BUSD, USTD, DAI)
// 2. Supply USDX to Lend
type StableCoinStrategy Keeper

var _ Strategy = (*StableCoinStrategy)(nil)

func (s *StableCoinStrategy) GetName() string {
	return "Stablecoin Stakers"
}

func (s *StableCoinStrategy) GetDescription() string {
	return "Mints USDX from stablecoin, then supplies the USDX to Lend"
}

func (s *StableCoinStrategy) GetSupportedDenoms() []string {
	return []string{"busd", "usdc", "usdt", "dai"}
}

func (s *StableCoinStrategy) GetEstimatedTotalAssets(denom string) (sdk.Coin, error) {
	// 1. Get amount of USDX in Lend

	// 2. Get denom amount if repaying minted USDX

	// 3. Get denom amount if using Kava Swap for remaining extra USDX from supply APY

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
