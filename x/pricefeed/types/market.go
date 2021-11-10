package types

import (
	"errors"
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewMarket returns a new Market
func NewMarket(id, base, quote string, oracles []sdk.AccAddress, active bool) Market {
	return Market{
		MarketID:   id,
		BaseAsset:  base,
		QuoteAsset: quote,
		Oracles:    oracles,
		Active:     active,
	}
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
		if seenOracles[oracle.String()] {
			return fmt.Errorf("duplicated oracle %s", oracle)
		}
		seenOracles[oracle.String()] = true
	}
	return nil
}

// ToMarketResponse returns a new MarketResponse from a Market
func (m Market) ToMarketResponse() MarketResponse {
	return NewMarketResponse(m.MarketID, m.BaseAsset, m.QuoteAsset, m.Oracles, m.Active)
}

// Markets is a slice of Market
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

// NewMarketResponse returns a new MarketResponse
func NewMarketResponse(id, base, quote string, oracles []sdk.AccAddress, active bool) MarketResponse {
	var strOracles []string
	for _, oracle := range oracles {
		strOracles = append(strOracles, oracle.String())
	}

	return MarketResponse{
		MarketID:   id,
		BaseAsset:  base,
		QuoteAsset: quote,
		Oracles:    strOracles,
		Active:     active,
	}
}

// MarketResponses is a slice of MarketResponse
type MarketResponses []MarketResponse

// NewCurrentPrice returns an instance of CurrentPrice
func NewCurrentPrice(marketID string, price sdk.Dec) CurrentPrice {
	return CurrentPrice{MarketID: marketID, Price: price}
}

// CurrentPrices is a slice of CurrentPrice
type CurrentPrices []CurrentPrice

// NewCurrentPriceResponse returns an instance of CurrentPriceResponse
func NewCurrentPriceResponse(marketID string, price sdk.Dec) CurrentPriceResponse {
	return CurrentPriceResponse{MarketID: marketID, Price: price}
}

// CurrentPriceResponses is a slice of CurrentPriceResponse
type CurrentPriceResponses []CurrentPriceResponse

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

// PostedPrices is a slice of PostedPrice
type PostedPrices []PostedPrice

// Validate checks if all the posted prices are valid and there are no
// duplicated entries.
func (pps PostedPrices) Validate() error {
	seenPrices := make(map[string]bool)
	for _, pp := range pps {
		if !pp.OracleAddress.Empty() && seenPrices[pp.MarketID+pp.OracleAddress.String()] {
			return fmt.Errorf("duplicated posted price for marked id %s and oracle address %s", pp.MarketID, pp.OracleAddress)
		}

		if err := pp.Validate(); err != nil {
			return err
		}
		seenPrices[pp.MarketID+pp.OracleAddress.String()] = true
	}

	return nil
}

// NewPostedPrice returns a new PostedPrice
func NewPostedPriceResponse(marketID string, oracle sdk.AccAddress, price sdk.Dec, expiry time.Time) PostedPriceResponse {
	return PostedPriceResponse{
		MarketID:      marketID,
		OracleAddress: oracle.String(),
		Price:         price,
		Expiry:        expiry,
	}
}

// PostedPriceResponses is a slice of PostedPriceResponse
type PostedPriceResponses []PostedPriceResponse

// SortDecs provides the interface needed to sort sdk.Dec slices
type SortDecs []sdk.Dec

func (a SortDecs) Len() int           { return len(a) }
func (a SortDecs) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a SortDecs) Less(i, j int) bool { return a[i].LT(a[j]) }
