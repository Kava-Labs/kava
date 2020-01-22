# Begin Blocker

At the start of every block the BeginBlocker of the cdp module:

- updates total CDP fees
- liquidates CDPs under the collateral ratio
- nets out system debt and, if necessary, starts auctions to re-balance it
- records the last block time

## Update Fees

- The total fees accumulated since the last block across all CDPs are calculated.
- An equal amount of debt coins are minted and sent to the system's liquidator module account.

## Liquidate CDP

- Get every cdp that is under the liquidation ratio for its collateral type.
- For each cdp:
  - Calculate and update fees since last update.
  - Remove all collateral and internal debt coins from cdp and deposits and delete it. Send coins to liquidator module account.
  - Start auctions of a fixed size from this collateral (with the remainder in a smaller sized auction), sending collateral and debt coins to the auction module account.
  - Decrement total principal.

## Net Out System Debt, Re-Balance

- Burn the maximum possible equal amount of debt and pegged asset from the liquidator module account.
- If there is enough debt remaining for an auction, start one.
- If there is enough surplus stable asset remaining for an auction, start one.
- Otherwise do nothing, leaving debt/surplus to accumulate.

## Update Previous Block Time

The current block time is recorded.
