<!--
order: 6
-->

# Begin Block

At the start of each block, proposals are processed. Active proposals with "first-past-the-post" vote tallying are evaluated and if they meet quorum and voting threshold requirements are enacted, resulting in the deletion of the proposal and any associated votes. If a "first-past-the-post" proposal doesn't meet quorum and voting threshold requirements by its deadline it is not enacted and is deleted. Proposals with "deadline" vote tallying are evaluated at their deadline before being deleted.

```go
// BeginBlocker runs at the start of every block.
func BeginBlocker(ctx sdk.Context, _ abci.RequestBeginBlock, k Keeper) {
	k.ProcessProposals(ctx)
}
```
