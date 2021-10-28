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
	DefaultPoolRecords = []PoolRecord{}
	// DefaultShareRecords is used to set default records in default genesis state
	DefaultShareRecords = []ShareRecord{}
)

// NewGenesisState creates a new genesis state.
func NewGenesisState(params Params, poolRecords []PoolRecord, shareRecords []ShareRecord) GenesisState {
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
	if err := ValidatePoolRecords(gs.PoolRecords); err != nil {
		return err
	}
	if err := ValidateShareRecords(gs.ShareRecords); err != nil {
		return err
	}

	totalShares := make(map[string]poolShares)
	for _, pr := range gs.PoolRecords {
		totalShares[pr.PoolId] = poolShares{
			totalShares:      pr.TotalShares,
			totalSharesOwned: sdk.ZeroInt(),
		}
	}
	for _, sr := range gs.ShareRecords {
		if shares, found := totalShares[sr.PoolId]; found {
			shares.totalSharesOwned = shares.totalSharesOwned.Add(sr.SharesOwned)
			totalShares[sr.PoolId] = shares
		} else {
			totalShares[sr.PoolId] = poolShares{
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
	b1 := ModuleCdc.MustMarshal(&gs)
	b2 := ModuleCdc.MustMarshal(&gs2)
	return bytes.Equal(b1, b2)
}

// IsEmpty returns true if a GenesisState is empty
func (gs GenesisState) IsEmpty() bool {
	return gs.Equal(GenesisState{})
}
