package types

import (
	"fmt"
	"time"
)

var (
	DefaultUSDXClaims         = USDXMintingClaims{}
	DefaultHardClaims         = HardLiquidityProviderClaims{}
	DefaultDelegatorClaims    = DelegatorClaims{}
	DefaultSwapClaims         = SwapClaims{}
	DefaultSavingsClaims      = SavingsClaims{}
	DefaultGenesisRewardState = NewGenesisRewardState(
		AccumulationTimes{},
		MultiRewardIndexes{},
	)
	DefaultEarnClaims = EarnClaims{}
)

// NewGenesisState returns a new genesis state
func NewGenesisState(
	params Params,
	usdxState, hardSupplyState, hardBorrowState, delegatorState, swapState, savingsState, earnState GenesisRewardState,
	c USDXMintingClaims, hc HardLiquidityProviderClaims, dc DelegatorClaims, sc SwapClaims, savingsc SavingsClaims,
	earnc EarnClaims,
) GenesisState {
	return GenesisState{
		Params: params,

		USDXRewardState:       usdxState,
		HardSupplyRewardState: hardSupplyState,
		HardBorrowRewardState: hardBorrowState,
		DelegatorRewardState:  delegatorState,
		SwapRewardState:       swapState,
		SavingsRewardState:    savingsState,
		EarnRewardState:       earnState,

		USDXMintingClaims:           c,
		HardLiquidityProviderClaims: hc,
		DelegatorClaims:             dc,
		SwapClaims:                  sc,
		SavingsClaims:               savingsc,
		EarnClaims:                  earnc,
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
		SavingsRewardState:          DefaultGenesisRewardState,
		EarnRewardState:             DefaultGenesisRewardState,
		USDXMintingClaims:           DefaultUSDXClaims,
		HardLiquidityProviderClaims: DefaultHardClaims,
		DelegatorClaims:             DefaultDelegatorClaims,
		SwapClaims:                  DefaultSwapClaims,
		SavingsClaims:               DefaultSavingsClaims,
		EarnClaims:                  DefaultEarnClaims,
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
	if err := gs.SavingsRewardState.Validate(); err != nil {
		return err
	}
	if err := gs.EarnRewardState.Validate(); err != nil {
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
	if err := gs.SwapClaims.Validate(); err != nil {
		return err
	}

	if err := gs.SavingsClaims.Validate(); err != nil {
		return err
	}

	return gs.EarnClaims.Validate()
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
