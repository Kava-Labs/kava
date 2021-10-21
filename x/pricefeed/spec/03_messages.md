<!--
order: 3
-->

# Messages

## Posting Prices

An authorized oraclef for a particular market can post the current price for that market using the `MsgPostPrice` type.

```go
// MsgPostPrice struct representing a posted price message.
// Used by oracles to input prices to the pricefeed
type MsgPostPrice struct {
	From     sdk.AccAddress `json:"from" yaml:"from"`           // client that sent in this address
	MarketID string         `json:"market_id" yaml:"market_id"` // asset code used by exchanges/api
	Price    sdk.Dec        `json:"price" yaml:"price"`         // price in decimal (max precision 18)
	Expiry   time.Time      `json:"expiry" yaml:"expiry"`       // expiry time
}
```

### State Modifications

* Update the raw price for the oracle for this market. This replaces any previous price for that oracle.
