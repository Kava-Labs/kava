package pricefeed_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/pricefeed/types"
)

func NewPricefeedGen() types.GenesisState {
	return types.GenesisState{
		Params: types.Params{
			Markets: []types.Market{
				{MarketID: "btc:usd", BaseAsset: "btc", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
				{MarketID: "xrp:usd", BaseAsset: "xrp", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
			},
		},
		PostedPrices: []types.PostedPrice{
			{
				MarketID:      "btc:usd",
				OracleAddress: sdk.AccAddress("oracle1"),
				Price:         sdk.MustNewDecFromStr("8000.00"),
				Expiry:        time.Now().Add(1 * time.Hour),
			},
			{
				MarketID:      "xrp:usd",
				OracleAddress: sdk.AccAddress("oracle2"),
				Price:         sdk.MustNewDecFromStr("0.25"),
				Expiry:        time.Now().Add(1 * time.Hour),
			},
		},
	}
}

func NewPricefeedGenStateMulti() app.GenesisState {
	pfGenesis := NewPricefeedGen()
	return app.GenesisState{types.ModuleName: types.ModuleCdc.LegacyAmino.MustMarshalJSON(pfGenesis)}
}

func NewPricefeedGenStateWithOracles(addrs []sdk.AccAddress) app.GenesisState {
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
