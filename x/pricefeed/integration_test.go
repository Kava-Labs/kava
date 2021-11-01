package pricefeed_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/pricefeed/types"
)

func NewPricefeedGenStateMulti() app.GenesisState {
	pfGenesis := types.GenesisState{
		Params: types.Params{
			Markets: []types.Market{
				{MarketID: "btc:usd", BaseAsset: "btc", QuoteAsset: "usd", Oracles: []string{}, Active: true},
				{MarketID: "xrp:usd", BaseAsset: "xrp", QuoteAsset: "usd", Oracles: []string{}, Active: true},
			},
		},
		PostedPrices: []types.PostedPrice{
			{
				MarketID:      "btc:usd",
				OracleAddress: "",
				Price:         sdk.MustNewDecFromStr("8000.00"),
				Expiry:        time.Now().Add(1 * time.Hour),
			},
			{
				MarketID:      "xrp:usd",
				OracleAddress: "",
				Price:         sdk.MustNewDecFromStr("0.25"),
				Expiry:        time.Now().Add(1 * time.Hour),
			},
		},
	}
	return app.GenesisState{types.ModuleName: types.ModuleCdc.LegacyAmino.MustMarshalJSON(pfGenesis)}
}

func NewPricefeedGenStateWithOracles(addrs []string) app.GenesisState {
	pfGenesis := types.GenesisState{
		Params: types.Params{
			Markets: []types.Market{
				{MarketID: "btc:usd", BaseAsset: "btc", QuoteAsset: "usd", Oracles: addrs, Active: true},
				{MarketID: "xrp:usd", BaseAsset: "xrp", QuoteAsset: "usd", Oracles: addrs, Active: true},
			},
		},
		PostedPrices: []types.PostedPrice{
			{
				MarketID:      "btc:usd",
				OracleAddress: addrs[0],
				Price:         sdk.MustNewDecFromStr("8000.00"),
				Expiry:        time.Now().Add(1 * time.Hour),
			},
			{
				MarketID:      "xrp:usd",
				OracleAddress: addrs[0],
				Price:         sdk.MustNewDecFromStr("0.25"),
				Expiry:        time.Now().Add(1 * time.Hour),
			},
		},
	}
	return app.GenesisState{types.ModuleName: types.ModuleCdc.LegacyAmino.MustMarshalJSON(pfGenesis)}
}
