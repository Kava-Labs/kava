<!--
order: 5
-->

# Parameters

The hard module has the following parameters:

| Key                            | Type                                  | Example       | Description                                                                 |
| ------------------------------ | ------------------------------------- | ------------- | ----------------------------------------------------------------------------|
| MoneyMarkets                   | array (MoneyMarket)                   | [{see below}] | array of params for each supported market                                   |
| CheckLtvIndexCount             | int                                   | 6             | Number of borrow positions to check for liquidation in each begin blocker   |

Each `MoneyMarket` has the following parameters

| Key                       | Type               | Example                | Description                                                           |
| ------------------------- | ------------------ | ---------------------- | --------------------------------------------------------------------- |
| Denom                     | string             | "bnb"                  | coin denom of the asset which can be deposited and borrowed           |
| BorrowLimit               | BorrowLimit        | [{see below}]          | borrow limits applied to this money market                            |
| SpotMarketID              | string             | "bnb:usd"              | the market id which determines the price of the asset                 |
| ConversionFactor          | Int                | "6"                    | conversion factor for one unit (ie BNB) to the smallest internal unit |
| InterestRateModel         | InterestRateModel  | [{see below}]          | Model which determines the prevailing interest rate per block         |
| ReserveFactor             | Dec                | "0.01"                 | Percentage of interest that is kept as protocol reserves              |
| AuctionSize               | Int                | "1000000000"           | The maximum size of an individual auction                             |
| KeeperRewardPercentage    | Dec                | "0.02"                 | Percentage of deposit rewarded to keeper who liquidates a position    |

Each `BorrowLimit` has the following parameters

| Key                   | Type               | Example                | Description                                                              |
| --------------------- | ------------------ | ---------------------- | ------------------------------------------------------------------------ |
| HasMaxLimit           | bool               | "true"                 | boolean for if a maximum limit is in effect                              |
| MaximumLimit          | Dec                | "10000000.0"           | global maximum amount of coins that can be borrowed                      |
| LoanToValue           | Dec                | "0.5"                  | the percentage amount of borrow power each unit of deposit accounts for  |

Each `InterestRateModel` has the following parameters

| Key              | Type   | Example | Description                                                                                                      |
| ---------------- | ------ | ------- | ---------------------------------------------------------------------------------------------------------------- |
| BaseRateAPY      | Dec    | "0.0"   | the base rate of APY interest when borrows are zero.                                                             |
| BaseMultiplier   | Dec    | "0.01"  | the percentage rate at which the interest rate APY increases for each percentage increase in borrow utilization. |
| Kink             | Dec    | "0.5"   | the inflection point of utilization at which the BaseMultiplier no longer applies and the JumpMultiplier does.   |
| JumpMultiplier   | Dec    | "0.5"   | same as BaseMultiplier, but only applied when utilization is above the Kink                                      |
