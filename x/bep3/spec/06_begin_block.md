# Begin Block

At the start of each block, atomic swaps that have reached `ExpireHeight` are expired. The logic to expire atomic swaps is as follows:

```go
	var expiredSwaps [][]byte
	k.IterateAtomicSwapsByBlock(ctx, uint64(ctx.BlockHeight()), func(id []byte) bool {
		expiredSwaps = append(expiredSwaps, id)
		return false
	})

	// Expire incomplete swaps (claimed swaps have already been removed from byBlock index)
	for _, id := range expiredSwaps {
		atomicSwap, _ := k.GetAtomicSwap(ctx, id)
		atomicSwap.Status = types.Expired
		k.SetAtomicSwap(ctx, atomicSwap)
		k.RemoveFromByBlockIndex(ctx, atomicSwap)
	}
```
