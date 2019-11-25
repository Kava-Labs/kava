package types

import (
	"strings"
)

// price Takes an [assetcode] and returns CurrentPrice for that asset
// pricefeed Takes an [assetcode] and returns the raw []PostedPrice for that asset
// assets Returns []Assets in the pricefeed system

const (
	// QueryCurrentPrice command for current price queries
	QueryCurrentPrice = "price"
	// QueryRawPrices command for raw price queries
	QueryRawPrices = "rawprices"
	// QueryAssets command for assets query
	QueryAssets = "assets"
)

// QueryRawPricesResp response to a rawprice query
type QueryRawPricesResp []string

// implement fmt.Stringer
func (n QueryRawPricesResp) String() string {
	return strings.Join(n[:], "\n")
}

// QueryAssetsResp response to a assets query
type QueryAssetsResp []string

// implement fmt.Stringer
func (n QueryAssetsResp) String() string {
	return strings.Join(n[:], "\n")
}

