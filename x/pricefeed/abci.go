package pricefeed

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// EndBlocker updates the current pricefeed
func EndBlocker(ctx sdk.Context, k Keeper) {
	// Update the current price of each asset.
	for _, market := range k.GetMarkets(ctx) {
		if !market.Active {
			continue
		}

		err := k.SetCurrentPrices(ctx, market.MarketID)
		if err != nil {
			// TODO: this should panic
			k.Logger(ctx).Error(fmt.Sprintf("failed to set prices for market id %s: %s", market.MarketID, err.Error()))
		}
	}
}
