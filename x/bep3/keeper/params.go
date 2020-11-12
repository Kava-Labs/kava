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

// ------------------------------------------
//				Asset
// ------------------------------------------

// GetAsset returns the asset param associated with the input denom
func (k Keeper) GetAsset(ctx sdk.Context, denom string) (types.AssetParam, error) {
	params := k.GetParams(ctx)
	for _, asset := range params.AssetParams {
		if denom == asset.Denom {
			return asset, nil
		}
	}
	return types.AssetParam{}, sdkerrors.Wrap(types.ErrAssetNotSupported, denom)
}

// SetAsset sets an asset in the params
func (k Keeper) SetAsset(ctx sdk.Context, asset types.AssetParam) {
	params := k.GetParams(ctx)
	for i := range params.AssetParams {
		if params.AssetParams[i].Denom == asset.Denom {
			params.AssetParams[i] = asset
		}
	}
	k.SetParams(ctx, params)
}

// GetAssets returns a list containing all supported assets
func (k Keeper) GetAssets(ctx sdk.Context) (types.AssetParams, bool) {
	params := k.GetParams(ctx)
	return params.AssetParams, len(params.AssetParams) > 0
}

// ------------------------------------------
//				Asset-specific getters
// ------------------------------------------

// GetDeputyAddress returns the deputy address for the input denom
func (k Keeper) GetDeputyAddress(ctx sdk.Context, denom string) (sdk.AccAddress, error) {
	asset, err := k.GetAsset(ctx, denom)
	if err != nil {
		return sdk.AccAddress{}, err
	}
	return asset.DeputyAddress, nil
}

// GetFixedFee returns the fixed fee for incoming swaps
func (k Keeper) GetFixedFee(ctx sdk.Context, denom string) (sdk.Int, error) {
	asset, err := k.GetAsset(ctx, denom)
	if err != nil {
		return sdk.Int{}, err
	}
	return asset.FixedFee, nil
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

// GetSupplyLimit returns the supply limit for the input denom
func (k Keeper) GetSupplyLimit(ctx sdk.Context, denom string) (types.SupplyLimit, error) {
	asset, err := k.GetAsset(ctx, denom)
	if err != nil {
		return types.SupplyLimit{}, err
	}
	return asset.SupplyLimit, nil
}

// ------------------------------------------
//				Cross Asset Getters
// ------------------------------------------

// GetAuthorizedAddresses returns a list of addresses that have special authorization within this module, eg all the deputies.
func (k Keeper) GetAuthorizedAddresses(ctx sdk.Context) []sdk.AccAddress {
	assetParams, found := k.GetAssets(ctx)
	if !found {
		// no assets params is a valid genesis state
		return nil
	}
	addresses := []sdk.AccAddress{}
	uniqueAddresses := map[string]bool{}

	for _, ap := range assetParams {
		a := ap.DeputyAddress
		// de-dup addresses
		if _, found := uniqueAddresses[a.String()]; !found {
			addresses = append(addresses, a)
		}
		uniqueAddresses[a.String()] = true
	}
	return addresses
}
