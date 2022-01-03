package v0_15

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName The name that will be used throughout the module
	ModuleName = "incentive"
)

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

// Params governance parameters for the incentive module
type Params struct {
	USDXMintingRewardPeriods RewardPeriods       `json:"usdx_minting_reward_periods" yaml:"usdx_minting_reward_periods"`
	HardSupplyRewardPeriods  MultiRewardPeriods  `json:"hard_supply_reward_periods" yaml:"hard_supply_reward_periods"`
	HardBorrowRewardPeriods  MultiRewardPeriods  `json:"hard_borrow_reward_periods" yaml:"hard_borrow_reward_periods"`
	DelegatorRewardPeriods   MultiRewardPeriods  `json:"delegator_reward_periods" yaml:"delegator_reward_periods"`
	SwapRewardPeriods        MultiRewardPeriods  `json:"swap_reward_periods" yaml:"swap_reward_periods"`
	ClaimMultipliers         MultipliersPerDenom `json:"claim_multipliers" yaml:"claim_multipliers"`
	ClaimEnd                 time.Time           `json:"claim_end" yaml:"claim_end"`
}

// RewardPeriods array of RewardPeriod
type RewardPeriods []RewardPeriod

// RewardPeriod stores the state of an ongoing reward
type RewardPeriod struct {
	Active           bool      `json:"active" yaml:"active"`
	CollateralType   string    `json:"collateral_type" yaml:"collateral_type"`
	Start            time.Time `json:"start" yaml:"start"`
	End              time.Time `json:"end" yaml:"end"`
	RewardsPerSecond sdk.Coin  `json:"rewards_per_second" yaml:"rewards_per_second"` // per second reward payouts
}

// GenesisRewardState groups together the global state for a particular reward so it can be exported in genesis.
type GenesisRewardState struct {
	AccumulationTimes  AccumulationTimes  `json:"accumulation_times" yaml:"accumulation_times"`
	MultiRewardIndexes MultiRewardIndexes `json:"multi_reward_indexes" yaml:"multi_reward_indexes"`
}

// AccumulationTimes slice of GenesisAccumulationTime
type AccumulationTimes []AccumulationTime

// AccumulationTime stores the previous reward distribution time and its corresponding collateral type
type AccumulationTime struct {
	CollateralType           string    `json:"collateral_type" yaml:"collateral_type"`
	PreviousAccumulationTime time.Time `json:"previous_accumulation_time" yaml:"previous_accumulation_time"`
}

// MultiRewardIndexes slice of MultiRewardIndex
type MultiRewardIndexes []MultiRewardIndex

// MultiRewardIndex stores reward accumulation information on multiple reward types
type MultiRewardIndex struct {
	CollateralType string        `json:"collateral_type" yaml:"collateral_type"`
	RewardIndexes  RewardIndexes `json:"reward_indexes" yaml:"reward_indexes"`
}

// RewardIndexes slice of RewardIndex
type RewardIndexes []RewardIndex

// RewardIndex stores reward accumulation information
type RewardIndex struct {
	CollateralType string  `json:"collateral_type" yaml:"collateral_type"`
	RewardFactor   sdk.Dec `json:"reward_factor" yaml:"reward_factor"`
}

// USDXMintingClaims slice of USDXMintingClaim
type USDXMintingClaims []USDXMintingClaim

// USDXMintingClaim is for USDX minting rewards
type USDXMintingClaim struct {
	BaseClaim     `json:"base_claim" yaml:"base_claim"`
	RewardIndexes RewardIndexes `json:"reward_indexes" yaml:"reward_indexes"`
}

// BaseClaim is a common type shared by all Claims
type BaseClaim struct {
	Owner  sdk.AccAddress `json:"owner" yaml:"owner"`
	Reward sdk.Coin       `json:"reward" yaml:"reward"`
}

// HardLiquidityProviderClaims slice of HardLiquidityProviderClaim
type HardLiquidityProviderClaims []HardLiquidityProviderClaim

// HardLiquidityProviderClaim stores the hard liquidity provider rewards that can be claimed by owner
type HardLiquidityProviderClaim struct {
	BaseMultiClaim      `json:"base_claim" yaml:"base_claim"`
	SupplyRewardIndexes MultiRewardIndexes `json:"supply_reward_indexes" yaml:"supply_reward_indexes"`
	BorrowRewardIndexes MultiRewardIndexes `json:"borrow_reward_indexes" yaml:"borrow_reward_indexes"`
}

// BaseMultiClaim is a common type shared by all Claims with multiple reward denoms
type BaseMultiClaim struct {
	Owner  sdk.AccAddress `json:"owner" yaml:"owner"`
	Reward sdk.Coins      `json:"reward" yaml:"reward"`
}

// DelegatorClaim slice of DelegatorClaim
type DelegatorClaims []DelegatorClaim

// DelegatorClaim stores delegation rewards that can be claimed by owner
type DelegatorClaim struct {
	BaseMultiClaim `json:"base_claim" yaml:"base_claim"`
	RewardIndexes  MultiRewardIndexes `json:"reward_indexes" yaml:"reward_indexes"`
}

// SwapClaims slice of SwapClaim
type SwapClaims []SwapClaim

// SwapClaim stores the swap rewards that can be claimed by owner
type SwapClaim struct {
	BaseMultiClaim `json:"base_claim" yaml:"base_claim"`
	RewardIndexes  MultiRewardIndexes `json:"reward_indexes" yaml:"reward_indexes"`
}

// MultiRewardPeriods array of MultiRewardPeriod
type MultiRewardPeriods []MultiRewardPeriod

// MultiRewardPeriod supports multiple reward types
type MultiRewardPeriod struct {
	Active           bool      `json:"active" yaml:"active"`
	CollateralType   string    `json:"collateral_type" yaml:"collateral_type"`
	Start            time.Time `json:"start" yaml:"start"`
	End              time.Time `json:"end" yaml:"end"`
	RewardsPerSecond sdk.Coins `json:"rewards_per_second" yaml:"rewards_per_second"` // per second reward payouts
}

// MultipliersPerDenom is a map of denoms to a set of multipliers
type MultipliersPerDenom []struct {
	Denom       string      `json:"denom" yaml:"denom"`
	Multipliers Multipliers `json:"multipliers" yaml:"multipliers"`
}

// Multipliers is a slice of Multiplier
type Multipliers []Multiplier

// Multiplier amount the claim rewards get increased by, along with how long the claim rewards are locked
type Multiplier struct {
	Name         MultiplierName `json:"name" yaml:"name"`
	MonthsLockup int64          `json:"months_lockup" yaml:"months_lockup"`
	Factor       sdk.Dec        `json:"factor" yaml:"factor"`
}

// MultiplierName is the user facing ID for a multiplier. There is a restricted set of possible values.
type MultiplierName string

// Available reward multipliers names
const (
	Small  MultiplierName = "small"
	Medium MultiplierName = "medium"
	Large  MultiplierName = "large"
)
