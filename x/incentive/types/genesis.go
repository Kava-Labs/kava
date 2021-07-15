package types

import (
	"bytes"
	"fmt"
	"time"
)

var (
	DefaultUSDXClaims         = USDXMintingClaims{}
	DefaultHardClaims         = HardLiquidityProviderClaims{}
	DefaultDelegatorClaims    = DelegatorClaims{}
	DefaultSwapClaims         = SwapClaims{}
	DefaultGenesisRewardState = NewGenesisRewardState(
		AccumulationTimes{},
		MultiRewardIndexes{},
	)
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

// NewGenesisState returns a new genesis state
func NewGenesisState(
	params Params,
	usdxState, hardSupplyState, hardBorrowState, delegatorState, swapState GenesisRewardState,
	c USDXMintingClaims, hc HardLiquidityProviderClaims, dc DelegatorClaims, sc SwapClaims,
) GenesisState {
	return GenesisState{
		Params: params,

		USDXRewardState:       usdxState,
		HardSupplyRewardState: hardSupplyState,
		HardBorrowRewardState: hardBorrowState,
		DelegatorRewardState:  delegatorState,
		SwapRewardState:       swapState,

		USDXMintingClaims:           c,
		HardLiquidityProviderClaims: hc,
		DelegatorClaims:             dc,
		SwapClaims:                  sc,
	}
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Params:                      DefaultParams(),
		USDXRewardState:             DefaultGenesisRewardState,
		HardSupplyRewardState:       DefaultGenesisRewardState,
		HardBorrowRewardState:       DefaultGenesisRewardState,
		DelegatorRewardState:        DefaultGenesisRewardState,
		SwapRewardState:             DefaultGenesisRewardState,
		USDXMintingClaims:           DefaultUSDXClaims,
		HardLiquidityProviderClaims: DefaultHardClaims,
		DelegatorClaims:             DefaultDelegatorClaims,
		SwapClaims:                  DefaultSwapClaims,
	}
}

// Validate performs basic validation of genesis data returning an
// error for any failed validation criteria.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	if err := gs.USDXRewardState.Validate(); err != nil {
		return err
	}
	if err := gs.HardSupplyRewardState.Validate(); err != nil {
		return err
	}
	if err := gs.HardBorrowRewardState.Validate(); err != nil {
		return err
	}
	if err := gs.DelegatorRewardState.Validate(); err != nil {
		return err
	}
	if err := gs.SwapRewardState.Validate(); err != nil {
		return err
	}

	if err := gs.USDXMintingClaims.Validate(); err != nil {
		return err
	}
	if err := gs.HardLiquidityProviderClaims.Validate(); err != nil {
		return err
	}
	if err := gs.DelegatorClaims.Validate(); err != nil {
		return err
	}
	return gs.SwapClaims.Validate()
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

// GenesisRewardState groups together the global state for a particular reward so it can be exported in genesis.
type GenesisRewardState struct {
	AccumulationTimes  AccumulationTimes  `json:"accumulation_times" yaml:"accumulation_times"`
	MultiRewardIndexes MultiRewardIndexes `json:"multi_reward_indexes" yaml:"multi_reward_indexes"`
}

// NewGenesisRewardState returns a new GenesisRewardState
func NewGenesisRewardState(accumTimes AccumulationTimes, indexes MultiRewardIndexes) GenesisRewardState {
	return GenesisRewardState{
		AccumulationTimes:  accumTimes,
		MultiRewardIndexes: indexes,
	}
}

// Validate performs validation of a GenesisRewardState
func (grs GenesisRewardState) Validate() error {
	if err := grs.AccumulationTimes.Validate(); err != nil {
		return err
	}
	return grs.MultiRewardIndexes.Validate()
}

// AccumulationTime stores the previous reward distribution time and its corresponding collateral type
type AccumulationTime struct {
	CollateralType           string    `json:"collateral_type" yaml:"collateral_type"`
	PreviousAccumulationTime time.Time `json:"previous_accumulation_time" yaml:"previous_accumulation_time"`
}

// NewAccumulationTime returns a new GenesisAccumulationTime
func NewAccumulationTime(ctype string, prevTime time.Time) AccumulationTime {
	return AccumulationTime{
		CollateralType:           ctype,
		PreviousAccumulationTime: prevTime,
	}
}

// Validate performs validation of GenesisAccumulationTime
func (gat AccumulationTime) Validate() error {
	if len(gat.CollateralType) == 0 {
		return fmt.Errorf("genesis accumulation time's collateral type must be defined")
	}
	return nil
}

// AccumulationTimes slice of GenesisAccumulationTime
type AccumulationTimes []AccumulationTime

// Validate performs validation of GenesisAccumulationTimes
func (gats AccumulationTimes) Validate() error {
	for _, gat := range gats {
		if err := gat.Validate(); err != nil {
			return err
		}
	}
	return nil
}
