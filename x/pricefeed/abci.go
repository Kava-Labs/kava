package pricefeed

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// EndBlocker updates the current pricefeed
func EndBlocker(ctx sdk.Context, k Keeper) {
	// Update the current price of each asset.
	for _, a := range k.GetMarkets(ctx) {
		if a.Active {
			err := k.SetCurrentPrices(ctx, a.MarketID)
			if err != nil {
				// In the event of failure, emit an event.
				ctx.EventManager().EmitEvent(
					sdk.NewEvent(
						EventTypeNoValidPrices,
						sdk.NewAttribute(AttributeKeyPriceUpdateFailed, fmt.Sprintf("%s", a.MarketID)),
					),
				)
				continue
			}
		}
	}
	return
}
