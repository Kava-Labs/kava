package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/bep3/types"
)

// IncrementAssetSupply increments an asset's supply by the coin
func (k Keeper) IncrementAssetSupply(ctx sdk.Context, coin sdk.Coin) {
	currSupply, _ := k.GetAssetSupply(ctx, []byte(coin.Denom))
	k.SetAssetSupply(ctx, currSupply.Add(coin), []byte(coin.Denom))
}

// DecrementAssetSupply decrement an asset's supply by the coin
func (k Keeper) DecrementAssetSupply(ctx sdk.Context, coin sdk.Coin) {
	currSupply, _ := k.GetAssetSupply(ctx, []byte(coin.Denom))
	k.SetAssetSupply(ctx, currSupply.Sub(coin), []byte(coin.Denom))
}

// ValidateProposedSupplyIncrease checks if the proposed asset amount increase is within the asset's supply limit
func (k Keeper) ValidateProposedSupplyIncrease(ctx sdk.Context, coin sdk.Coin) sdk.Error {
	currSupply := k.LoadAssetSupply(ctx, coin.Denom)
	asset, _ := k.GetAssetByDenom(ctx, coin.Denom)
	if currSupply.Add(coin).Amount.GT(asset.Limit) {
		return types.ErrAboveAssetSupplyLimit(
			k.codespace, coin.Denom, currSupply.Add(coin).Amount, asset.Limit,
		)
	}
	return nil
}

// ValidateProposedSupplyDecrease checks if the proposed asset amount decrease is greater than 0
func (k Keeper) ValidateProposedSupplyDecrease(ctx sdk.Context, coin sdk.Coin) sdk.Error {
	currSupply := k.LoadAssetSupply(ctx, coin.Denom)
	asset, _ := k.GetAssetByDenom(ctx, coin.Denom)
	if currSupply.Sub(coin).Amount.IsPositive() {
		return types.ErrAboveAssetSupplyLimit(
			k.codespace, coin.Denom, currSupply.Add(coin).Amount, asset.Limit,
		)
	}
	return nil
}

// LoadAssetSupply loads an asset's current supply. If it's the first swap of this asset type, set it to 0.
func (k Keeper) LoadAssetSupply(ctx sdk.Context, denom string) sdk.Coin {
	currSupply, found := k.GetAssetSupply(ctx, []byte(denom))
	if !found {
		initialSupply := sdk.NewInt64Coin(denom, 0)
		k.SetAssetSupply(ctx, initialSupply, []byte(denom))
		return initialSupply
	}
	return currSupply
}

// ValidateLiveAsset checks if an asset is both supported and active
func (k Keeper) ValidateLiveAsset(ctx sdk.Context, coin sdk.Coin) sdk.Error {
	asset, found := k.GetAssetByDenom(ctx, coin.Denom)
	if !found {
		return types.ErrAssetNotSupported(k.codespace, coin.Denom)
	}
	if !asset.Active {
		return types.ErrAssetNotActive(k.codespace, asset.Denom)
	}
	return nil
}
