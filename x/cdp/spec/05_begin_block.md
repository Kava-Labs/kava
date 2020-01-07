# Begin Blocker

At the start of every block the BeginBlocker of the cdp module updates CDPs, fees, and a record of the last block time.

## State Changes

### CDP Liquidated

- Get every cdp that is under the liquidation ratio for its collateral type.
- For each cdp
  - set the in-liquidation flag on all the cdp's deposits
  - send collateral from cdp's module account to liquidator's module account, equal to the total stored in the deposits (don't update deposits)
  - send cdp.Principal + cdp.AccumulatedFees amount of internal debt coin from cdp's module account to the liquidator's.
  - decrement total principal by some amount <!-- TODO fees -->

### Fees Updated

<!-- TODO -->

### Previous Block Time

The current block time is stored as the previous block time.
<!-- TODO should PreviousBlockTime not be the previous BlockTime? -->
