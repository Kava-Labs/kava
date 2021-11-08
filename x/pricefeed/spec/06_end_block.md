<!--
order: 6
-->

# End Block

At the end of each block, the current price is calculated as the median of all raw prices for each market. The logic is as follows:

```go
// EndBlocker updates the current pricefeed
func EndBlocker(ctx sdk.Context, k Keeper) {
	// Update the current price of each asset.
	for _, market := range k.GetMarkets(ctx) {
		if market.Active {
			err := k.SetCurrentPrices(ctx, market.MarketId)
			if err != nil {
				// In the event of failure, emit an event.
				ctx.EventManager().EmitEvent(
					sdk.NewEvent(
						EventTypeNoValidPrices,
						sdk.NewAttribute(AttributeMarketID, fmt.Sprintf("%s", market.MarketId)),
					),
				)
				continue
			}
		}
	}
	return
}
```
