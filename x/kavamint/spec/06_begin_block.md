<!--
order: 6
-->

# Begin Block

At the start of each block, new KAVA tokens are minted and distributed

```go
// BeginBlocker mints & distributes new tokens for the previous block.
func BeginBlocker(ctx sdk.Context, k Keeper) {
  // fetch the last block time from state
	previousBlockTime, found := k.GetPreviousBlockTime(ctx)

	// ------------- Staking Rewards -------------
  // determine amount of the bond denom to mint for staking rewards
	stakingRewardCoins, err := k.AccumulateStakingRewards(ctx, previousBlockTime)
  // mint the staking rewards
  k.MintCoins(ctx, stakingRewardCoins)
  // distribute them to the fee pool for distribution by x/distribution
  k.AddCollectedFees(ctx, stakingRewardCoins)

	// ------------- Community Pool -------------
	communityPoolCoins, err := k.AccumulateCommunityPoolInflation(ctx, previousBlockTime)
  // mint the community pool tokens
  k.MintCoins(ctx, communityPoolCoins)
  // send them to the community module account (the community pool)
  k.AddCommunityPoolFunds(ctx, communityPoolCoins)

  // set block time for next iteration's minting
	k.SetPreviousBlockTime(ctx, ctx.BlockTime())
}
```
