# Begin Block

At the start of each block, the inflationary coins for the ongoing period, if any, are minted. The logic is as follows:

```go
  func BeginBlocker(ctx sdk.Context, k Keeper) {
    err := k.MintPeriodInflation(ctx)
    if err != nil {
      panic(err)
    }
  }
```
