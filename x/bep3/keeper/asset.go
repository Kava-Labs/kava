package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/bep3/types"
)

// IncrementAssetSupply increments an asset's supply by the coin param
func (k Keeper) IncrementAssetSupply(ctx sdk.Context, coin sdk.Coin) sdk.Error {
	err := k.ValidateActiveAsset(ctx, coin)
	if err != nil {
		return err
	}

	err = k.ValidateProposedIncrease(ctx, coin)
	if err != nil {
		return err
	}

	currSupply, _ := k.GetAssetSupply(ctx, []byte(coin.Denom))
	k.SetAssetSupply(ctx, currSupply.Add(coin), []byte(coin.Denom))
	return nil
}

// ValidateProposedIncrease checks if a proposed token increase is within supply limits
func (k Keeper) ValidateProposedIncrease(ctx sdk.Context, coin sdk.Coin) sdk.Error {
	coinID := []byte(coin.Denom)

	if coin.IsZero() {
		return types.ErrAmountTooSmall(k.codespace, coin)
	}

	currSupply, found := k.GetAssetSupply(ctx, coinID)
	if !found {
		return types.ErrAssetSupplyNotSet(k.codespace, coin.Denom)
	}

	asset, _ := k.GetAssetByDenom(ctx, coin.Denom)
	if currSupply.Add(coin).Amount.GT(asset.Limit) {
		return types.ErrAboveAssetSupplyLimit(k.codespace, coin.Denom, currSupply.Amount.Int64(), coin.Amount.Int64(), asset.Limit.Int64())
	}
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
