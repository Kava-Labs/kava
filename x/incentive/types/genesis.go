package types

import (
	"bytes"
	"errors"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenesisClaimPeriodID stores the next claim id and its corresponding denom
type GenesisClaimPeriodID struct {
	Denom string `json:"denom" yaml:"denom"`
	ID    uint64 `json:"id" yaml:"id"`
}

// Validate performs a basic check of a GenesisClaimPeriodID fields.
func (gcp GenesisClaimPeriodID) Validate() error {
	if gcp.ID == 0 {
		return errors.New("genesis claim period id cannot be 0")
	}
	return sdk.ValidateDenom(gcp.Denom)
}

// GenesisClaimPeriodIDs array of GenesisClaimPeriodID
type GenesisClaimPeriodIDs []GenesisClaimPeriodID

// Validate checks if all the GenesisClaimPeriodIDs are valid and there are no duplicated
// entries.
func (gcps GenesisClaimPeriodIDs) Validate() error {
	seenIDS := make(map[string]bool)
	var key string
	for _, gcp := range gcps {
		key = gcp.Denom + string(gcp.ID)
		if seenIDS[key] {
			return fmt.Errorf("duplicated genesis claim period with id %d and denom %s", gcp.ID, gcp.Denom)
		}

		if err := gcp.Validate(); err != nil {
			return err
		}
		seenIDS[key] = true
	}

	return nil
}

// GenesisState is the state that must be provided at genesis.
type GenesisState struct {
	Params             Params                `json:"params" yaml:"params"`
	PreviousBlockTime  time.Time             `json:"previous_block_time" yaml:"previous_block_time"`
	RewardPeriods      RewardPeriods         `json:"reward_periods" yaml:"reward_periods"`
	ClaimPeriods       ClaimPeriods          `json:"claim_periods" yaml:"claim_periods"`
	Claims             Claims                `json:"claims" yaml:"claims"`
	NextClaimPeriodIDs GenesisClaimPeriodIDs `json:"next_claim_period_ids" yaml:"next_claim_period_ids"`
}

// NewGenesisState returns a new genesis state
func NewGenesisState(params Params, previousBlockTime time.Time, rp RewardPeriods, cp ClaimPeriods, c Claims, ids GenesisClaimPeriodIDs) GenesisState {
	return GenesisState{
		Params:             params,
		PreviousBlockTime:  previousBlockTime,
		RewardPeriods:      rp,
		ClaimPeriods:       cp,
		Claims:             c,
		NextClaimPeriodIDs: ids,
	}
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Params:             DefaultParams(),
		PreviousBlockTime:  DefaultPreviousBlockTime,
		RewardPeriods:      RewardPeriods{},
		ClaimPeriods:       ClaimPeriods{},
		Claims:             Claims{},
		NextClaimPeriodIDs: GenesisClaimPeriodIDs{},
	}
}

// Validate performs basic validation of genesis data returning an
// error for any failed validation criteria.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}
	if gs.PreviousBlockTime.IsZero() {
		return errors.New("previous block time cannot be 0")
	}
	if err := gs.RewardPeriods.Validate(); err != nil {
		return err
	}
	if err := gs.ClaimPeriods.Validate(); err != nil {
		return err
	}
	if err := gs.Claims.Validate(); err != nil {
		return err
	}
	return gs.NextClaimPeriodIDs.Validate()
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
