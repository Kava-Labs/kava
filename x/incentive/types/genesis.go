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

	// New fields
	DefaultClaims                 = Claims{}
	DefaultAccrualTimes           = AccrualTimes{}
	DefaultTypedRewardIndexesList = TypedRewardIndexesList{}
)

// NewGenesisState returns a new genesis state
func NewGenesisState(
	params Params,
	usdxState, hardSupplyState, hardBorrowState, delegatorState, swapState, savingsState, earnState GenesisRewardState,
	c Claims,
	uc USDXMintingClaims, hc HardLiquidityProviderClaims, dc DelegatorClaims, sc SwapClaims, savingsc SavingsClaims,
	earnc EarnClaims,
	accrualTimes AccrualTimes,
	rewardIndexes TypedRewardIndexesList,
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

		USDXMintingClaims:           uc,
		HardLiquidityProviderClaims: hc,
		DelegatorClaims:             dc,
		SwapClaims:                  sc,
		SavingsClaims:               savingsc,
		EarnClaims:                  earnc,

		// New fields
		// Claims of all types
		Claims:        c,
		AccrualTimes:  accrualTimes,
		RewardIndexes: rewardIndexes,
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
		Claims:                      DefaultClaims,
		USDXMintingClaims:           DefaultUSDXClaims,
		HardLiquidityProviderClaims: DefaultHardClaims,
		DelegatorClaims:             DefaultDelegatorClaims,
		SwapClaims:                  DefaultSwapClaims,
		SavingsClaims:               DefaultSavingsClaims,
		EarnClaims:                  DefaultEarnClaims,
		AccrualTimes:                DefaultAccrualTimes,
		RewardIndexes:               DefaultTypedRewardIndexesList,
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

	if err := gs.EarnClaims.Validate(); err != nil {
		return err
	}

	// Refactored methods -- these will eventually replace the claim and state methods above
	if err := gs.Claims.Validate(); err != nil {
		return err
	}

	if err := gs.AccrualTimes.Validate(); err != nil {
		return err
	}

	return gs.RewardIndexes.Validate()
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

// NewAccrualTime returns a new AccrualTime
func NewAccrualTime(claimType ClaimType, collateralType string, prevTime time.Time) AccrualTime {
	return AccrualTime{
		ClaimType:                claimType,
		CollateralType:           collateralType,
		PreviousAccumulationTime: prevTime,
	}
}

// Validate performs validation of AccrualTime
func (at AccrualTime) Validate() error {
	if at.PreviousAccumulationTime.IsZero() {
		return fmt.Errorf("previous accumulation time cannot be zero")
	}

	if err := at.ClaimType.Validate(); err != nil {
		return err
	}

	if len(at.CollateralType) == 0 {
		return fmt.Errorf("collateral type cannot be empty")
	}

	return nil
}

// AccrualTimes slice of AccrualTime
type AccrualTimes []AccrualTime

// Validate performs validation of AccrualTimes
func (gats AccrualTimes) Validate() error {
	seenAccrualTimes := make(map[string]bool)

	for _, gat := range gats {
		if err := gat.Validate(); err != nil {
			return err
		}

		key := fmt.Sprintf("%s-%s", gat.ClaimType, gat.CollateralType)
		if seenAccrualTimes[key] {
			return fmt.Errorf("duplicate accrual time found for %s", key)
		}
		seenAccrualTimes[key] = true
	}
	return nil
}
