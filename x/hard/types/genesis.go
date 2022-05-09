package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewGenesisState returns a new genesis state
func NewGenesisState(
	params Params, prevAccumulationTimes GenesisAccumulationTimes, deposits Deposits,
	borrows Borrows, totalSupplied, totalBorrowed, totalReserves sdk.Coins,
) GenesisState {
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
