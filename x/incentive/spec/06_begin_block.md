# Begin Block

At the start of each block, expired claims and claim periods are deleted, rewards are applied to CDPs for any ongoing reward periods, expired reward periods are deleted and replaced with a new reward period (if active), and claim periods are created for expiring reward periods. The logic is as follows:

```go
func BeginBlocker(ctx sdk.Context, k Keeper) {
  k.DeleteExpiredClaimsAndClaimPeriods(ctx)
  k.ApplyRewardsToCdps(ctx)
  k.CreateAndDeleteRewardPeriods(ctx)
}
```
