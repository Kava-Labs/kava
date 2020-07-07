<!--
order: 6
-->

# Begin Block

At the start of each block, auctions that have reached `EndTime` are closed. The logic to close auctions is as follows:

```go
var expiredAuctions []uint64
	k.IterateAuctionsByTime(ctx, ctx.BlockTime(), func(id uint64) bool {
		expiredAuctions = append(expiredAuctions, id)
		return false
	})

	for _, id := range expiredAuctions {
		err := k.CloseAuction(ctx, id)
		if err != nil {
			panic(err)
		}
  }
```
