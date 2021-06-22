package types

import (
	"bytes"
	"fmt"
	"time"
)

var (
	DefaultUSDXClaims               = USDXMintingClaims{}
	DefaultHardClaims               = HardLiquidityProviderClaims{}
	DefaultGenesisAccumulationTimes = GenesisAccumulationTimes{}
	DefaultGenesisRewardIndexes     = GenesisRewardIndexesSlice{}
)

// GenesisState is the state that must be provided at genesis.
type GenesisState struct {
	Params                         Params                      `json:"params" yaml:"params"`
	USDXAccumulationTimes          GenesisAccumulationTimes    `json:"usdx_accumulation_times" yaml:"usdx_accumulation_times"`
	USDXRewardIndexes              GenesisRewardIndexesSlice   `json:"usdx_reward_indexes" yaml:"usdx_reward_indexes"`
	HardSupplyAccumulationTimes    GenesisAccumulationTimes    `json:"hard_supply_accumulation_times" yaml:"hard_supply_accumulation_times"`
	HardSupplyRewardIndexes        GenesisRewardIndexesSlice   `json:"hard_supply_reward_indexes" yaml:"hard_supply_reward_indexes"`
	HardBorrowAccumulationTimes    GenesisAccumulationTimes    `json:"hard_borrow_accumulation_times" yaml:"hard_borrow_accumulation_times"`
	HardBorrowRewardIndexes        GenesisRewardIndexesSlice   `json:"hard_borrow_reward_indexes" yaml:"hard_borrow_reward_indexes"`
	HardDelegatorAccumulationTimes GenesisAccumulationTimes    `json:"hard_delegator_accumulation_times" yaml:"hard_delegator_accumulation_times"`
	HardDelegatorRewardIndexes     GenesisRewardIndexesSlice   `json:"hard_delegator_reward_indexes" yaml:"hard_delegator_reward_indexes"`
	USDXMintingClaims              USDXMintingClaims           `json:"usdx_minting_claims" yaml:"usdx_minting_claims"`
	HardLiquidityProviderClaims    HardLiquidityProviderClaims `json:"hard_liquidity_provider_claims" yaml:"hard_liquidity_provider_claims"`
}

// NewGenesisState returns a new genesis state
func NewGenesisState(
	params Params,
	usdxAccumTimes, hardSupplyAccumTimes, hardBorrowAccumTimes, hardDelegatorAccumTimes GenesisAccumulationTimes,
	usdxIndexes, hardSupplyIndexes, hardBorrowIndexes, hardDelegatorIndexes GenesisRewardIndexesSlice,
	c USDXMintingClaims,
	hc HardLiquidityProviderClaims,
) GenesisState {
	return GenesisState{
		Params:                         params,
		USDXAccumulationTimes:          usdxAccumTimes,
		USDXRewardIndexes:              usdxIndexes,
		HardSupplyAccumulationTimes:    hardSupplyAccumTimes,
		HardSupplyRewardIndexes:        hardSupplyIndexes,
		HardBorrowAccumulationTimes:    hardBorrowAccumTimes,
		HardBorrowRewardIndexes:        hardBorrowIndexes,
		HardDelegatorAccumulationTimes: hardDelegatorAccumTimes,
		HardDelegatorRewardIndexes:     hardDelegatorIndexes,
		USDXMintingClaims:              c,
		HardLiquidityProviderClaims:    hc,
	}
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Params:                         DefaultParams(),
		USDXAccumulationTimes:          DefaultGenesisAccumulationTimes,
		USDXRewardIndexes:              DefaultGenesisRewardIndexes,
		HardSupplyAccumulationTimes:    DefaultGenesisAccumulationTimes,
		HardSupplyRewardIndexes:        DefaultGenesisRewardIndexes,
		HardBorrowAccumulationTimes:    DefaultGenesisAccumulationTimes,
		HardBorrowRewardIndexes:        DefaultGenesisRewardIndexes,
		HardDelegatorAccumulationTimes: DefaultGenesisAccumulationTimes,
		HardDelegatorRewardIndexes:     DefaultGenesisRewardIndexes,
		USDXMintingClaims:              DefaultUSDXClaims,
		HardLiquidityProviderClaims:    DefaultHardClaims,
	}
}

