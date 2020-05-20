package v038

import (
	v038staking "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/kava-labs/kava/migrate/v0_8/sdk/staking/v18de63"
)

func Migrate(oldGenState v18de63.GenesisState) v038staking.GenesisState {

	// old and new types are identical except for a new HistoricalEntries field

	newParams := v038staking.Params{
		UnbondingTime:     oldGenState.Params.UnbondingTime,
		MaxValidators:     oldGenState.Params.MaxValidators,
		MaxEntries:        oldGenState.Params.MaxEntries,
		HistoricalEntries: v038staking.DefaultHistoricalEntries, // TODO this is zero, so this whole migration might not be needed
		BondDenom:         oldGenState.Params.BondDenom,
	}

	return v038staking.GenesisState{
		Params:               newParams,
		LastTotalPower:       oldGenState.LastTotalPower,
		LastValidatorPowers:  oldGenState.LastValidatorPowers,
		Validators:           oldGenState.Validators,
		Delegations:          oldGenState.Delegations,
		UnbondingDelegations: oldGenState.UnbondingDelegations,
		Redelegations:        oldGenState.Redelegations,
		Exported:             oldGenState.Exported,
	}
}
