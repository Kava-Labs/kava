<!--
order: 6
-->

# Begin Block

At the start of each block, new KAVA tokens are minted and distributed

```go
// BeginBlocker mints & distributes new tokens for the previous block.
func BeginBlocker(ctx sdk.Context, k Keeper) {
	if err := k.AccumulateAndMintInflation(ctx); err != nil {
		panic(err)
	}
}
```

`AccumulateAndMintInflation` defines all sources of inflation from yearly APYs set via the parameters.
Those rates are converted to the effective rate of the yearly interest rate assuming it is
compounded once per second, for the number of seconds since the previous mint. See concepts for
more details on calculations & the defined sources of inflation.
