package testutil

import (
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/kava-labs/kava/app"

	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

// MintGenesisBuilder is a tool for creating a mint genesis state.
// Helper methods add values onto a default genesis state.
// All methods are immutable and return updated copies of the builder.
type MintGenesisBuilder struct {
	minttypes.GenesisState
}

var _ GenesisBuilder = (*MintGenesisBuilder)(nil)

func NewMintGenesisBuilder() MintGenesisBuilder {
	gen := minttypes.DefaultGenesisState()
	gen.Params.MintDenom = "ukava"

	return MintGenesisBuilder{
		GenesisState: *gen,
	}
}

func (builder MintGenesisBuilder) Build() minttypes.GenesisState {
	return builder.GenesisState
}

func (builder MintGenesisBuilder) BuildMarshalled(cdc codec.JSONCodec) app.GenesisState {
	built := builder.Build()

	return app.GenesisState{
		minttypes.ModuleName: cdc.MustMarshalJSON(&built),
	}
}

func (builder MintGenesisBuilder) WithMinter(
	inflation sdkmath.LegacyDec,
	annualProvisions sdkmath.LegacyDec,
) MintGenesisBuilder {
	builder.Minter = minttypes.NewMinter(inflation, annualProvisions)
	return builder
}

func (builder MintGenesisBuilder) WithInflationMax(
	inflationMax sdkmath.LegacyDec,
) MintGenesisBuilder {
	builder.Params.InflationMax = inflationMax
	return builder
}

func (builder MintGenesisBuilder) WithInflationMin(
	inflationMin sdkmath.LegacyDec,
) MintGenesisBuilder {
	builder.Params.InflationMin = inflationMin
	return builder
}

func (builder MintGenesisBuilder) WithMintDenom(
	mintDenom string,
) MintGenesisBuilder {
	builder.Params.MintDenom = mintDenom
	return builder
}
