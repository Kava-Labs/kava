package types

// Querier routes for the issuance module
const (
	QueryGetParams = "parameters"
	QueryGetAsset  = "asset"
)

// QueryAssetParams params for querying an asset by denom
type QueryAssetParams struct {
	Denom string `json:"denom" yaml:"denom"`
}
