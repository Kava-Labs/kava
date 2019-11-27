package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/pricefeed/types"
)

// GetParams gets params from the store
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	return types.NewParams(k.GetAssetParams(ctx))
}

// SetParams updates params in the store
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramstore.SetParamSet(ctx, &params)
}

// GetAssetParams get asset params from store
func (k Keeper) GetAssetParams(ctx sdk.Context) types.Assets {
	var assets types.Assets
	k.paramstore.Get(ctx, types.KeyAssets, &assets)
	return assets
}

// GetOracles returns the oracles in the pricefeed store
func (k Keeper) GetOracles(ctx sdk.Context, assetCode string) (types.Oracles, error) {

	for _, a := range k.GetAssetParams(ctx) {
		if assetCode == a.AssetCode {
			return a.Oracles, nil
		}
	}
	return types.Oracles{}, fmt.Errorf("asset %s not found", assetCode)
}

// GetOracle returns the oracle from the store or an error if not found
func (k Keeper) GetOracle(ctx sdk.Context, assetCode string, address sdk.AccAddress) (types.Oracle, error) {
	oracles, err := k.GetOracles(ctx, assetCode)
	if err != nil {
		return types.Oracle{}, fmt.Errorf("asset %s not found", assetCode)
	}
	for _, o := range oracles {
		if address.Equals(o.Address) {
			return o, nil
		}
	}
	return types.Oracle{}, fmt.Errorf("oracle %s not found for asset %s", address, assetCode)
}

// GetAsset returns the asset if it is in the pricefeed system
func (k Keeper) GetAsset(ctx sdk.Context, assetCode string) (types.Asset, bool) {
	assets := k.GetAssetParams(ctx)

	for i := range assets {
		if assets[i].AssetCode == assetCode {
			return assets[i], true
		}
	}
	return types.Asset{}, false

}
