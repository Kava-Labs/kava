<!--
order: 5
-->

# Parameters

The hard module has the following parameters:

| Key                            | Type                                  | Example       | Description                                  |
| ------------------------------ | ------------------------------------- | ------------- | -------------------------------------------- |
| Active                         | bool                                  | "true"        | boolean for if token distribution is active  |
| LiquidityProviderSchedules     | array (LiquidityProviderSchedule)     | [{see below}] | array of params for each supported asset     |
| DelegatorDistributionSchedules | array (DelegatorDistributionSchedule) | [{see below}] | array of params for staking incentive assets |

Each `LiquidityProviderSchedules` has the following parameters

| Key              | Type               | Example                | Description                                                   |
| ---------------- | ------------------ | ---------------------- | ------------------------------------------------------------- |
| Active           | bool               | "true"                 | boolean for if token distribution is active for this schedule |
| DepositDenom     | string             | "bnb"                  | coin denom of the asset which can be deposited                |
| Start            | time.Time          | "2020-06-01T15:20:00Z" | the time when the period will end                             |
| End              | time.Time          | "2020-06-01T15:20:00Z" | the time when the period will end                             |
| RewardsPerSecond | Coin               | "500hard"              | HARD tokens per second that can be claimed by depositors      |
| ClaimEnd         | time.Time          | "2022-06-01T15:20:00Z" | the time at which users can no longer claim HARD tokens       |
| ClaimMultipliers | array (Multiplier) | [{see below}]          | reward multipliers for users claiming HARD tokens             |

Each `DelegatorDistributionSchedule` has the following parameters

| Key                   | Type               | Example                | Description                                                   |
| --------------------- | ------------------ | ---------------------- | ------------------------------------------------------------- |
| Active                | bool               | "true"                 | boolean for if token distribution is active for this schedule |
| DepositDenom          | string             | "bnb"                  | coin denom of the asset which can be deposited                |
| Start                 | time.Time          | "2020-06-01T15:20:00Z" | the time when the period will end                             |
| End                   | time.Time          | "2020-06-01T15:20:00Z" | the time when the period will end                             |
| RewardsPerSecond      | Coin               | "500hard"              | HARD tokens per second that can be claimed by depositors      |
| ClaimEnd              | time.Time          | "2022-06-01T15:20:00Z" | the time at which users can no longer claim HARD tokens       |
| ClaimMultipliers      | array (Multiplier) | [{see below}]          | reward multipliers for users claiming HARD tokens             |
| DistributionFrequency | time.Duration      | "24hr"                 | frequency at which delegation rewards are accumulated         |

Each `ClaimMultiplier` has the following parameters

| Key          | Type   | Example | Description                                                     |
| ------------ | ------ | ------- | --------------------------------------------------------------- |
| Name         | string | "large" | the unique name of the reward multiplier                        |
| MonthsLockup | int    | "6"     | number of months HARD tokens with this multiplier are locked    |
| Factor       | Dec    | "0.5"   | the scaling factor for HARD tokens claimed with this multiplier |
