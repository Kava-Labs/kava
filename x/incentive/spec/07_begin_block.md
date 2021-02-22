<!--
order: 7
-->

# Begin Block

At the start of each block, rewards are accumulated for each reward time. Accumulation refers to computing the total amount of rewards that have accumulated since the previous block and updating a global accumulator value such that whenever a `Claim` object is accessed, it is synchronized with the latest global state. This ensures that all rewards are accurately accounted for without having to iterate over each claim object in the begin blocker

```go
// BeginBlocker runs at the start of every block
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
  params := k.GetParams(ctx)
  for _, rp := range params.USDXMintingRewardPeriods {
    err := k.AccumulateUSDXMintingRewards(ctx, rp)
    if err != nil {
      panic(err)
    }
  }
  for _, rp := range params.HardSupplyRewardPeriods {
    err := k.AccumulateHardSupplyRewards(ctx, rp)
    if err != nil {
      panic(err)
    }
  }
  for _, rp := range params.HardBorrowRewardPeriods {
    err := k.AccumulateHardBorrowRewards(ctx, rp)
    if err != nil {
      panic(err)
    }
  }
  for _, rp := range params.HardDelegatorRewardPeriods {
    err := k.AccumulateHardDelegatorRewards(ctx, rp)
    if err != nil {
      panic(err)
    }
  }
}
```
