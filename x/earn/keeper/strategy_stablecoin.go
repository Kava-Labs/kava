package keeper

import sdk "github.com/cosmos/cosmos-sdk/types"

// StableCoinStrategy defines the stablecoin strategy:
// 1. Mint USDX from stablecoin (USDC, BUSD, USTD, DAI)
// 2. Supply USDX to Lend
type StableCoinStrategy Keeper

func (s *StableCoinStrategy) GetName() string {
	return "stablecoin"
}

func (s *StableCoinStrategy) GetDescription() string {
	return "Mint USDX from stablecoin, then supply the USDX to Lend"
}

func (s *StableCoinStrategy) GetSupportedDenoms() []string {
	return []string{"busd", "usdc", "usdt", "dai"}
}

func (s *StableCoinStrategy) GetEstimatedTotalAssets(denom string) (sdk.Coin, error) {
	// Amount of USDX in Lend

	// Convert USDX to denom in Mint

	return sdk.Coin{}, nil
}

func (s *StableCoinStrategy) Deposit(amount sdk.Coin) error {
	return nil
}

func (s *StableCoinStrategy) Withdraw(amount sdk.Coin) error {
	return nil
}

func (s *StableCoinStrategy) LiquidateAll() (amount sdk.Coin, err error) {
	return sdk.Coin{}, nil
}
