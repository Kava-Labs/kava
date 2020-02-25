package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/bep3/types"
)

// IncrementAssetSupply increments an asset's supply by the coin param
func (k Keeper) IncrementAssetSupply(ctx sdk.Context, coin sdk.Coin) sdk.Error {
	// Don't use ValidateActiveAsset so we don't have to refetch asset
	asset, ok := k.GetAssetByDenom(ctx, coin.Denom)
	if !ok {
		return types.ErrAssetNotSupported(k.codespace, coin.Denom)
	}
	if !asset.Active {
		return types.ErrAssetNotActive(k.codespace, asset.Denom)
	}

	coinID := []byte(coin.Denom)

	currSupply, found := k.GetAssetSupply(ctx, coinID)
	if !found {
		k.SetAssetSupply(ctx, coin, coinID)
		return nil
	}

	newSupply := currSupply.Add(coin)
	if newSupply.Amount.Int64() > asset.Limit {
		return types.ErrAboveAssetSupplyLimit(k.codespace, coin.Denom, currSupply.Amount.Int64(), coin.Amount.Int64(), asset.Limit)
	}

	k.SetAssetSupply(ctx, newSupply, coinID)
	return nil
}

// ValidateActiveAsset checks if an asset is both supported and active
func (k Keeper) ValidateActiveAsset(ctx sdk.Context, coin sdk.Coin) sdk.Error {
	asset, found := k.GetAssetByDenom(ctx, coin.Denom)
	if !found {
		return types.ErrAssetNotSupported(k.codespace, coin.Denom)
	}
	if !asset.Active {
		return types.ErrAssetNotActive(k.codespace, asset.Denom)
	}
	return nil
}

func (k Keeper) GetAllAssets(ctx sdk.Context) (assets []sdk.Coin) {
	k.IterateAssetSupplies(ctx, func(asset sdk.Coin) bool {
		assets = append(assets, asset)
		return false
	})
	return
}
