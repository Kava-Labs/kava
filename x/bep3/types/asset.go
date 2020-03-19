package types

// AssetSupplyInfo contains information about an asset's supply
type AssetSupplyInfo struct {
	Denom        string `json:"denom"  yaml:"denom"`
	InSwapSupply int64  `json:"in_swap_supply"  yaml:"in_swap_supply"`
	AssetSupply  int64  `json:"asset_supply"  yaml:"asset_supply"`
	AssetLimit   int64  `json:"asset_limit"  yaml:"asset_limit"`
}

// NewAssetSupplyInfo initializes a new AssetSupplyInfo
func NewAssetSupplyInfo(denom string, inSwapSupply, assetSupply, assetLimit int64) AssetSupplyInfo {
	return AssetSupplyInfo{
		Denom:        denom,
		InSwapSupply: inSwapSupply,
		AssetSupply:  assetSupply,
		AssetLimit:   assetLimit,
	}
}
