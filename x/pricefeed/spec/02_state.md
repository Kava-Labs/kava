<!--
order: 2
-->

# State

## Parameters and genesis state

`Paramaters` determine which markets are tracked by the pricefeed and which oracles are authorized to post prices for a given market. There is only one active parameter set at any given time. Updates to parameters can be made via on-chain parameter update proposals.

```go
// Params params for pricefeed. Can be altered via governance
type Params struct {
	Markets Markets `json:"markets" yaml:"markets"` //  Array containing the markets supported by the pricefeed
}

// Market an asset in the pricefeed
type Market struct {
	MarketID   string           `json:"market_id" yaml:"market_id"`
	BaseAsset  string           `json:"base_asset" yaml:"base_asset"`
	QuoteAsset string           `json:"quote_asset" yaml:"quote_asset"`
	Oracles    []sdk.AccAddress `json:"oracles" yaml:"oracles"`
	Active     bool             `json:"active" yaml:"active"`
}

type Markets []Market
```

`GenesisState` defines the state that must be persisted when the blockchain stops/stars in order for the normal function of the pricefeed to resume.

```go
// GenesisState - pricefeed state that must be provided at genesis
type GenesisState struct {
	Params       Params        `json:"params" yaml:"params"`
	PostedPrices []PostedPrice `json:"posted_prices" yaml:"posted_prices"`
}

// PostedPrice price for market posted by a specific oracle
type PostedPrice struct {
	MarketID      string         `json:"market_id" yaml:"market_id"`
	OracleAddress sdk.AccAddress `json:"oracle_address" yaml:"oracle_address"`
	Price         sdk.Dec        `json:"price" yaml:"price"`
	Expiry        time.Time      `json:"expiry" yaml:"expiry"`
}

type PostedPrices []PostedPrice
```

