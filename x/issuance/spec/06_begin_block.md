<!--
order: 6
-->

# Begin Block

At the start of each block, coins held by blocked addresses are redeemed

```go
  func BeginBlocker(ctx sdk.Context, k Keeper) {
    err := k.RedeemTokensFromBlockedAddresses(ctx, k)
    if err != nil {
      panic(err)
    }
  }
```