// Validate performs basic validation of genesis data returning an
// error for any failed validation criteria.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	if err := gs.USDXAccumulationTimes.Validate(); err != nil {
		return err
	}
	if err := gs.USDXRewardIndexes.Validate(); err != nil {
		return err
	}

	if err := gs.HardSupplyAccumulationTimes.Validate(); err != nil {
		return err
	}
	if err := gs.HardSupplyRewardIndexes.Validate(); err != nil {
		return err
	}

	if err := gs.HardBorrowAccumulationTimes.Validate(); err != nil {
		return err
	}
	if err := gs.HardBorrowRewardIndexes.Validate(); err != nil {
		return err
	}

	if err := gs.HardDelegatorAccumulationTimes.Validate(); err != nil {
		return err
	}
	if err := gs.HardDelegatorRewardIndexes.Validate(); err != nil {
		return err
	}

	if err := gs.HardLiquidityProviderClaims.Validate(); err != nil {
		return err
	}
	if err := gs.USDXMintingClaims.Validate(); err != nil {
		return err
	}
	return nil
}

// Equal checks whether two gov GenesisState structs are equivalent
func (gs GenesisState) Equal(gs2 GenesisState) bool {
	b1 := ModuleCdc.MustMarshalBinaryBare(gs)
	b2 := ModuleCdc.MustMarshalBinaryBare(gs2)
	return bytes.Equal(b1, b2)
}

// IsEmpty returns true if a GenesisState is empty
func (gs GenesisState) IsEmpty() bool {
	return gs.Equal(GenesisState{})
}

// GenesisAccumulationTime stores the previous reward distribution time and its corresponding collateral type
type GenesisAccumulationTime struct {
	CollateralType           string    `json:"collateral_type" yaml:"collateral_type"`
	PreviousAccumulationTime time.Time `json:"previous_accumulation_time" yaml:"previous_accumulation_time"`
}

// NewGenesisAccumulationTime returns a new GenesisAccumulationTime
func NewGenesisAccumulationTime(ctype string, prevTime time.Time) GenesisAccumulationTime {
	return GenesisAccumulationTime{
		CollateralType:           ctype,
		PreviousAccumulationTime: prevTime,
	}
}

// Validate performs validation of GenesisAccumulationTime
func (gat GenesisAccumulationTime) Validate() error {
	if len(gat.CollateralType) == 0 {
		return fmt.Errorf("genesis accumulation time's collateral type must be defined")
	}
	return nil
}

// GenesisAccumulationTimes slice of GenesisAccumulationTime
type GenesisAccumulationTimes []GenesisAccumulationTime

// Validate performs validation of GenesisAccumulationTimes
func (gats GenesisAccumulationTimes) Validate() error {
	for _, gat := range gats {
		if err := gat.Validate(); err != nil {
			return err
		}
	}
	return nil
}

type GenesisRewardIndexes struct {
	CollateralType string        `json:"collateral_type" yaml:"collateral_type"`
	RewardIndexes  RewardIndexes `json:"reward_indexes" yaml:"reward_indexes"`
}

// NewGenesisRewardIndexes returns a new GenesisRewardIndexes
func NewGenesisRewardIndexes(ctype string, indexes RewardIndexes) GenesisRewardIndexes {
	return GenesisRewardIndexes{
		CollateralType: ctype,
		RewardIndexes:  indexes,
	}
}

// Validate performs validation of GenesisAccumulationTime
func (gris GenesisRewardIndexes) Validate() error {
	if len(gris.CollateralType) == 0 {
		return fmt.Errorf("genesis reward indexes's collateral type must be defined")
	}
	if err := gris.RewardIndexes.Validate(); err != nil {
		return fmt.Errorf("invalid reward indexes: %v", err)
	}
	return nil
}

type GenesisRewardIndexesSlice []GenesisRewardIndexes

// Validate performs validation of GenesisAccumulationTimes
func (gris GenesisRewardIndexesSlice) Validate() error {
	for _, gri := range gris {
		if err := gri.Validate(); err != nil {
			return err
		}
	}
	return nil
}
