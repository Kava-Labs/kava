<!--
order: 2
-->

# State

## Parameters and Genesis State

`Parameters` define the distribution schedule of hard tokens that will be distributed to delegators and depositors, respectively.

```go
// Params governance parameters for harvest module
type Params struct {
  Active                         bool                           `json:"active" yaml:"active"`
  LiquidityProviderSchedules     DistributionSchedules          `json:"liquidity_provider_schedules" yaml:"liquidity_provider_schedules"`
  DelegatorDistributionSchedules DelegatorDistributionSchedules `json:"delegator_distribution_schedules" yaml:"delegator_distribution_schedules"`
}

// DistributionSchedule distribution schedule for liquidity providers
type DistributionSchedule struct {
  Active           bool        `json:"active" yaml:"active"`
  DepositDenom     string      `json:"deposit_denom" yaml:"deposit_denom"`
  Start            time.Time   `json:"start" yaml:"start"`
  End              time.Time   `json:"end" yaml:"end"`
  RewardsPerSecond sdk.Coin    `json:"rewards_per_second" yaml:"rewards_per_second"`
  ClaimEnd         time.Time   `json:"claim_end" yaml:"claim_end"`
  ClaimMultipliers Multipliers `json:"claim_multipliers" yaml:"claim_multipliers"`
}

// DistributionSchedules slice of DistributionSchedule
type DistributionSchedules []DistributionSchedule

// DelegatorDistributionSchedule distribution schedule for delegators
type DelegatorDistributionSchedule struct {
  DistributionSchedule DistributionSchedule `json:"distribution_schedule" yaml:"distribution_schedule"`

  DistributionFrequency time.Duration `json:"distribution_frequency" yaml:"distribution_frequency"`
}

// DelegatorDistributionSchedules slice of DelegatorDistributionSchedule
type DelegatorDistributionSchedules []DelegatorDistributionSchedule
```

`GenesisState` defines the state that must be persisted when the blockchain stops/restarts in order for normal function of the harvest module to resume.

```go
// GenesisState is the state that must be provided at genesis.
type GenesisState struct {
  Params            Params    `json:"params" yaml:"params"`
  PreviousBlockTime time.Time `json:"previous_block_time" yaml:"previous_block_time"`
}
```
