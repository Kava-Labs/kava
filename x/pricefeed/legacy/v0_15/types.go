package v0_15

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName The name that will be used throughout the module
	ModuleName = "pricefeed"
)

// GenesisState - pricefeed state that must be provided at genesis
type GenesisState struct {
	Params       Params       `json:"params" yaml:"params"`
	PostedPrices PostedPrices `json:"posted_prices" yaml:"posted_prices"`
}

// Params params for pricefeed. Can be altered via governance
type Params struct {
	Markets Markets `json:"markets" yaml:"markets"` //  Array containing the markets supported by the pricefeed
}

// Markets array type for oracle
type Markets []Market

// Market an asset in the pricefeed
type Market struct {
	MarketID   string           `json:"market_id" yaml:"market_id"`
	BaseAsset  string           `json:"base_asset" yaml:"base_asset"`
	QuoteAsset string           `json:"quote_asset" yaml:"quote_asset"`
	Oracles    []sdk.AccAddress `json:"oracles" yaml:"oracles"`
	Active     bool             `json:"active" yaml:"active"`
}

// PostedPrices type for an array of PostedPrice
type PostedPrices []PostedPrice

// PostedPrice price for market posted by a specific oracle
type PostedPrice struct {
	MarketID      string         `json:"market_id" yaml:"market_id"`
	OracleAddress sdk.AccAddress `json:"oracle_address" yaml:"oracle_address"`
	Price         sdk.Dec        `json:"price" yaml:"price"`
	Expiry        time.Time      `json:"expiry" yaml:"expiry"`
}
