package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/subspace"
)

// Parameter store key
var (
	// ParamStoreKeyOracles Param store key for oracles
	ParamStoreKeyOracles = []byte("oracles")
	// ParamStoreKeyAssets Param store key for assets
	ParamStoreKeyAssets = []byte("assets")
)

// ParamKeyTable Key declaration for parameters
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable(
		ParamStoreKeyOracles, OracleParams{},
		ParamStoreKeyAssets, AssetParams{},
	)
}

// AssetParams params for assets. Can be altered via governance
type AssetParams struct {
	Assets []Asset `json:"assets,omitempty" yaml:"assets,omitempty"` //  Array containing the assets supported by the pricefeed
}

// NewAssetParams creates a new AssetParams object
func NewAssetParams(assets []Asset) AssetParams {
	return AssetParams{
		Assets: assets,
	}
}

// DefaultAssetParams default params for assets
func DefaultAssetParams() AssetParams {
	return NewAssetParams([]Asset{})
}

// implements fmt.stringer
func (ap AssetParams) String() string {
	var assetListString []string
	for _, asset := range ap.Assets {
		assetListString = append(assetListString, asset.String())
	}
	return strings.TrimSpace(fmt.Sprintf(`Asset Params:
	Assets:        %s\`, strings.Join(assetListString, ", ")))
}

// OracleParams params for assets. Can be altered via governance
type OracleParams struct {
	Oracles []Oracle `json:"oracles,omitempty" yaml:"oracles,omitempty"` //  Array containing the oracles supported by the pricefeed
}

// NewOracleParams creates a new OracleParams object
func NewOracleParams(oracles []Oracle) OracleParams {
	return OracleParams{
		Oracles: oracles,
	}
}

// DefaultOracleParams default params for assets
func DefaultOracleParams() OracleParams {
	return NewOracleParams([]Oracle{})
}

// implements fmt.stringer
func (op OracleParams) String() string {
	var oracleListString []string
	for _, oracle := range op.Oracles {
		oracleListString = append(oracleListString, oracle.String())
	}
	return strings.TrimSpace(fmt.Sprintf(`Oracle Params:
  Oracles:        %s\`, strings.Join(oracleListString, ", ")))
}

// ParamSubspace defines the expected Subspace interface for parameters
type ParamSubspace interface {
	Get(ctx sdk.Context, key []byte, ptr interface{})
	Set(ctx sdk.Context, key []byte, param interface{})
}
