<!--
order: 2
-->

# State

## Parameters and Genesis State

`Parameters` define the types of incentives that are available and the rewards that are available for each incentive.

```go
// Params governance parameters for the incentive module
type Params struct {
	USDXMintingRewardPeriods RewardPeriods      `json:"usdx_minting_reward_periods" yaml:"usdx_minting_reward_periods"`
	HardSupplyRewardPeriods  MultiRewardPeriods `json:"hard_supply_reward_periods" yaml:"hard_supply_reward_periods"`
	HardBorrowRewardPeriods  MultiRewardPeriods `json:"hard_borrow_reward_periods" yaml:"hard_borrow_reward_periods"`
	DelegatorRewardPeriods   MultiRewardPeriods `json:"delegator_reward_periods" yaml:"delegator_reward_periods"`
	SwapRewardPeriods        MultiRewardPeriods `json:"swap_reward_periods" yaml:"swap_reward_periods"`
	ClaimMultipliers         Multipliers        `json:"claim_multipliers" yaml:"claim_multipliers"`
	ClaimEnd                 time.Time          `json:"claim_end" yaml:"claim_end"`
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
	Params Params `json:"params" yaml:"params"`

	USDXRewardState       GenesisRewardState `json:"usdx_reward_state" yaml:"usdx_reward_state"`
	HardSupplyRewardState GenesisRewardState `json:"hard_supply_reward_state" yaml:"hard_supply_reward_state"`
	HardBorrowRewardState GenesisRewardState `json:"hard_borrow_reward_state" yaml:"hard_borrow_reward_state"`
	DelegatorRewardState  GenesisRewardState `json:"delegator_reward_state" yaml:"delegator_reward_state"`
	SwapRewardState       GenesisRewardState `json:"swap_reward_state" yaml:"swap_reward_state"`

	USDXMintingClaims           USDXMintingClaims           `json:"usdx_minting_claims" yaml:"usdx_minting_claims"`
	HardLiquidityProviderClaims HardLiquidityProviderClaims `json:"hard_liquidity_provider_claims" yaml:"hard_liquidity_provider_claims"`
	DelegatorClaims             DelegatorClaims             `json:"delegator_claims" yaml:"delegator_claims"`
	SwapClaims                  SwapClaims                  `json:"swap_claims" yaml:"swap_claims"`
}
```

## Store

For complete details for how items are stored, see [keys.go](../types/keys.go).

### Claim Creation

When users take incentivized actions, the `incentive` module will create or update a `Claim` object in the store, which represents the amount of rewards that the user is eligible to claim. Each `Claim` object contains one or several RewardIndexes, which are used to calculate the amount of rewards a user can claim. There are four defined claim objects:

- `USDXMintingClaim`
- `HardLiquidityProviderClaim`
- `DelegatorClaim`
- `SwapClaim`

```go

// Claim is an interface for handling common claim actions
type Claim interface {
	GetOwner() sdk.AccAddress
	GetReward() sdk.Coin
	GetType() string
}

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

// RewardIndex stores reward accumulation information
type RewardIndex struct {
	CollateralType string  `json:"collateral_type" yaml:"collateral_type"`
	RewardFactor   sdk.Dec `json:"reward_factor" yaml:"reward_factor"`
}

// MultiRewardIndex stores reward accumulation information on multiple reward types
type MultiRewardIndex struct {
	CollateralType string        `json:"collateral_type" yaml:"collateral_type"`
	RewardIndexes  RewardIndexes `json:"reward_indexes" yaml:"reward_indexes"`
}

// USDXMintingClaim is for USDX minting rewards
type USDXMintingClaim struct {
	BaseClaim     `json:"base_claim" yaml:"base_claim"`
	RewardIndexes RewardIndexes `json:"reward_indexes" yaml:"reward_indexes"`
}

// HardLiquidityProviderClaim stores the hard liquidity provider rewards that can be claimed by owner
type HardLiquidityProviderClaim struct {
	BaseMultiClaim      `json:"base_claim" yaml:"base_claim"`
	SupplyRewardIndexes MultiRewardIndexes `json:"supply_reward_indexes" yaml:"supply_reward_indexes"`
	BorrowRewardIndexes MultiRewardIndexes `json:"borrow_reward_indexes" yaml:"borrow_reward_indexes"`
}

// DelegatorClaim stores delegation rewards that can be claimed by owner
type DelegatorClaim struct {
	BaseMultiClaim `json:"base_claim" yaml:"base_claim"`
	RewardIndexes  MultiRewardIndexes `json:"reward_indexes" yaml:"reward_indexes"`
}

// SwapClaim stores the swap rewards that can be claimed by owner
type SwapClaim struct {
	BaseMultiClaim `json:"base_claim" yaml:"base_claim"`
	RewardIndexes  MultiRewardIndexes `json:"reward_indexes" yaml:"reward_indexes"`
}
```
