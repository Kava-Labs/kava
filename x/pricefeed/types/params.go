package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/subspace"
)

var (
	// KeyAssets store key for assets
	KeyAssets = []byte("assets")
)

// ParamKeyTable Key declaration for parameters
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

// Params params for pricefeed. Can be altered via governance
type Params struct {
	Assets []Asset `json:"assets" yaml:"assets"` //  Array containing the assets supported by the pricefeed
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// pairs of pricefeed module's parameters.
func (p Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		{Key: KeyAssets, Value: &p.Assets},
	}
}

// NewParams creates a new AssetParams object
func NewParams(assets []Asset) Params {
	return Params{
		Assets: assets,
	}
}

// DefaultParams default params for pricefeed
func DefaultParams() Params {
	return NewParams(Assets{})
}

// String implements fmt.stringer
func (p Params) String() string {
	out := "Params:\n"
	for _, a := range p.Assets {
		out += a.String()
	}
	return strings.TrimSpace(out)
}

// ParamSubspace defines the expected Subspace interface for parameters
type ParamSubspace interface {
	Get(ctx sdk.Context, key []byte, ptr interface{})
	Set(ctx sdk.Context, key []byte, param interface{})
}

// Validate ensure that params have valid values
func (p Params) Validate() error {
	// iterate over assets and verify them
	for _, asset := range p.Assets {
		if asset.AssetCode == "" {
			return fmt.Errorf("invalid asset: %s. missing asset code", asset.String())
		}
	}
	return nil
}
