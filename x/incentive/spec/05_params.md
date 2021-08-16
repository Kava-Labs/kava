<!--
order: 5
-->

# Parameters

The incentive module contains the following parameters:

| Key                      | Type               | Example                | Description                                  |
| ------------------------ | ------------------ | ---------------------- | -------------------------------------------- |
| USDXMintingRewardPeriods | RewardPeriods      | [{see below}]          | USDX minting reward periods                  |
| HardSupplyRewardPeriods  | MultiRewardPeriods | [{see below}]          | Hard supply reward periods                   |
| HardBorrowRewardPeriods  | MultiRewardPeriods | [{see below}]          | Hard borrow reward periods                   |
| DelegatorRewardPeriods   | MultiRewardPeriods | [{see below}]          | Delegator reward periods                     |
| SwapRewardPeriods        | MultiRewardPeriods | [{see below}]          | Swap reward periods                          |
| ClaimMultipliers         | Multipliers        | [{see below}]          | Multipliers applied when rewards are claimed |
| ClaimMultipliers         | Time               | "2025-12-02T14:00:00Z" | Time when reward claiming ends               |

Each `RewardPeriod` has the following parameters

| Key              | Type          | Example                            | Description                                           |
| ---------------- | ------------- | ---------------------------------- | ----------------------------------------------------- |
| Active           | bool          | "true                              | boolean for if rewards for this collateral are active |
| CollateralType   | string        | "bnb-a"                            | the collateral for which rewards are eligible         |
| Start            | Time          | "2020-12-02T14:00:00Z"             | the time at which rewards start                       |
| End              | Time          | "2023-12-02T14:00:00Z"             | the time at which rewards end                         |
| AvailableRewards | object (coin) | `{"denom":"hard","amount":"1000"}` | the rewards available per reward period               |

Each `MultiRewardPeriod` has the following parameters

| Key              | Type          | Example                                                                 | Description                                           |
| ---------------- | ------------- | ----------------------------------------------------------------------- | ----------------------------------------------------- |
| Active           | bool          | "true                                                                   | boolean for if rewards for this collateral are active |
| CollateralType   | string        | "bnb-a"                                                                 | the collateral for which rewards are eligible         |
| Start            | Time          | "2020-12-02T14:00:00Z"                                                  | the time at which rewards start                       |
| End              | Time          | "2023-12-02T14:00:00Z"                                                  | the time at which rewards end                         |
| AvailableRewards | array (coins) | `[{"denom":"hard","amount":"1000"}, {"denom":"ukava","amount":"1000"}]` | the rewards available per reward period               |

Each `Multiplier` has the following parameters:

| Key          | Type   | Example | Description                                                |
| ------------ | ------ | ------- | ---------------------------------------------------------- |
| Name         | string | "large" | the unique name of the reward multiplier                   |
| MonthsLockup | int    | "6"     | number of months tokens with this multiplier are locked    |
| Factor       | Dec    | "0.5"   | the scaling factor for tokens claimed with this multiplier |
