<!--
order: 6
-->

# Begin Block

At the start of each block, atomic swaps that meet certain criteria are expired or deleted.

```go
func BeginBlocker(ctx sdk.Context, k Keeper) {
	k.UpdateExpiredAtomicSwaps(ctx)
	k.DeleteClosedAtomicSwapsFromLongtermStorage(ctx)
}
```

## Expiration

If an atomic swap's `ExpireHeight` is greater than the current block height, it will be expired. The logic to expire atomic swaps is as follows:

```go
	var expiredSwapIDs []string
	k.IterateAtomicSwapsByBlock(ctx, uint64(ctx.BlockHeight()), func(id []byte) bool {
		atomicSwap, found := k.GetAtomicSwap(ctx, id)
		if !found {
			return false
		}
		// Expire the uncompleted swap and update both indexes
		atomicSwap.Status = types.Expired
		k.RemoveFromByBlockIndex(ctx, atomicSwap)
		k.SetAtomicSwap(ctx, atomicSwap)
		expiredSwapIDs = append(expiredSwapIDs, hex.EncodeToString(atomicSwap.GetSwapID()))
		return false
	})
```

## Deletion

Atomic swaps are deleted 86400 blocks (one week, assuming a block time of 7 seconds) after being completed. The logic to delete atomic swaps is as follows:

```go
k.IterateAtomicSwapsLongtermStorage(ctx, uint64(ctx.BlockHeight()), func(id []byte) bool {
	swap, found := k.GetAtomicSwap(ctx, id)
	if !found {
		return false
	}
	k.RemoveAtomicSwap(ctx, swap.GetSwapID())
	k.RemoveFromLongtermStorage(ctx, swap)
	return false
})
```