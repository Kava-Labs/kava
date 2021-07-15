package types

import (
	"bytes"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type poolShares struct {
	totalShares      sdk.Int
	totalSharesOwned sdk.Int
}

var (
	// DefaultPoolRecords is used to set default records in default genesis state
	DefaultPoolRecords = PoolRecords{}
	// DefaultShareRecords is used to set default records in default genesis state
	DefaultShareRecords = ShareRecords{}
)

// GenesisState is the state that must be provided at genesis.
type GenesisState struct {
	Params       Params `json:"params" yaml:"params"`
	PoolRecords  `json:"pool_records" yaml:"pool_records"`
	ShareRecords `json:"share_records" yaml:"share_records"`
}

// NewGenesisState creates a new genesis state.
func NewGenesisState(params Params, poolRecords PoolRecords, shareRecords ShareRecords) GenesisState {
	return GenesisState{
		Params:       params,
		PoolRecords:  poolRecords,
		ShareRecords: shareRecords,
	}
}

// Validate validates the module's genesis state
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}
	if err := gs.PoolRecords.Validate(); err != nil {
		return err
	}
	if err := gs.ShareRecords.Validate(); err != nil {
		return err
	}

	totalShares := make(map[string]poolShares)
	for _, pr := range gs.PoolRecords {
		totalShares[pr.PoolID] = poolShares{
			totalShares:      pr.TotalShares,
			totalSharesOwned: sdk.ZeroInt(),
		}
	}
	for _, sr := range gs.ShareRecords {
		if shares, found := totalShares[sr.PoolID]; found {
			shares.totalSharesOwned = shares.totalSharesOwned.Add(sr.SharesOwned)
			totalShares[sr.PoolID] = shares
		} else {
			totalShares[sr.PoolID] = poolShares{
				totalShares:      sdk.ZeroInt(),
				totalSharesOwned: sr.SharesOwned,
			}
		}
	}

	for poolID, ps := range totalShares {
		if !ps.totalShares.Equal(ps.totalSharesOwned) {
			return fmt.Errorf("total depositor shares %s not equal to pool '%s' total shares %s", ps.totalSharesOwned.String(), poolID, ps.totalShares.String())
		}
	}

	return nil
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() GenesisState {
	return NewGenesisState(
		DefaultParams(),
		DefaultPoolRecords,
		DefaultShareRecords,
	)
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
