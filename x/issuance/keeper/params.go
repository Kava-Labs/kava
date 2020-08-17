package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/issuance/types"
)

// GetParams returns the params from the store
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	var p types.Params
	k.paramSubspace.GetParamSet(ctx, &p)
	return p
}

// SetParams sets params on the store
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSubspace.SetParamSet(ctx, &params)
}

// GetAsset returns an asset from the params and a boolean for if it was found
func (k Keeper) GetAsset(ctx sdk.Context, denom string) (types.Asset, bool) {
	params := k.GetParams(ctx)
	for _, asset := range params.Assets {
		if asset.Denom == denom {
			return asset, true
		}
	}
	return types.Asset{}, false
}

// SetAsset sets an asset in the params
func (k Keeper) SetAsset(ctx sdk.Context, asset types.Asset) {
	params := k.GetParams(ctx)
	for i := range params.Assets {
		if params.Assets[i].Denom == asset.Denom {
			params.Assets[i] = asset
		}
	}
	k.SetParams(ctx, params)
}
