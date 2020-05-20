package v038

import (
	v038slashing "github.com/cosmos/cosmos-sdk/x/slashing"

	"github.com/kava-labs/kava/migrate/v0_8/sdk/slashing/v18de63"
)

func Migrate(oldGenState v18de63.GenesisState) v038slashing.GenesisState {

	// old and new types are identical except for a new HistoricalEntries field

	newParams := v038slashing.Params{
		// no MaxEvidenceAge
		SignedBlocksWindow:      oldGenState.Params.SignedBlocksWindow,
		MinSignedPerWindow:      oldGenState.Params.MinSignedPerWindow,
		DowntimeJailDuration:    oldGenState.Params.DowntimeJailDuration,
		SlashFractionDoubleSign: oldGenState.Params.SlashFractionDoubleSign,
		SlashFractionDowntime:   oldGenState.Params.SlashFractionDowntime,
	}

	return v038slashing.GenesisState{
		Params:       newParams,
		SigningInfos: oldGenState.SigningInfos,
		MissedBlocks: oldGenState.MissedBlocks,
	}
}
