<!--
order: 2
-->

# State

## Parameters and Genesis State

`Parameters` define the governance parameters and default behavior of each money market. **Money markets should not be removed from params without careful procedures** as it will disable withdraws and liquidations. To deprecate a money market, the following steps should be observed:

1. Borrowing: prevent new borrows by setting param MoneyMarket.BorrowLimit.MaximumLimit to 0.
2. Interest: turn off interest accumulation by setting params MoneyMarket.InterestRateModel.BaseRateAPY and MoneyMarket.InterestRateModel.Kink to 0.
3. Rewards: turn off supply side and/or borrow side rewards by removing any coins in the relevant RewardsPerSecond param in the Incentive module.

Without financial incentives, borrowers and suppliers will withdraw their funds from the money market over time. Once the balances have reached an acceptable level the money market can be deprecated and removed from params, with any additional lingering user funds reimbursed/reallocated as appropriate via a chain upgrade.

```go
// Params governance parameters for hard module
type Params struct {
	MoneyMarkets          MoneyMarkets `json:"money_markets" yaml:"money_markets"`
	MinimumBorrowUSDValue sdk.Dec      `json:"minimum_borrow_usd_value" yaml:"minimum_borrow_usd_value"`
}

// MoneyMarket is a money market for an individual asset
type MoneyMarket struct {
  Denom                  string            `json:"denom" yaml:"denom"` // the denomination of the token for this money market
  BorrowLimit            BorrowLimit       `json:"borrow_limit" yaml:"borrow_limit"` // the borrow limits, if any, applied to this money market
  SpotMarketID           string            `json:"spot_market_id" yaml:"spot_market_id"` // the pricefeed market where price data is fetched
  ConversionFactor       sdk.Int           `json:"conversion_factor" yaml:"conversion_factor"` //the internal conversion factor for going from the smallest unit of a token to a whole unit (ie. 8 for BTC, 6 for KAVA, 18 for ETH)
  InterestRateModel      InterestRateModel `json:"interest_rate_model" yaml:"interest_rate_model"` // the model that determines the prevailing interest rate at each block
  ReserveFactor          sdk.Dec           `json:"reserve_factor" yaml:"reserve_factor"` // the percentage of interest that is accumulated by the protocol as reserves
  KeeperRewardPercentage sdk.Dec           `json:"keeper_reward_percentage" yaml:"keeper_reward_percentages"` // the percentage of a liquidation that is given to the keeper that liquidated the position
}

// MoneyMarkets slice of MoneyMarket
type MoneyMarkets []MoneyMarket

// InterestRateModel contains information about an asset's interest rate
type InterestRateModel struct {
  BaseRateAPY    sdk.Dec `json:"base_rate_apy" yaml:"base_rate_apy"` // the base rate of APY when borrows are zero. Ex. A value of "0.02" would signify an interest rate of 2% APY as the Y-intercept of the interest rate model for the money market. Note that internally, interest rates are stored as per-second interest.
  BaseMultiplier sdk.Dec `json:"base_multiplier" yaml:"base_multiplier"` // the percentage rate at which the interest rate APY increases for each percentage increase in borrow utilization. Ex. A value of "0.01" signifies that the APY interest rate increases by 1% for each additional percentage increase in borrow utilization.
  Kink           sdk.Dec `json:"kink" yaml:"kink"` // the inflection point at which the BaseMultiplier no longer applies and the JumpMultiplier does apply. For example, a value of "0.8" signifies that at 80% utilization, the JumpMultiplier applies
  JumpMultiplier sdk.Dec `json:"jump_multiplier" yaml:"jump_multiplier"` // same as BaseMultiplier, but only applied when utilization is above the Kink
}

// BorrowLimit enforces restrictions on a money market
type BorrowLimit struct {
  HasMaxLimit  bool    `json:"has_max_limit" yaml:"has_max_limit"` // boolean for if the money market has a max amount that can be borrowed, irrespective of utilization.
  MaximumLimit sdk.Dec `json:"maximum_limit" yaml:"maximum_limit"` // the maximum amount that can be borrowed for this money market, irrespective of utilization. Ignored if HasMaxLimit is false
  LoanToValue  sdk.Dec `json:"loan_to_value" yaml:"loan_to_value"` // the percentage amount of borrow power each unit of deposit accounts for. Ex. A value of "0.5" signifies that for $1 of supply of a particular asset, borrow limits will be increased by $0.5
}
```

`GenesisState` defines the state that must be persisted when the blockchain stops/restarts in order for normal function of the hard module to resume and all outstanding funds + interest to be accounted for.

```go
// GenesisState is the state that must be provided at genesis.
type GenesisState struct {
  Params                    Params                   `json:"params" yaml:"params"` // governance parameters
  PreviousAccumulationTimes GenesisAccumulationTimes `json:"previous_accumulation_times"  yaml:"previous_accumulation_times"` // stores the last time interest was calculated for a particular money market
  Deposits                  Deposits                 `json:"deposits" yaml:"deposits"` // stores existing deposits when the chain starts, if any
  Borrows                   Borrows                  `json:"borrows" yaml:"borrows"` // stores existing borrows when the chain starts, if any
  TotalSupplied             sdk.Coins                `json:"total_supplied" yaml:"total_supplied"` // stores the running total of supplied (deposits + interest) coins when the chain starts, if any
  TotalBorrowed             sdk.Coins                `json:"total_borrowed" yaml:"total_borrowed"` // stores the running total of borrowed coins when the chain starts, if any
  TotalReserves             sdk.Coins                `json:"total_reserves" yaml:"total_reserves"` // stores the running total of reserves when the chain starts, if any
}
```
