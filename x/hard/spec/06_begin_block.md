<!--
order: 6
-->

# Begin Block

At the start of each block, interest is accumulated, and automated liquidations are attempted

```go
// BeginBlocker updates interest rates and attempts liquidations
func BeginBlocker(ctx sdk.Context, k Keeper) {
  k.ApplyInterestRateUpdates(ctx)
  k.AttemptIndexLiquidations(ctx)
}
```
