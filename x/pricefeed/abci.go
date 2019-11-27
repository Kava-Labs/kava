package pricefeed

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// EndBlocker updates the current pricefeed
func EndBlocker(ctx sdk.Context, k Keeper) {
	// Update the current price of each asset.
	for _, a := range k.GetAssetParams(ctx) {
		if a.Active {
			err := k.SetCurrentPrices(ctx, a.AssetCode)
			if err != nil {
				// TODO emit an event that price failed to update
				continue
			}
		}
	}
	return
}
