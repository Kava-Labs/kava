<!--
order: 2
-->

# State

## Parameters and Genesis State

`Parameters` define the collateral types which are eligible for rewards, the rate at which rewards are given to users, and the amount of time rewards must vest before users can transfer them.

```go
// Params governance parameters for the incentive module
type Params struct {
  Active  bool    `json:"active" yaml:"active"` // top level governance switch to disable all rewards
  Rewards Rewards `json:"rewards" yaml:"rewards"`
}

// Reward stores the specified state for a single reward period.
type Reward struct {
  Active           bool          `json:"active" yaml:"active"`                       // governance switch to disable a period
  CollateralType   string        `json:"collateral_type" yaml:"collateral_type"`     // the collateral type rewards apply to, must be found in the cdp collaterals
  AvailableRewards sdk.Coin      `json:"available_rewards" yaml:"available_rewards"` // the total amount of coins distributed per period
  Duration         time.Duration `json:"duration" yaml:"duration"`                   // the duration of the period
  TimeLock         time.Duration `json:"time_lock" yaml:"time_lock"`                 // how long rewards for this period are timelocked
  ClaimDuration    time.Duration `json:"claim_duration" yaml:"claim_duration"`       // how long users have after the period ends to claim their rewards
}
```

`GenesisState` defines the state that must be persisted when the blockchain stops/restarts in order for normal function of the incentive module to resume.

```go
// GenesisState is the state that must be provided at genesis.
type GenesisState struct {
  Params             Params                `json:"params" yaml:"params"`
  PreviousBlockTime  time.Time             `json:"previous_block_time" yaml:"previous_block_time"`
  RewardPeriods      RewardPeriods         `json:"reward_periods" yaml:"reward_periods"`
  ClaimPeriods       ClaimPeriods          `json:"claim_periods" yaml:"claim_periods"`
  Claims             Claims                `json:"claims" yaml:"claims"`
  NextClaimPeriodIDs GenesisClaimPeriodIDs `json:"next_claim_period_ids" yaml:"next_claim_period_ids"`
}
```

## Store

For complete details for how items are stored, see [keys.go](../types/keys.go).

### Reward Period Creation

At genesis, or when a collateral is added to rewards, a `RewardPeriod` is created in the store by adding to the existing array of `[]RewardPeriod`. If the previous period for that collateral expired, it is deleted. This implies that, for each collateral, there will only ever be one reward period.

### Reward Period Deletion

When a `RewardPeriod` expires, a new `ClaimPeriod` is created in the store with the next sequential ID for that collateral (ie, if the previous claim period was ID 1, the next one will be ID 2) and the current `RewardPeriod` is deleted from the array of `[]RewardPeriod`.

### Reward Claim Creation

Every block, CDPs are iterated over and the collateral denom is checked for rewards eligibility. For eligible CDPs, a `Claim` is created in the store for all CDP owners, if one doesn't already exist. The claim object is associated with a `ClaimPeriod` via the ID. This implies that a `Claim` is created before `ClaimPeriod` are created. Therefore, a user who submits a `MsgClaimReward` will only be paid out IF 1) they have one or more active `Claim` objects, and 2) the `ClaimPeriod` with the associated ID for that object exists AND the current block time is between the start time and end time for that `ClaimPeriod`.

### Reward Claim Deletion

For claimed rewards, the `Claim` is deleted from the store by deleting the key associated with that denom, ID, and owner. Unclaimed rewards are handled as follows: Each block, the `ClaimPeriod` objects for each denom are iterated over and checked for expiry. If expired, all `Claim` objects for that ID are deleted, as well as the `ClaimPeriod` object. Since claim periods are monotonically increasing, once a non-expired claim period is reached, the iteration can be stopped.
