package types

// Asset struct
type Asset struct {
	Symbol           string
	Price            float64
	TargetMarketCode string
}

// CoinGeckoTickers struct
type CoinGeckoTickers struct {
	Name    string            `json:"name"`
	Tickers []CoinGeckoTicker `json:"tickers"`
}

// CoinGeckoTicker struct
type CoinGeckoTicker struct {
	Base   string          `json:"base"`
	Target string          `json:"target"`
	Market CoinGeckoMarket `json:"market"`
	Last   float64         `json:"last"`
	CoinID string          `json:"coin_id"`
}

// CoinGeckoMarket struct
type CoinGeckoMarket struct {
	Name                string `json:"name"`
	Identifier          string `json:"identifier"`
	HasTradingIncentive bool   `json:"has_trading_incentive"`
	CoinID              string `json:"coin_id"`
}

// MarketsRes struct
type MarketsRes struct {
	Height string   `json:"height"`
	Result []string `json:"result"`
}

// TODO: Replace MarketRes once cli query result has been fixed to no longer be a string
// type MarketRes struct {
// 	MarketID   string `json:"market_id" yaml:"market_id"`
// 	BaseAsset  string `json:"base_asset" yaml:"base_asset"`
// 	QuoteAsset string `json:"quote_asset" yaml:"quote_asset"`
// 	Oracles    string `json:"oracles" yaml:"oracles"`
// 	Active     bool   `json:"active" yaml:"active"`
// }
