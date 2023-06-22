package testutil

import (
	"github.com/cosmos/cosmos-sdk/codec"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/kava-labs/kava/app"
)

// StakingGenesisBuilder is a tool for creating a staking genesis state.
// Helper methods add values onto a default genesis state.
// All methods are immutable and return updated copies of the builder.
type StakingGenesisBuilder struct {
	stakingtypes.GenesisState
}

var _ GenesisBuilder = (*StakingGenesisBuilder)(nil)

func NewStakingGenesisBuilder() StakingGenesisBuilder {
	gen := stakingtypes.DefaultGenesisState()
	gen.Params.BondDenom = "ukava"

	return StakingGenesisBuilder{
		GenesisState: *gen,
	}
}

func (builder StakingGenesisBuilder) Build() stakingtypes.GenesisState {
	return builder.GenesisState
}

func (builder StakingGenesisBuilder) BuildMarshalled(cdc codec.JSONCodec) app.GenesisState {
	built := builder.Build()

	return app.GenesisState{
		stakingtypes.ModuleName: cdc.MustMarshalJSON(&built),
	}
}
