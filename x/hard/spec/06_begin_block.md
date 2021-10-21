<!--
order: 6
-->

# Begin Block

At the start of each block interest is accumulated

```go
// BeginBlocker updates interest rates
func BeginBlocker(ctx sdk.Context, k Keeper) {
  k.ApplyInterestRateUpdates(ctx)
}
```
