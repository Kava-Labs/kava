# Begin Blocker

At the start of every block the BeginBlocker of the cdp module:

- updates CDP fees
- liquidates CDPs under the collateral ratio
- nets out system debt and if necessary starts auctions to re-balance it
- records the last block time

## Fees Updated

<!-- TODO -->

## CDP Liquidated

- Get every cdp that is under the liquidation ratio for its collateral type.
- For each cdp:
<!-- TODO - update fees -->
  - Remove all collateral and internal debt coins from cdp and deposits and delete it. Send coins to liquidator account.
  - Start auctions of a fixed size from this collateral (with the remainder in a smaller sized auction), sending collateral and debt coins to the auction module account.
  - decrement total principal by some amount <!-- TODO fees -->

## Net Out System Debt, Re-Balance

- Burn an equal amount of debt and stable asset from the liquidator module.
- If there is enough debt remaining for an auction, start one.
- If there is enough surplus stable asset remaining for an auction, start one.

## Previous Block Time

The current block time is stored as the previous block time.
