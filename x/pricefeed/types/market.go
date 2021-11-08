package types

import (
	"errors"
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewMarket returns a new Market
func NewMarket(id, base, quote string, oracles []string, active bool) Market {
	return Market{
		MarketID:   id,
		BaseAsset:  base,
		QuoteAsset: quote,
		Oracles:    oracles,
		Active:     active,
	}
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
		if len(oracle) == 0 {
			return fmt.Errorf("oracle %d is empty", i)
		}
		if seenOracles[oracle] {
			return fmt.Errorf("duplicated oracle %s", oracle)
		}
		seenOracles[oracle] = true
	}
	return nil
}

// ValidateMarkets checks if all the markets are valid and there are no
// duplicated entries.
func ValidateMarkets(ms []Market) error {
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

// NewCurrentPrice returns an instance of CurrentPrice
func NewCurrentPrice(marketID string, price sdk.Dec) CurrentPrice {
	return CurrentPrice{MarketID: marketID, Price: price}
}

// NewPostedPrice returns a new PostedPrice
func NewPostedPrice(marketID string, oracle string, price sdk.Dec, expiry time.Time) PostedPrice {
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
	if len(pp.OracleAddress) == 0 {
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

// ValidatePostedPrices checks if all the posted prices are valid and there are
// no duplicated entries.
func ValidatePostedPrices(pps []PostedPrice) error {
	seenPrices := make(map[string]bool)
	for _, pp := range pps {
		if pp.OracleAddress != "" && seenPrices[pp.MarketID+pp.OracleAddress] {
			return fmt.Errorf("duplicated posted price for marked id %s and oracle address %s", pp.MarketID, pp.OracleAddress)
		}

		if err := pp.Validate(); err != nil {
			return err
		}
		seenPrices[pp.MarketID+pp.OracleAddress] = true
	}

	return nil
}

// String implements fmt.Stringer
func (cp CurrentPrice) String() string {
	return strings.TrimSpace(fmt.Sprintf(`Market ID: %s
Price: %s`, cp.MarketID, cp.Price))
}

// String implements fmt.Stringer
func (pp PostedPrice) String() string {
	return strings.TrimSpace(fmt.Sprintf(`Market ID: %s
Oracle Address: %s
Price: %s
Expiry: %s`, pp.MarketID, pp.OracleAddress, pp.Price, pp.Expiry))
}

// SortDecs provides the interface needed to sort sdk.Dec slices
type SortDecs []sdk.Dec

func (a SortDecs) Len() int           { return len(a) }
func (a SortDecs) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a SortDecs) Less(i, j int) bool { return a[i].LT(a[j]) }
