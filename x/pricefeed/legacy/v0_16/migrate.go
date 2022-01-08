package v0_16

import (
	v015pricefeed "github.com/kava-labs/kava/x/pricefeed/legacy/v0_15"
	v016pricefeed "github.com/kava-labs/kava/x/pricefeed/types"
)

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
	return v016pricefeed.Params{Markets: markets}
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
