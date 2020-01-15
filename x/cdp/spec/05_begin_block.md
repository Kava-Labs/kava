# Begin Blocker

At the start of every block the BeginBlocker of the cdp module:

- updates CDP fees
- liquidates CDPs under the collateral ratio
- nets out system debt and starts auctions to re-balance it
- records the last block time.

## State Changes

### Fees Updated

<!-- TODO -->

### CDP Liquidated

<!-- TODO update after liquidator stuff merge -->

- Get every cdp that is under the liquidation ratio for its collateral type.
- For each cdp
  - set the in-liquidation flag on all the cdp's deposits
  - send collateral from cdp's module account to liquidator's module account, equal to the total stored in the deposits (don't update deposits)
  - send cdp.Principal + cdp.AccumulatedFees amount of internal debt coin from cdp's module account to the liquidator's.
  - decrement total principal by some amount <!-- TODO fees -->

### Net Out System Debt

<!-- TODO update after liquidator stuff merge -->

### Previous Block Time

The current block time is stored as the previous block time.
