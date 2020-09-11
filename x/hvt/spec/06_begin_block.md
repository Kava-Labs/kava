<!--
order: 6
-->

# Begin Block

At the start of each block, hard tokens are distributed (as claims) to liquidity providers and delegators, respectively.

```go
// BeginBlocker applies rewards to liquidity providers and delegators according to params
func BeginBlocker(ctx sdk.Context, k Keeper) {
  k.ApplyDepositRewards(ctx)
  if k.ShouldDistributeValidatorRewards(ctx, k.BondDenom(ctx)) {
    k.ApplyDelegationRewards(ctx, k.BondDenom(ctx))
    k.SetPreviousDelegationDistribution(ctx, ctx.BlockTime(), k.BondDenom(ctx))
  }
  k.SetPreviousBlockTime(ctx, ctx.BlockTime())
}
```
