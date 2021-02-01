package v0_9

import (
	"errors"
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	ModuleName = "pricefeed"
)

// GenesisState - pricefeed state that must be provided at genesis
type GenesisState struct {
	Params       Params       `json:"params" yaml:"params"`
	PostedPrices PostedPrices `json:"posted_prices" yaml:"posted_prices"`
}

// NewGenesisState creates a new genesis state for the pricefeed module
func NewGenesisState(p Params, pp []PostedPrice) GenesisState {
	return GenesisState{
		Params:       p,
		PostedPrices: pp,
	}
}

type Params struct {
	Markets Markets `json:"markets" yaml:"markets"` //  Array containing the markets supported by the pricefeed
}

type Market struct {
	// TODO: rename to ID
	MarketID   string           `json:"market_id" yaml:"market_id"`
	BaseAsset  string           `json:"base_asset" yaml:"base_asset"`
	QuoteAsset string           `json:"quote_asset" yaml:"quote_asset"`
	Oracles    []sdk.AccAddress `json:"oracles" yaml:"oracles"`
	Active     bool             `json:"active" yaml:"active"`
}

// String implement fmt.Stringer
func (m Market) String() string {
	return fmt.Sprintf(`Asset:
	Market ID: %s
	Base Asset: %s
	Quote Asset: %s
	Oracles: %s
	Active: %t`,
		m.MarketID, m.BaseAsset, m.QuoteAsset, m.Oracles, m.Active)
}

// Validate performs a basic validation of the market params
func (m Market) Validate() error {
	if strings.TrimSpace(m.MarketID) == "" {
		return errors.New("market id cannot be blank")
	}
	if err := sdk.ValidateDenom(m.BaseAsset); err != nil {
		return fmt.Errorf("invalid base asset: %w", err)
	}
	if err := sdk.ValidateDenom(m.QuoteAsset); err != nil {
		return fmt.Errorf("invalid quote asset: %w", err)
	}
	seenOracles := make(map[string]bool)
	for i, oracle := range m.Oracles {
		if oracle.Empty() {
			return fmt.Errorf("oracle %d is empty", i)
		}
		if seenOracles[oracle.String()] {
			return fmt.Errorf("duplicated oracle %s", oracle)
		}
		seenOracles[oracle.String()] = true
	}
	return nil
}

// Markets array type for oracle
type Markets []Market

// Validate checks if all the markets are valid and there are no duplicated
// entries.
func (ms Markets) Validate() error {
	seenMarkets := make(map[string]bool)
	for _, m := range ms {
		if seenMarkets[m.MarketID] {
			return fmt.Errorf("duplicated market %s", m.MarketID)
		}
		if err := m.Validate(); err != nil {
			return err
		}
		seenMarkets[m.MarketID] = true
	}
	return nil
}

// String implements fmt.Stringer
func (ms Markets) String() string {
	out := "Markets:\n"
	for _, m := range ms {
		out += fmt.Sprintf("%s\n", m.String())
	}
	return strings.TrimSpace(out)
}

// CurrentPrice struct that contains the metadata of a current price for a particular market in the pricefeed module.
type CurrentPrice struct {
	MarketID string  `json:"market_id" yaml:"market_id"`
	Price    sdk.Dec `json:"price" yaml:"price"`
}

// NewCurrentPrice returns an instance of CurrentPrice
func NewCurrentPrice(marketID string, price sdk.Dec) CurrentPrice {
	return CurrentPrice{MarketID: marketID, Price: price}
}

// CurrentPrices type for an array of CurrentPrice
type CurrentPrices []CurrentPrice

// PostedPrice price for market posted by a specific oracle
type PostedPrice struct {
	MarketID      string         `json:"market_id" yaml:"market_id"`
	OracleAddress sdk.AccAddress `json:"oracle_address" yaml:"oracle_address"`
	Price         sdk.Dec        `json:"price" yaml:"price"`
	Expiry        time.Time      `json:"expiry" yaml:"expiry"`
}

// NewPostedPrice returns a new PostedPrice
func NewPostedPrice(marketID string, oracle sdk.AccAddress, price sdk.Dec, expiry time.Time) PostedPrice {
	return PostedPrice{
		MarketID:      marketID,
		OracleAddress: oracle,
		Price:         price,
		Expiry:        expiry,
	}
}

// Validate performs a basic check of a PostedPrice params.
func (pp PostedPrice) Validate() error {
	if strings.TrimSpace(pp.MarketID) == "" {
		return errors.New("market id cannot be blank")
	}
	if pp.OracleAddress.Empty() {
		return errors.New("oracle address cannot be empty")
	}
	if pp.Price.IsNegative() {
		return fmt.Errorf("posted price cannot be negative %s", pp.Price)
	}
	if pp.Expiry.Unix() <= 0 {
		return errors.New("expiry time cannot be zero")
	}
	return nil
}

// PostedPrices type for an array of PostedPrice
type PostedPrices []PostedPrice

// Validate checks if all the posted prices are valid and there are no duplicated
// entries.
func (pps PostedPrices) Validate() error {
	seenPrices := make(map[string]bool)
	for _, pp := range pps {
		if pp.OracleAddress != nil && seenPrices[pp.MarketID+pp.OracleAddress.String()] {
			return fmt.Errorf("duplicated posted price for marked id %s and oracle address %s", pp.MarketID, pp.OracleAddress)
		}

		if err := pp.Validate(); err != nil {
			return err
		}
		seenPrices[pp.MarketID+pp.OracleAddress.String()] = true
	}

	return nil
}

// implement fmt.Stringer
func (cp CurrentPrice) String() string {
	return strings.TrimSpace(fmt.Sprintf(`Market ID: %s
Price: %s`, cp.MarketID, cp.Price))
}

// implement fmt.Stringer
func (pp PostedPrice) String() string {
	return strings.TrimSpace(fmt.Sprintf(`Market ID: %s
Oracle Address: %s
Price: %s
Expiry: %s`, pp.MarketID, pp.OracleAddress, pp.Price, pp.Expiry))
}

// String implements fmt.Stringer
func (ps PostedPrices) String() string {
	out := "Posted Prices:\n"
	for _, p := range ps {
		out += fmt.Sprintf("%s\n", p.String())
	}
	return strings.TrimSpace(out)
}
