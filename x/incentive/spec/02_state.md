<!--
order: 2
-->

# State

## Parameters and Genesis State

`Parameters` define the types of incentives that are available and the rewards that are available for each incentive.

```go
// Params governance parameters for the incentive module
type Params struct {
  USDXMintingRewardPeriods   RewardPeriods      `json:"usdx_minting_reward_periods" yaml:"usdx_minting_reward_periods"` // rewards for minting USDX
  HardSupplyRewardPeriods    MultiRewardPeriods `json:"hard_supply_reward_periods" yaml:"hard_supply_reward_periods"` // rewards for hard supply
  HardBorrowRewardPeriods    MultiRewardPeriods `json:"hard_borrow_reward_periods" yaml:"hard_borrow_reward_periods"` // rewards for hard borrow
  HardDelegatorRewardPeriods RewardPeriods      `json:"hard_delegator_reward_periods" yaml:"hard_delegator_reward_periods"` // rewards for kava delegators
  ClaimMultipliers           Multipliers        `json:"claim_multipliers" yaml:"claim_multipliers"` // the available claim multipliers that determine who much rewards are paid out and how long rewards are locked for
  ClaimEnd                   time.Time          `json:"claim_end" yaml:"claim_end"` // the time at which claims expire
}

```

Each `RewardPeriod` defines a particular collateral for which rewards are eligible and the amount of rewards available.

```go

// RewardPeriod stores the state of an ongoing reward
type RewardPeriod struct {
  Active           bool      `json:"active" yaml:"active"` // if the reward is active
  CollateralType   string    `json:"collateral_type" yaml:"collateral_type"` // the collateral type for which rewards apply
  Start            time.Time `json:"start" yaml:"start"` // when the rewards start
  End              time.Time `json:"end" yaml:"end"` // when the rewards end
  RewardsPerSecond sdk.Coin  `json:"rewards_per_second" yaml:"rewards_per_second"` // per second reward payouts
}
```

Each `MultiRewardPeriod` defines a particular collateral for which one or more reward tokens are eligible and the amount of rewards available

```go
// MultiRewardPeriod supports multiple reward types
type MultiRewardPeriod struct {
  Active           bool      `json:"active" yaml:"active"`
  CollateralType   string    `json:"collateral_type" yaml:"collateral_type"`
  Start            time.Time `json:"start" yaml:"start"`
  End              time.Time `json:"end" yaml:"end"`
  RewardsPerSecond sdk.Coins `json:"rewards_per_second" yaml:"rewards_per_second"` // per second reward payouts
}
```

`GenesisState` defines the state that must be persisted when the blockchain stops/restarts in order for normal function of the incentive module to resume.

```go
// GenesisState is the state that must be provided at genesis.
type GenesisState struct {
  Params                         Params                      `json:"params" yaml:"params"` // governance parameters
  USDXAccumulationTimes          GenesisAccumulationTimes    `json:"usdx_accumulation_times" yaml:"usdx_accumulation_times"` // when USDX rewards were last accumulated
  HardSupplyAccumulationTimes    GenesisAccumulationTimes    `json:"hard_supply_accumulation_times" yaml:"hard_supply_accumulation_times"`  // when hard supply rewards were last accumulated
  HardBorrowAccumulationTimes    GenesisAccumulationTimes    `json:"hard_borrow_accumulation_times" yaml:"hard_borrow_accumulation_times"` // when hard borrow rewards were last accumulated
  HardDelegatorAccumulationTimes GenesisAccumulationTimes    `json:"hard_delegator_accumulation_times"  yaml:"hard_delegator_accumulation_times"` // when hard delegator rewards were last accumulated
  USDXMintingClaims              USDXMintingClaims           `json:"usdx_minting_claims" yaml:"usdx_minting_claims"` // USDX minting claims at genesis, if any
  HardLiquidityProviderClaims    HardLiquidityProviderClaims `json:"hard_liquidity_provider_claims" yaml:"hard_liquidity_provider_claims"` // Hard liquidity provider claims at genesis, if any
}
```

## Store

For complete details for how items are stored, see [keys.go](../types/keys.go).

### Claim Creation

When users take incentivized actions, the `incentive` module will create or update a `Claim` object in the store, which represents the amount of rewards that the user is eligible to claim. The two defined claim objects are `USDXMintingClaims` and `HardLiquidityProviderClaims`:

```go

// BaseClaim is a common type shared by all Claims
type BaseClaim struct {
  Owner  sdk.AccAddress `json:"owner" yaml:"owner"`
  Reward sdk.Coin       `json:"reward" yaml:"reward"`
}

// BaseMultiClaim is a common type shared by all Claims with multiple reward denoms
type BaseMultiClaim struct {
  Owner  sdk.AccAddress `json:"owner" yaml:"owner"`
  Reward sdk.Coins      `json:"reward" yaml:"reward"`
}

// USDXMintingClaim is for USDX minting rewards
type USDXMintingClaim struct {
  BaseClaim     `json:"base_claim" yaml:"base_claim"` // Base claim object
  RewardIndexes RewardIndexes `json:"reward_indexes" yaml:"reward_indexes"` // indexes which are used to calculate the amount of rewards a user can claim
}

// HardLiquidityProviderClaim stores the hard liquidity provider rewards that can be claimed by owner
type HardLiquidityProviderClaim struct {
  BaseMultiClaim         `json:"base_claim" yaml:"base_claim"` // base claim object
  SupplyRewardIndexes    MultiRewardIndexes `json:"supply_reward_indexes" yaml:"supply_reward_indexes"` // indexes which are used to calculate the amount of hard supply rewards a user can claim
  BorrowRewardIndexes    MultiRewardIndexes `json:"borrow_reward_indexes" yaml:"borrow_reward_indexes"` // indexes which are used to calculate the amount of hard borrow rewards a user can claim
  DelegatorRewardIndexes RewardIndexes      `json:"delegator_reward_indexes" yaml:"delegator_reward_indexes"` // indexes which are used to calculate the amount of hard delegator rewards a user can claim
}
```
