package cdp_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/cdp"
	"github.com/kava-labs/kava/x/pricefeed"
	"github.com/kava-labs/kava/app"
)

// Avoid cluttering test cases with long function name
func i(in int64) sdk.Int                    { return sdk.NewInt(in) }
func d(str string) sdk.Dec                  { return sdk.MustNewDecFromStr(str) }
func c(denom string, amount int64) sdk.Coin { return sdk.NewInt64Coin(denom, amount) }
func cs(coins ...sdk.Coin) sdk.Coins        { return sdk.NewCoins(coins...) }

func NewPFGenState(asset string, price sdk.Dec) app.GenesisState {
	quote := "usd"
	ap := pricefeed.Params{
		Markets: []pricefeed.Market{
			pricefeed.Market{MarketID: asset, BaseAsset: asset, QuoteAsset: quote, Oracles: pricefeed.Oracles{}, Active: true},
		},
	}
	pfGenesis := pricefeed.GenesisState{
		Params: ap,
		PostedPrices: []pricefeed.PostedPrice{
			pricefeed.PostedPrice{
				MarketID:      asset,
				OracleAddress: sdk.AccAddress{},
				Price:         price,
				Expiry:        time.Unix(9999999999, 0), // some deterministic future date,
			},
		},
	}
	return app.GenesisState{pricefeed.ModuleName: pricefeed.ModuleCdc.MustMarshalJSON(pfGenesis)}
}

func NewCDPGenState(asset string, liquidationRatio sdk.Dec) app.GenesisState {
	cdpGenesis := cdp.GenesisState{
		Params: cdp.CdpParams{
			GlobalDebtLimit: sdk.NewInt(1000000),
			CollateralParams: []cdp.CollateralParams{
				{
					Denom:            asset,
					LiquidationRatio: liquidationRatio,
					DebtLimit:        sdk.NewInt(500000),
				},
			},
		},
		GlobalDebt: sdk.ZeroInt(),
		CDPs:       cdp.CDPs{},
	}
	return app.GenesisState{cdp.ModuleName: cdp.ModuleCdc.MustMarshalJSON(cdpGenesis)}
}