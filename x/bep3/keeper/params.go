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

// GetAsset returns the asset param associated with the input denom
func (k Keeper) GetAsset(ctx sdk.Context, denom string) (types.AssetParam, error) {
	params := k.GetParams(ctx)
	for _, asset := range params.AssetParams {
		if denom == asset.Denom {
			return asset, nil
		}
	}
	return types.AssetParam{}, types.ErrAssetNotSupported
}

// GetDeputyAddress returns the deputy address for the input denom
func (k Keeper) GetDeputyAddress(ctx sdk.Context, denom string) (sdk.AccAddress, error) {
	asset, err := k.GetAsset(ctx, denom)
	if err != nil {
		return sdk.AccAddress{}, err
	}
	return asset.DeputyAddress, nil
}

// GetIncomingSwapFixedFee returns the fixed fee for incoming swaps
func (k Keeper) GetIncomingSwapFixedFee(ctx sdk.Context, denom string) (sdk.Int, error) {
	asset, err := k.GetAsset(ctx, denom)
	if err != nil {
		return sdk.Int{}, err
	}
	return asset.IncomingSwapFixedFee, nil
}

// GetMinSwapAmount returns the minimum swap amount
func (k Keeper) GetMinSwapAmount(ctx sdk.Context, denom string) (sdk.Int, error) {
	asset, err := k.GetAsset(ctx, denom)
	if err != nil {
		return sdk.Int{}, err
	}
	return asset.MinSwapAmount, nil
}

// GetMaxSwapAmount returns the maximum swap amount
func (k Keeper) GetMaxSwapAmount(ctx sdk.Context, denom string) (sdk.Int, error) {
	asset, err := k.GetAsset(ctx, denom)
	if err != nil {
		return sdk.Int{}, err
	}
	return asset.MaxSwapAmount, nil
}

// GetMinBlockLock returns the minimum block lock
func (k Keeper) GetMinBlockLock(ctx sdk.Context, denom string) (uint64, error) {
	asset, err := k.GetAsset(ctx, denom)
	if err != nil {
		return uint64(0), err
	}
	return asset.MinBlockLock, nil
}

// GetMaxBlockLock returns the maximum block lock
func (k Keeper) GetMaxBlockLock(ctx sdk.Context, denom string) (uint64, error) {
	asset, err := k.GetAsset(ctx, denom)
	if err != nil {
		return uint64(0), err
	}
	return asset.MaxBlockLock, nil
}

// GetAssets returns a list containing all supported assets
func (k Keeper) GetAssets(ctx sdk.Context) (types.AssetParams, bool) {
	params := k.GetParams(ctx)
	return params.AssetParams, len(params.AssetParams) > 0
}

// GetAssetByCoinID returns an asset by its denom
func (k Keeper) GetAssetByCoinID(ctx sdk.Context, coinID int) (types.AssetParam, bool) {
	params := k.GetParams(ctx)
	for _, asset := range params.AssetParams {
		if asset.CoinID == coinID {
			return asset, true
		}
	}
	return types.AssetParam{}, false
}

// ValidateLiveAsset checks if an asset is both supported and active
func (k Keeper) ValidateLiveAsset(ctx sdk.Context, coin sdk.Coin) error {
	asset, err := k.GetAsset(ctx, coin.Denom)
	if err != nil {
		return err
	}
	if !asset.Active {
		return sdkerrors.Wrap(types.ErrAssetNotActive, asset.Denom)
	}
	return nil
}
