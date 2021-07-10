package types

import (
	"bytes"
	"fmt"
	"time"
)

var (
	DefaultUSDXClaims               = USDXMintingClaims{}
	DefaultHardClaims               = HardLiquidityProviderClaims{}
	DefaultDelegatorClaims          = DelegatorClaims{}
	DefaultSwapClaims               = SwapClaims{}
	DefaultGenesisAccumulationTimes = GenesisAccumulationTimes{}
)

// GenesisState is the state that must be provided at genesis.
type GenesisState struct {
	Params                      Params                      `json:"params" yaml:"params"`
	USDXAccumulationTimes       GenesisAccumulationTimes    `json:"usdx_accumulation_times" yaml:"usdx_accumulation_times"`
	HardSupplyAccumulationTimes GenesisAccumulationTimes    `json:"hard_supply_accumulation_times" yaml:"hard_supply_accumulation_times"`
	HardBorrowAccumulationTimes GenesisAccumulationTimes    `json:"hard_borrow_accumulation_times" yaml:"hard_borrow_accumulation_times"`
	DelegatorAccumulationTimes  GenesisAccumulationTimes    `json:"delegator_accumulation_times" yaml:"delegator_accumulation_times"`
	SwapAccumulationTimes       GenesisAccumulationTimes    `json:"swap_accumulation_times" yaml:"swap_accumulation_times"`
	USDXMintingClaims           USDXMintingClaims           `json:"usdx_minting_claims" yaml:"usdx_minting_claims"`
	HardLiquidityProviderClaims HardLiquidityProviderClaims `json:"hard_liquidity_provider_claims" yaml:"hard_liquidity_provider_claims"`
	DelegatorClaims             DelegatorClaims             `json:"delegator_claims" yaml:"delegator_claims"`
	SwapClaims                  SwapClaims                  `json:"swap_claims" yaml:"swap_claims"`
}

// NewGenesisState returns a new genesis state
func NewGenesisState(params Params, usdxAccumTimes, hardSupplyAccumTimes, hardBorrowAccumTimes, delegatorAccumTimes,
	swapAccumTimes GenesisAccumulationTimes, c USDXMintingClaims, hc HardLiquidityProviderClaims, dc DelegatorClaims, sc SwapClaims) GenesisState {
	return GenesisState{
		Params:                      params,
		USDXAccumulationTimes:       usdxAccumTimes,
		HardSupplyAccumulationTimes: hardSupplyAccumTimes,
		HardBorrowAccumulationTimes: hardBorrowAccumTimes,
		DelegatorAccumulationTimes:  delegatorAccumTimes,
		SwapAccumulationTimes:       swapAccumTimes,
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
		USDXAccumulationTimes:       GenesisAccumulationTimes{},
		HardSupplyAccumulationTimes: GenesisAccumulationTimes{},
		HardBorrowAccumulationTimes: GenesisAccumulationTimes{},
		DelegatorAccumulationTimes:  GenesisAccumulationTimes{},
		SwapAccumulationTimes:       GenesisAccumulationTimes{},
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
	if err := gs.USDXAccumulationTimes.Validate(); err != nil {
		return err
	}
	if err := gs.HardSupplyAccumulationTimes.Validate(); err != nil {
		return err
	}
	if err := gs.HardBorrowAccumulationTimes.Validate(); err != nil {
		return err
	}
	if err := gs.DelegatorAccumulationTimes.Validate(); err != nil {
		return err
	}
	if err := gs.SwapAccumulationTimes.Validate(); err != nil {
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

// Validate performs validation of GenesisAccumulationTime
func (gat GenesisAccumulationTime) Validate() error {
	if len(gat.CollateralType) == 0 {
		return fmt.Errorf("genesis accumulation time's collateral type must be defined")
	}
	return nil
}
