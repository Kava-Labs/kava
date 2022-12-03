package testutil

import (
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/app"

	kavaminttypes "github.com/kava-labs/kava/x/kavamint/types"
)

// KavamintGenesisBuilder is a tool for creating a mint genesis state.
// Helper methods add values onto a default genesis state.
// All methods are immutable and return updated copies of the builder.
type KavamintGenesisBuilder struct {
	kavaminttypes.GenesisState
}

var _ GenesisBuilder = (*KavamintGenesisBuilder)(nil)

func NewKavamintGenesisBuilder() KavamintGenesisBuilder {
	gen := kavaminttypes.DefaultGenesisState()
	gen.Params.CommunityPoolInflation = sdk.ZeroDec()
	gen.Params.StakingRewardsApy = sdk.ZeroDec()

	return KavamintGenesisBuilder{
		GenesisState: *gen,
	}
}

func (builder KavamintGenesisBuilder) Build() kavaminttypes.GenesisState {
	return builder.GenesisState
}

func (builder KavamintGenesisBuilder) BuildMarshalled(cdc codec.JSONCodec) app.GenesisState {
	built := builder.Build()

	return app.GenesisState{
		kavaminttypes.ModuleName: cdc.MustMarshalJSON(&built),
	}
}

func (builder KavamintGenesisBuilder) WithPreviousBlockTime(t time.Time) KavamintGenesisBuilder {
	builder.PreviousBlockTime = t
	return builder
}

func (builder KavamintGenesisBuilder) WithStakingRewardsApy(apy sdk.Dec) KavamintGenesisBuilder {
	builder.Params.StakingRewardsApy = apy
	return builder
}

func (builder KavamintGenesisBuilder) WithCommunityPoolInflation(
	inflation sdk.Dec,
) KavamintGenesisBuilder {
	builder.Params.CommunityPoolInflation = inflation
	return builder
}
