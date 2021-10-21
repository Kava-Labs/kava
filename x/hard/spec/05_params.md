<!--
order: 5
-->

# Parameters

Example parameters for the Hard module:

| Key                   | Type                | Example       | Description                                  |
| --------------------- | ------------------- | ------------- | -------------------------------------------- |
| MoneyMarkets          | array (MoneyMarket) | [{see below}] | Array of params for each supported market    |
| MinimumBorrowUSDValue | sdk.Dec             | 10.0          | Minimum amount an individual user can borrow |

Example parameters for `MoneyMarket`:

| Key                    | Type              | Example       | Description                                                           |
| ---------------------- | ----------------- | ------------- | --------------------------------------------------------------------- |
| Denom                  | string            | "bnb"         | Coin denom of the asset which can be deposited and borrowed           |
| BorrowLimit            | BorrowLimit       | [{see below}] | Borrow limits applied to this money market                            |
| SpotMarketID           | string            | "bnb:usd"     | The market id which determines the price of the asset                 |
| ConversionFactor       | Int               | "6"           | Conversion factor for one unit (ie BNB) to the smallest internal unit |
| InterestRateModel      | InterestRateModel | [{see below}] | Model which determines the prevailing interest rate per block         |
| ReserveFactor          | Dec               | "0.01"        | Percentage of interest that is kept as protocol reserves              |
| KeeperRewardPercentage | Dec               | "0.02"        | Percentage of deposit rewarded to keeper who liquidates a position    |

Example parameters for `BorrowLimit`:

| Key          | Type | Example      | Description                                                             |
| ------------ | ---- | ------------ | ----------------------------------------------------------------------- |
| HasMaxLimit  | bool | "true"       | Boolean for if a maximum limit is in effect                             |
| MaximumLimit | Dec  | "10000000.0" | Global maximum amount of coins that can be borrowed                     |
| LoanToValue  | Dec  | "0.5"        | The percentage amount of borrow power each unit of deposit accounts for |

Example parameters for `InterestRateModel`:

| Key            | Type | Example | Description                                                                                                     |
| -------------- | ---- | ------- | --------------------------------------------------------------------------------------------------------------- |
| BaseRateAPY    | Dec  | "0.0"   | The base rate of APY interest when borrows are zero                                                             |
| BaseMultiplier | Dec  | "0.01"  | The percentage rate at which the interest rate APY increases for each percentage increase in borrow utilization |
| Kink           | Dec  | "0.5"   | The inflection point of utilization at which the BaseMultiplier no longer applies and the JumpMultiplier does   |
| JumpMultiplier | Dec  | "0.5"   | Same as BaseMultiplier, but only applied when utilization is above the Kink                                     |
