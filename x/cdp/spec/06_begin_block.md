<!--
order: 6
-->

# Begin Block

At the start of every block the BeginBlock of the cdp module:

- updates the status of the pricefeed for each collateral asset
- If the pricefeed is active (reporting a price):
  - updates fees for CDPs
  - liquidates CDPs under the collateral ratio
- nets out system debt and, if necessary, starts auctions to re-balance it
- pays out the savings rate if sufficient time has past
- records the last savings rate distribution, if one occurred

## Update Fees

- The total fees accumulated since the last block for each CDP are calculated.
- If the fee amount is non-zero:
  - Set the updated value for fees
  - Set the fees updated time for the CDP to the current block time
  - An equal amount of debt coins are minted and sent to the system's CDP module account.
  - An equal amount of stable asset coins are minted and sent to the system's liquidator module account
  - Increment total principal.

## Liquidate CDP

- Get every cdp that is under the liquidation ratio for its collateral type.
- For each cdp:
  - Remove all collateral and internal debt coins from cdp and deposits and delete it. Send the coins to the liquidator module account.
  - Start auctions of a fixed size from this collateral (with any remainder in a smaller sized auction), sending collateral and debt coins to the auction module account.
  - Decrement total principal.

## Net Out System Debt, Re-Balance

- Burn the maximum possible equal amount of debt and stable asset from the liquidator module account.
- If there is enough debt remaining for an auction, start one.
- If there is enough surplus stable asset, minus surplus reserved for the savings rate, remaining for an auction, start one.
- Otherwise do nothing, leave debt/surplus to accumulate over subsequent blocks.

## Distribute Surplus Stable Asset According to the Savings Rate

- If `SavingsDistributionFrequency` seconds have elapsed since the previous distribution, the savings rate is applied to all accounts that hold stable asset.
- Each account that holds stable asset is distributed a ratable portion of the surplus that is apportioned to the savings rate.
- If distribution occurred, the time of the distribution is recorded.
