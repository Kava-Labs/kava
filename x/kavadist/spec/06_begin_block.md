<!--
order: 6
-->

# Begin Block

At the start of each block, the inflationary coins for the ongoing period, if any, are minted. The logic is as follows:

```go
  func BeginBlocker(ctx sdk.Context, k Keeper) {
    err := k.MintPeriodInflation(ctx)
    if err != nil {
      panic(err)
    }
  }
```

## Inflationary Coin Minting

The `MintPeriodInflation` method mints inflationary coins for the two schedules defined in the parameters when `params.Active` is `true`. Coins are minted based off the number of seconds passed since the last block. When `params.Active` is `false`, the method is a no-op.

Firstly, it mints coins at a per second rate derived from `params.Periods`. The coins are minted into `x/kavadist`'s module account.

Next, it mints coins for infrastructure partner rewards at a per second rate derived from `params.InfrastructureParams.InfrastructurePeriods`. The coins are minted to the module account but are then immediately distributed to the infrastructure partners.

## Infrastructure Partner Reward Distribution

The coins minted for the `InfrastructurePeriods` are distributed as follows:
* A distribution is made to each of the infrastructure partners based on the number of seconds since the last distribution for each of the defined `params.InfrastructureParams.PartnerRewards`.
* The remaining coins are distributed to the core infrastructure providers by the weights defined in `params.InfrastructureParams.CoreRewards`.
