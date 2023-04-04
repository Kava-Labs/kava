package types

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type poolShares struct {
	totalShares      sdkmath.Int
	totalSharesOwned sdkmath.Int
}

var (
	// DefaultPoolRecords is used to set default records in default genesis state
	DefaultPoolRecords = PoolRecords{}
	// DefaultShareRecords is used to set default records in default genesis state
	DefaultShareRecords = ShareRecords{}
)

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
