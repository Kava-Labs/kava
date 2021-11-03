package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewGenesisState returns a new genesis state
func NewGenesisState(params Params, cdps CDPs, deposits Deposits, startingCdpID uint64,
	debtDenom, govDenom string, prevAccumTimes GenesisAccumulationTimes,
	totalPrincipals GenesisTotalPrincipals) GenesisState {
	return GenesisState{
		Params:                    params,
		CDPs:                      cdps,
		Deposits:                  deposits,
		StartingCdpID:             startingCdpID,
		DebtDenom:                 debtDenom,
		GovDenom:                  govDenom,
		PreviousAccumulationTimes: prevAccumTimes,
		TotalPrincipals:           totalPrincipals,
	}
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() GenesisState {
	return NewGenesisState(
		DefaultParams(),
		CDPs{},
		Deposits{},
		DefaultCdpStartingID,
		DefaultDebtDenom,
		DefaultGovDenom,
		GenesisAccumulationTimes{},
		GenesisTotalPrincipals{},
	)
}

// Validate performs basic validation of genesis data returning an
// error for any failed validation criteria.
func (gs GenesisState) Validate() error {

	if err := gs.Params.Validate(); err != nil {
		return err
	}

	if err := gs.CDPs.Validate(); err != nil {
		return err
	}

	if err := gs.Deposits.Validate(); err != nil {
		return err
	}

	if err := gs.PreviousAccumulationTimes.Validate(); err != nil {
		return err
	}

	if err := gs.TotalPrincipals.Validate(); err != nil {
		return err
	}

	if err := sdk.ValidateDenom(gs.DebtDenom); err != nil {
		return fmt.Errorf(fmt.Sprintf("debt denom invalid: %v", err))
	}

	if err := sdk.ValidateDenom(gs.GovDenom); err != nil {
		return fmt.Errorf(fmt.Sprintf("gov denom invalid: %v", err))
	}

	return nil
}

func validateSavingsRateDistributed(i interface{}) error {
	savingsRateDist, ok := i.(sdk.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if savingsRateDist.IsNegative() {
		return fmt.Errorf("savings rate distributed should not be negative: %s", savingsRateDist)
	}

	return nil
}

// NewGenesisTotalPrincipal returns a new GenesisTotalPrincipal
func NewGenesisTotalPrincipal(ctype string, principal sdk.Int) GenesisTotalPrincipal {
	return GenesisTotalPrincipal{
		CollateralType: ctype,
		TotalPrincipal: principal,
	}
}

// GenesisTotalPrincipals slice of GenesisTotalPrincipal
type GenesisTotalPrincipals []GenesisTotalPrincipal

// Validate performs validation of GenesisTotalPrincipal
func (gtp GenesisTotalPrincipal) Validate() error {
	if gtp.TotalPrincipal.IsNegative() {
		return fmt.Errorf("total principal should be positive, is %s for %s", gtp.TotalPrincipal, gtp.CollateralType)
	}
	return nil
}

// Validate performs validation of GenesisTotalPrincipals
func (gtps GenesisTotalPrincipals) Validate() error {
	for _, gtp := range gtps {
		if err := gtp.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// NewGenesisAccumulationTime returns a new GenesisAccumulationTime
func NewGenesisAccumulationTime(ctype string, prevTime time.Time, factor sdk.Dec) GenesisAccumulationTime {
	return GenesisAccumulationTime{
		CollateralType:           ctype,
		PreviousAccumulationTime: prevTime,
		InterestFactor:           factor,
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
	if gat.InterestFactor.LT(sdk.OneDec()) {
		return fmt.Errorf("interest factor should be ≥ 1.0, is %s for %s", gat.InterestFactor, gat.CollateralType)
	}
	return nil
}
