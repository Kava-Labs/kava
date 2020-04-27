package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

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

// GetMaxBlockLock returns the maximum block lock
func (k Keeper) GetMaxBlockLock(ctx sdk.Context) int64 {
	params := k.GetParams(ctx)
	return params.MaxBlockLock
}

// GetMinBlockLock returns the minimum block lock
func (k Keeper) GetMinBlockLock(ctx sdk.Context) int64 {
	params := k.GetParams(ctx)
	return params.MinBlockLock
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
func (k Keeper) GetAssetByCoinID(ctx sdk.Context, coinID int) (types.AssetParam, bool) {
	params := k.GetParams(ctx)
	for _, asset := range params.SupportedAssets {
		if asset.CoinID == coinID {
			return asset, true
		}
	}
	return types.AssetParam{}, false
}

// ValidateLiveAsset checks if an asset is both supported and active
func (k Keeper) ValidateLiveAsset(ctx sdk.Context, coin sdk.Coin) error {
	asset, found := k.GetAssetByDenom(ctx, coin.Denom)
	if !found {
		return sdkerrors.Wrap(types.ErrAssetNotSupported, coin.Denom)
	}
	if !asset.Active {
		return sdkerrors.Wrap(types.ErrAssetNotActive, asset.Denom)
	}
	return nil
}
