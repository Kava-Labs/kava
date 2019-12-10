package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/cdp"
	"github.com/kava-labs/kava/x/cdp/types"
	"github.com/kava-labs/kava/x/pricefeed"
)

// Avoid cluttering test cases with long function name
func i(in int64) sdk.Int                    { return sdk.NewInt(in) }
func d(str string) sdk.Dec                  { return sdk.MustNewDecFromStr(str) }
func c(denom string, amount int64) sdk.Coin { return sdk.NewInt64Coin(denom, amount) }
func cs(coins ...sdk.Coin) sdk.Coins        { return sdk.NewCoins(coins...) }

func NewPricefeedGenState(asset string, price sdk.Dec) app.GenesisState {
	pfGenesis := pricefeed.GenesisState{
		Params: pricefeed.Params{
			Markets: []pricefeed.Market{
				pricefeed.Market{MarketID: asset, BaseAsset: asset, QuoteAsset: "usd", Oracles: pricefeed.Oracles{}, Active: true},
			},
		},
		PostedPrices: []pricefeed.PostedPrice{
			pricefeed.PostedPrice{
				MarketID:      asset,
				OracleAddress: sdk.AccAddress{},
				Price:         price,
				Expiry:        time.Now().Add(1 * time.Hour),
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

func NewPricefeedGenStateMulti() app.GenesisState {
	pfGenesis := pricefeed.GenesisState{
		Params: pricefeed.Params{
			Markets: []pricefeed.Market{
				pricefeed.Market{MarketID: "btc", BaseAsset: "btc", QuoteAsset: "usd", Oracles: pricefeed.Oracles{}, Active: true},
				pricefeed.Market{MarketID: "xrp", BaseAsset: "xrp", QuoteAsset: "usd", Oracles: pricefeed.Oracles{}, Active: true},
			},
		},
		PostedPrices: []pricefeed.PostedPrice{
			pricefeed.PostedPrice{
				MarketID:      "btc",
				OracleAddress: sdk.AccAddress{},
				Price:         sdk.MustNewDecFromStr("8000.00"),
				Expiry:        time.Now().Add(1 * time.Hour),
			},
			pricefeed.PostedPrice{
				MarketID:      "xrp",
				OracleAddress: sdk.AccAddress{},
				Price:         sdk.MustNewDecFromStr("0.25"),
				Expiry:        time.Now().Add(1 * time.Hour),
			},
		},
	}
	return app.GenesisState{pricefeed.ModuleName: pricefeed.ModuleCdc.MustMarshalJSON(pfGenesis)}
}
func NewCDPGenStateMulti() app.GenesisState {
	cdpGenesis := cdp.GenesisState{
		Params: cdp.CdpParams{
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
		},
		GlobalDebt: sdk.ZeroInt(),
		CDPs:       cdp.CDPs{},
	}
	return app.GenesisState{cdp.ModuleName: cdp.ModuleCdc.MustMarshalJSON(cdpGenesis)}
}
