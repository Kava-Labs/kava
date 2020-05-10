# Begin Block

At the start of each block, expired proposals are deleted. The logic is as follows:

```go
// BeginBlocker runs at the start of every block.
func BeginBlocker(ctx sdk.Context, _ abci.RequestBeginBlock, k Keeper) {
  k.CloseExpiredProposals(ctx)
}
```
