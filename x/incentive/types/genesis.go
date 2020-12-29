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
	PreviousAccumulationTimes GenesisAccumulationTimes `json:"previous_accumulation_times" yaml:"previous_accumulation_times"`
	Claims                    Claims                   `json:"claims" yaml:"claims"`
}

// NewGenesisState returns a new genesis state
func NewGenesisState(params Params, prevAccumTimes GenesisAccumulationTimes, c Claims) GenesisState {
	return GenesisState{
		Params:                    params,
		PreviousAccumulationTimes: prevAccumTimes,
		Claims:                    c,
	}
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Params:                    DefaultParams(),
		PreviousAccumulationTimes: GenesisAccumulationTimes{},
		Claims:                    Claims{},
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

	return gs.Claims.Validate()
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
	RewardFactor             sdk.Dec   `json:"reward_factor" yaml:"reward_factor"`
}

// NewGenesisAccumulationTime returns a new GenesisAccumulationTime
func NewGenesisAccumulationTime(ctype string, prevTime time.Time, factor sdk.Dec) GenesisAccumulationTime {
	return GenesisAccumulationTime{
		CollateralType:           ctype,
		PreviousAccumulationTime: prevTime,
		RewardFactor:             factor,
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
	if gat.RewardFactor.LT(sdk.ZeroDec()) {
		return fmt.Errorf("reward factor should be ≥ 0.0, is %s for %s", gat.RewardFactor, gat.CollateralType)
	}
	return nil
}
