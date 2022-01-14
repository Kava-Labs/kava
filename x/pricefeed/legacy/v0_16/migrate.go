package v0_16

import (
	"github.com/cosmos/cosmos-sdk/types"
	v015pricefeed "github.com/kava-labs/kava/x/pricefeed/legacy/v0_15"
	v016pricefeed "github.com/kava-labs/kava/x/pricefeed/types"
)

var NewIBCMarkets = []v016pricefeed.Market{
	{
		MarketID:   "atom:usd",
		BaseAsset:  "atom",
		QuoteAsset: "usd",
		Oracles:    nil,
		Active:     true,
	},
	{
		MarketID:   "atom:usd:30",
		BaseAsset:  "atom",
		QuoteAsset: "usd",
		Oracles:    nil,
		Active:     true,
	},
	{
		MarketID:   "akt:usd",
		BaseAsset:  "akt",
		QuoteAsset: "usd",
		Oracles:    nil,
		Active:     true,
	},
	{
		MarketID:   "akt:usd:30",
		BaseAsset:  "akt",
		QuoteAsset: "usd",
		Oracles:    nil,
		Active:     true,
	},
	{
		MarketID:   "luna:usd",
		BaseAsset:  "luna",
		QuoteAsset: "usd",
		Oracles:    nil,
		Active:     true,
	},
	{
		MarketID:   "luna:usd:30",
		BaseAsset:  "luna",
		QuoteAsset: "usd",
		Oracles:    nil,
		Active:     true,
	},
	{
		MarketID:   "osmo:usd",
		BaseAsset:  "osmo",
		QuoteAsset: "usd",
		Oracles:    nil,
		Active:     true,
	},
	{
		MarketID:   "osmo:usd:30",
		BaseAsset:  "osmo",
		QuoteAsset: "usd",
		Oracles:    nil,
		Active:     true,
	},
	{
		MarketID:   "ust:usd",
		BaseAsset:  "ust",
		QuoteAsset: "usd",
		Oracles:    nil,
		Active:     true,
	},
	{
		MarketID:   "ust:usd:30",
		BaseAsset:  "ust",
		QuoteAsset: "usd",
		Oracles:    nil,
		Active:     true,
	},
}

func migrateParams(params v015pricefeed.Params) v016pricefeed.Params {
	markets := make(v016pricefeed.Markets, len(params.Markets))
	for i, market := range params.Markets {
		markets[i] = v016pricefeed.Market{
			MarketID:   market.MarketID,
			BaseAsset:  market.BaseAsset,
			QuoteAsset: market.QuoteAsset,
			Oracles:    market.Oracles,
			Active:     market.Active,
		}
	}

	markets = addIbcMarkets(markets)

	return v016pricefeed.Params{Markets: markets}
}

func addIbcMarkets(markets v016pricefeed.Markets) v016pricefeed.Markets {
	var oracles []types.AccAddress

	if len(markets) > 0 {
		oracles = markets[0].Oracles
	}

	for _, newMarket := range NewIBCMarkets {
		// newMarket is a copy, should not affect other uses of NewIBCMarkets
		newMarket.Oracles = oracles
		markets = append(markets, newMarket)
	}

	return markets
}

func migratePostedPrices(oldPostedPrices v015pricefeed.PostedPrices) v016pricefeed.PostedPrices {
	newPrices := make(v016pricefeed.PostedPrices, len(oldPostedPrices))
	for i, price := range oldPostedPrices {
		newPrices[i] = v016pricefeed.PostedPrice{
			MarketID:      price.MarketID,
			OracleAddress: price.OracleAddress,
			Price:         price.Price,
			Expiry:        price.Expiry,
		}
	}
	return newPrices
}

// Migrate converts v0.15 pricefeed state and returns it in v0.16 format
func Migrate(oldState v015pricefeed.GenesisState) *v016pricefeed.GenesisState {
	return &v016pricefeed.GenesisState{
		Params:       migrateParams(oldState.Params),
		PostedPrices: migratePostedPrices(oldState.PostedPrices),
	}
}
