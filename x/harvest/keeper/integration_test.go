package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/pricefeed"
)

func NewPricefeedGenStateMulti() app.GenesisState {
	pfGenesis := pricefeed.GenesisState{
		Params: pricefeed.Params{
			Markets: []pricefeed.Market{
				{MarketID: "btc:usd", BaseAsset: "btc", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
				{MarketID: "xrp:usd", BaseAsset: "xrp", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
				{MarketID: "bnb:usd", BaseAsset: "bnb", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
				{MarketID: "kava:usd", BaseAsset: "kava", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
			},
		},
		PostedPrices: []pricefeed.PostedPrice{
			{
				MarketID:      "btc:usd",
				OracleAddress: sdk.AccAddress{},
				Price:         sdk.MustNewDecFromStr("8000.00"),
				Expiry:        time.Now().Add(1 * time.Hour),
			},
			{
				MarketID:      "xrp:usd",
				OracleAddress: sdk.AccAddress{},
				Price:         sdk.MustNewDecFromStr("0.25"),
				Expiry:        time.Now().Add(1 * time.Hour),
			},
			{
				MarketID:      "bnb:usd",
				OracleAddress: sdk.AccAddress{},
				Price:         sdk.MustNewDecFromStr("17.25"),
				Expiry:        time.Now().Add(1 * time.Hour),
			},
			{
				MarketID:      "kava:usd",
				OracleAddress: sdk.AccAddress{},
				Price:         sdk.MustNewDecFromStr("5.00"),
				Expiry:        time.Now().Add(1 * time.Hour),
			},
		},
	}
	return app.GenesisState{pricefeed.ModuleName: pricefeed.ModuleCdc.MustMarshalJSON(pfGenesis)}
}
