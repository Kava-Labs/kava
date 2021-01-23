package types

import (
	"bytes"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	tmtime "github.com/tendermint/tendermint/types/time"
)

// GenesisState default values
var (
	DefaultPreviousBlockTime = tmtime.Canonical(time.Unix(0, 0))
)

// GenesisState is the state that must be provided at genesis.
type GenesisState struct {
	Params                    Params                   `json:"params" yaml:"params"`
	PreviousAccumulationTimes GenesisAccumulationTimes `json:"previous_accumulation_times" yaml:"previous_accumulation_times"`
	Deposits                  Deposits                 `json:"deposits" yaml:"deposits"`
	Borrows                   Borrows                  `json:"borrows" yaml:"borrows"`
	TotalSupplied             sdk.Coins                `json:"total_supplied" yaml:"total_supplied"`
	TotalBorrowed             sdk.Coins                `json:"total_borrowed" yaml:"total_borrowed"`
	TotalReserves             sdk.Coins                `json:"total_reserves" yaml:"total_reserves"`
}

// NewGenesisState returns a new genesis state
func NewGenesisState(
	params Params, prevAccumulationTimes GenesisAccumulationTimes, deposits Deposits,
	borrows Borrows, totalSupplied, totalBorrowed, totalReserves sdk.Coins) GenesisState {
	return GenesisState{
		Params:                    params,
		PreviousAccumulationTimes: prevAccumulationTimes,
		Deposits:                  deposits,
		Borrows:                   borrows,
		TotalSupplied:             totalSupplied,
		TotalBorrowed:             totalBorrowed,
		TotalReserves:             totalReserves,
	}
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Params:                    DefaultParams(),
		PreviousAccumulationTimes: DefaultAccumulationTimes,
		Deposits:                  DefaultDeposits,
		Borrows:                   DefaultBorrows,
		TotalSupplied:             DefaultTotalSupplied,
		TotalBorrowed:             DefaultTotalBorrowed,
		TotalReserves:             DefaultTotalReserves,
	}
}

// Validate performs basic validation of genesis data returning an
// error for any failed validation criteria.
func (gs GenesisState) Validate() error {

	if err := gs.Params.Validate(); err != nil {
		return err
	}
	if err := gs.PreviousAccumulationTimes.Validate(); err != nil {
		return err
	}
	if err := gs.Deposits.Validate(); err != nil {
		return err
	}
	if err := gs.Borrows.Validate(); err != nil {
		return err
	}

	if !gs.TotalSupplied.IsValid() {
		return fmt.Errorf("invalid total supplied coins: %s", gs.TotalSupplied)
	}
	if !gs.TotalBorrowed.IsValid() {
		return fmt.Errorf("invalid total borrowed coins: %s", gs.TotalBorrowed)
	}
	if !gs.TotalReserves.IsValid() {
		return fmt.Errorf("invalid total reserves coins: %s", gs.TotalReserves)
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
	SupplyInterestFactor     sdk.Dec   `json:"supply_interest_factor" yaml:"supply_interest_factor"`
	BorrowInterestFactor     sdk.Dec   `json:"borrow_interest_factor" yaml:"borrow_interest_factor"`
}

// NewGenesisAccumulationTime returns a new GenesisAccumulationTime
func NewGenesisAccumulationTime(ctype string, prevTime time.Time, supplyFactor, borrowFactor sdk.Dec) GenesisAccumulationTime {
	return GenesisAccumulationTime{
		CollateralType:           ctype,
		PreviousAccumulationTime: prevTime,
		SupplyInterestFactor:     supplyFactor,
		BorrowInterestFactor:     borrowFactor,
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
	if gat.SupplyInterestFactor.LT(sdk.OneDec()) {
		return fmt.Errorf("supply interest factor should be ≥ 1.0, is %s for %s", gat.SupplyInterestFactor, gat.CollateralType)
	}
	if gat.BorrowInterestFactor.LT(sdk.OneDec()) {
		return fmt.Errorf("borrow interest factor should be ≥ 1.0, is %s for %s", gat.BorrowInterestFactor, gat.CollateralType)
	}
	return nil
}
