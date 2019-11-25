package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Asset struct that represents an asset in the pricefeed
type Asset struct {
	AssetCode   string `json:"asset_code"`
	Description string `json:"description"`
}

// implement fmt.Stringer
func (a Asset) String() string {
	return strings.TrimSpace(fmt.Sprintf(`AssetCode: %s
Description: %s`, a.AssetCode, a.Description))
}

// Oracle struct that documents which address an oracle is using
type Oracle struct {
	OracleAddress string `json:"oracle_address"`
}

// implement fmt.Stringer
func (o Oracle) String() string {
	return strings.TrimSpace(fmt.Sprintf(`OracleAddress: %s`, o.OracleAddress))
}

// CurrentPrice struct that contains the metadata of a current price for a particular asset in the pricefeed module.
type CurrentPrice struct {
	AssetCode string  `json:"asset_code"`
	Price     sdk.Dec `json:"price"`
	Expiry    sdk.Int `json:"expiry"`
}

// PostedPrice struct represented a price for an asset posted by a specific oracle
type PostedPrice struct {
	AssetCode     string  `json:"asset_code"`
	OracleAddress string  `json:"oracle_address"`
	Price         sdk.Dec `json:"price"`
	Expiry        sdk.Int `json:"expiry"`
}

// implement fmt.Stringer
func (cp CurrentPrice) String() string {
	return strings.TrimSpace(fmt.Sprintf(`AssetCode: %s
Price: %s
Expiry: %s`, cp.AssetCode, cp.Price, cp.Expiry))
}

// implement fmt.Stringer
func (pp PostedPrice) String() string {
	return strings.TrimSpace(fmt.Sprintf(`AssetCode: %s
OracleAddress: %s
Price: %s
Expiry: %s`, pp.AssetCode, pp.OracleAddress, pp.Price, pp.Expiry))
}

// SortDecs provides the interface needed to sort sdk.Dec slices
type SortDecs []sdk.Dec

func (a SortDecs) Len() int           { return len(a) }
func (a SortDecs) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a SortDecs) Less(i, j int) bool { return a[i].LT(a[j]) }
