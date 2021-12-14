package v0_16

import (
	v015kavadist "github.com/kava-labs/kava/x/kavadist/legacy/v0_15"
	v016kavadist "github.com/kava-labs/kava/x/kavadist/types"
)

func migrateParams(oldParams v015kavadist.Params) v016kavadist.Params {
	periods := make([]v016kavadist.Period, len(oldParams.Periods))
	for i, oldPeriod := range oldParams.Periods {
		periods[i] = v016kavadist.Period{
			Start:     oldPeriod.Start,
			End:       oldPeriod.End,
			Inflation: oldPeriod.Inflation,
		}
	}
	return v016kavadist.Params{
		Periods: periods,
		Active:  oldParams.Active,
	}
}

// Migrate converts v0.15 kavadist state and returns it in v0.16 format
func Migrate(oldState v015kavadist.GenesisState) *v016kavadist.GenesisState {
	return &v016kavadist.GenesisState{
		Params:            migrateParams(oldState.Params),
		PreviousBlockTime: oldState.PreviousBlockTime,
	}
}
