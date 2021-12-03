package v0_16

import (
	v015swap "github.com/kava-labs/kava/x/swap/legacy/v0_15"
	v016swap "github.com/kava-labs/kava/x/swap/types"
)

func migrateParams(params v015swap.Params) v016swap.Params {
	allowedPools := make(v016swap.AllowedPools, len(params.AllowedPools))
	for i, pool := range params.AllowedPools {
		allowedPools[i] = v016swap.AllowedPool{
			TokenA: pool.TokenA,
			TokenB: pool.TokenB,
		}
	}
	return v016swap.Params{
		AllowedPools: allowedPools,
		SwapFee:      params.SwapFee,
	}
}

func migratePoolRecords(oldRecords v015swap.PoolRecords) v016swap.PoolRecords {
	newRecords := make(v016swap.PoolRecords, len(oldRecords))
	for i, oldRecord := range oldRecords {
		newRecords[i] = v016swap.PoolRecord{
			PoolID:      oldRecord.PoolID,
			ReservesA:   oldRecord.ReservesA,
			ReservesB:   oldRecord.ReservesB,
			TotalShares: oldRecord.TotalShares,
		}
	}
	return newRecords
}

func migrateShareRecords(oldRecords v015swap.ShareRecords) v016swap.ShareRecords {
	newRecords := make(v016swap.ShareRecords, len(oldRecords))
	for i, oldRecord := range oldRecords {
		newRecords[i] = v016swap.ShareRecord{
			Depositor:   oldRecord.Depositor,
			PoolID:      oldRecord.PoolID,
			SharesOwned: oldRecord.SharesOwned,
		}
	}
	return newRecords
}

// Migrate converts v0.15 swap state and returns it in v0.16 format
func Migrate(oldState v015swap.GenesisState) *v016swap.GenesisState {
	return &v016swap.GenesisState{
		Params:       migrateParams(oldState.Params),
		PoolRecords:  migratePoolRecords(oldState.PoolRecords),
		ShareRecords: migrateShareRecords(oldState.ShareRecords),
	}
}
