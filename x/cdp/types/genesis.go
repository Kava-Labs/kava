package types

import (
	"bytes"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenesisState is the state that must be provided at genesis.
type GenesisState struct {
	Params                    Params                   `json:"params" yaml:"params"`
	CDPs                      CDPs                     `json:"cdps" yaml:"cdps"`
	Deposits                  Deposits                 `json:"deposits" yaml:"deposits"`
	StartingCdpID             uint64                   `json:"starting_cdp_id" yaml:"starting_cdp_id"`
	DebtDenom                 string                   `json:"debt_denom" yaml:"debt_denom"`
	GovDenom                  string                   `json:"gov_denom" yaml:"gov_denom"`
	PreviousDistributionTime  time.Time                `json:"previous_distribution_time" yaml:"previous_distribution_time"`
	SavingsRateDistributed    sdk.Int                  `json:"savings_rate_distributed" yaml:"savings_rate_distributed"`
	PreviousAccumulationTimes GenesisAccumulationTimes `json:"previous_accumulation_times" yaml:"previous_accumulation_times"`
}

// NewGenesisState returns a new genesis state
func NewGenesisState(params Params, cdps CDPs, deposits Deposits, startingCdpID uint64,
	debtDenom, govDenom string, previousDistTime time.Time, savingsRateDist sdk.Int, prevAccumTimes GenesisAccumulationTimes) GenesisState {
	return GenesisState{
		Params:                    params,
		CDPs:                      cdps,
		Deposits:                  deposits,
		StartingCdpID:             startingCdpID,
		DebtDenom:                 debtDenom,
		GovDenom:                  govDenom,
		PreviousDistributionTime:  previousDistTime,
		SavingsRateDistributed:    savingsRateDist,
		PreviousAccumulationTimes: prevAccumTimes,
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
		DefaultPreviousDistributionTime,
		DefaultSavingsRateDistributed,
		GenesisAccumulationTimes{},
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

	if gs.PreviousDistributionTime.IsZero() {
		return fmt.Errorf("previous distribution time not set")
	}

	if err := validateSavingsRateDistributed(gs.SavingsRateDistributed); err != nil {
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

// GenesisAccumulationTime stores the previous distribution time and its corresponding denom
type GenesisAccumulationTime struct {
	CollateralType           string    `json:"collateral_type" yaml:"collateral_type"`
	PreviousAccumulationTime time.Time `json:"previous_accumulation_time" yaml:"previous_accumulation_time"`
	InterestFactor           sdk.Dec   `json:"interest_factor" yaml:"interest_factor"`
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
