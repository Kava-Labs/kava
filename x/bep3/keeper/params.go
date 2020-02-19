package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/bep3/types"
)

// GetParams returns the total set of bep3 parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSubspace.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the bep3 parameters to the param space.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSubspace.SetParamSet(ctx, &params)
}

// GetBnbDeputyAddress returns the Bnbchain's deputy address
func (k Keeper) GetBnbDeputyAddress(ctx sdk.Context) sdk.AccAddress {
	params := k.GetParams(ctx)
	return params.BnbDeputyAddress
}

// GetMaxLockTime returns the maximum lock time
func (k Keeper) GetMaxLockTime(ctx sdk.Context) int64 {
	params := k.GetParams(ctx)
	return params.MaxLockTime
}

// GetMinLockTime returns the minimum lock time
func (k Keeper) GetMinLockTime(ctx sdk.Context) int64 {
	params := k.GetParams(ctx)
	return params.MinLockTime
}

// GetAssets returns a list containing all supported assets
func (k Keeper) GetAssets(ctx sdk.Context) (types.AssetParams, bool) {
	params := k.GetParams(ctx)
	return params.SupportedAssets, len(params.SupportedAssets) > 0
}

// GetAssetByDenom returns an asset by its denom
func (k Keeper) GetAssetByDenom(ctx sdk.Context, denom string) (types.AssetParam, bool) {
	params := k.GetParams(ctx)
	for _, asset := range params.SupportedAssets {
		if asset.Denom == denom {
			return asset, true
		}
	}
	return types.AssetParam{}, false
}

// GetAssetByCoinID returns an asset by its denom
func (k Keeper) GetAssetByCoinID(ctx sdk.Context, coinID string) (types.AssetParam, bool) {
	params := k.GetParams(ctx)
	for _, asset := range params.SupportedAssets {
		if asset.CoinID == coinID {
			return asset, true
		}
	}
	return types.AssetParam{}, false
}
