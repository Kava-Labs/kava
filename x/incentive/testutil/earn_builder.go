package testutil

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/kava-labs/kava/app"

	earntypes "github.com/kava-labs/kava/x/earn/types"
)

// EarnGenesisBuilder is a tool for creating a earn genesis state.
// Helper methods add values onto a default genesis state.
// All methods are immutable and return updated copies of the builder.
type EarnGenesisBuilder struct {
	earntypes.GenesisState
}

var _ GenesisBuilder = (*EarnGenesisBuilder)(nil)

func NewEarnGenesisBuilder() EarnGenesisBuilder {
	return EarnGenesisBuilder{
		GenesisState: earntypes.DefaultGenesisState(),
	}
}

func (builder EarnGenesisBuilder) Build() earntypes.GenesisState {
	return builder.GenesisState
}

func (builder EarnGenesisBuilder) BuildMarshalled(cdc codec.JSONCodec) app.GenesisState {
	built := builder.Build()

	return app.GenesisState{
		earntypes.ModuleName: cdc.MustMarshalJSON(&built),
	}
}

func (builder EarnGenesisBuilder) WithAllowedVaults(vault ...earntypes.AllowedVault) EarnGenesisBuilder {
	builder.Params.AllowedVaults = append(builder.Params.AllowedVaults, vault...)
	return builder
}
