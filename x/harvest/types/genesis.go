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
	DefaultDistributionTimes = GenesisDistributionTimes{}
	DefaultDeposits          = Deposits{}
	DefaultClaims            = Claims{}
)

// GenesisState is the state that must be provided at genesis.
type GenesisState struct {
	Params                    Params                   `json:"params" yaml:"params"`
	PreviousBlockTime         time.Time                `json:"previous_block_time" yaml:"previous_block_time"`
	PreviousDistributionTimes GenesisDistributionTimes `json:"previous_distribution_times" yaml:"previous_distribution_times"`
	Deposits                  Deposits                 `json:"deposits" yaml:"deposits"`
	Claims                    Claims                   `json:"claims" yaml:"claims"`
}

// NewGenesisState returns a new genesis state
func NewGenesisState(params Params, previousBlockTime time.Time, previousDistTimes GenesisDistributionTimes, deposits Deposits, claims Claims) GenesisState {
	return GenesisState{
		Params:                    params,
		PreviousBlockTime:         previousBlockTime,
		PreviousDistributionTimes: previousDistTimes,
		Deposits:                  deposits,
		Claims:                    claims,
	}
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Params:                    DefaultParams(),
		PreviousBlockTime:         DefaultPreviousBlockTime,
		PreviousDistributionTimes: DefaultDistributionTimes,
		Deposits:                  DefaultDeposits,
		Claims:                    DefaultClaims,
	}
}

// Validate performs basic validation of genesis data returning an
// error for any failed validation criteria.
func (gs GenesisState) Validate() error {

	if err := gs.Params.Validate(); err != nil {
		return err
	}
	if gs.PreviousBlockTime.Equal(time.Time{}) {
		return fmt.Errorf("previous block time not set")
	}
	for _, gdt := range gs.PreviousDistributionTimes {
		if gdt.PreviousDistributionTime.Equal(time.Time{}) {
			return fmt.Errorf("previous distribution time not set for %s", gdt.Denom)
		}
		if err := sdk.ValidateDenom(gdt.Denom); err != nil {
			return err
		}
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

// GenesisDistributionTime stores the previous distribution time and its corresponding denom
type GenesisDistributionTime struct {
	Denom                    string    `json:"denom" yaml:"denom"`
	PreviousDistributionTime time.Time `json:"previous_distribution_time" yaml:"previous_distribution_time"`
}

// GenesisDistributionTimes slice of GenesisDistributionTime
type GenesisDistributionTimes []GenesisDistributionTime
