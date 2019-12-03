package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/cdp/types"
)

// Avoid cluttering test cases with long function name
func i(in int64) sdk.Int                    { return sdk.NewInt(in) }
func d(str string) sdk.Dec                  { return sdk.MustNewDecFromStr(str) }
func c(denom string, amount int64) sdk.Coin { return sdk.NewInt64Coin(denom, amount) }
func cs(coins ...sdk.Coin) sdk.Coins        { return sdk.NewCoins(coins...) }

func defaultParamsMulti() types.CdpParams {
	return types.CdpParams{
		GlobalDebtLimit: sdk.NewInt(1000000),
		CollateralParams: []types.CollateralParams{
			{
				Denom:            "btc",
				LiquidationRatio: sdk.MustNewDecFromStr("1.5"),
				DebtLimit:        sdk.NewInt(500000),
			},
			{
				Denom:            "xrp",
				LiquidationRatio: sdk.MustNewDecFromStr("2.0"),
				DebtLimit:        sdk.NewInt(500000),
			},
		},
		StableDenoms: []string{"usdx"},
	}
}

func defaultParamsSingle() types.CdpParams {
	return types.CdpParams{
		GlobalDebtLimit: sdk.NewInt(1000000),
		CollateralParams: []types.CollateralParams{
			{
				Denom:            "xrp",
				LiquidationRatio: sdk.MustNewDecFromStr("2.0"),
				DebtLimit:        sdk.NewInt(500000),
			},
		},
		StableDenoms: []string{"usdx"},
	}
}
