package v0_16

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"

	v036distr "github.com/cosmos/cosmos-sdk/x/distribution/legacy/v036"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	v036params "github.com/cosmos/cosmos-sdk/x/params/legacy/v036"
	v038upgrade "github.com/cosmos/cosmos-sdk/x/upgrade/legacy/v038"

	v015auction "github.com/kava-labs/kava/x/auction/legacy/v0_15"
	v016auction "github.com/kava-labs/kava/x/auction/legacy/v0_16"
	v015committee "github.com/kava-labs/kava/x/committee/legacy/v0_15"
	v016committee "github.com/kava-labs/kava/x/committee/legacy/v0_16"
	v015kavadist "github.com/kava-labs/kava/x/kavadist/legacy/v0_15"
)

func migrateKavaAppState(appState genutiltypes.AppMap, clientCtx client.Context) genutiltypes.AppMap {
	v15Codec := codec.NewLegacyAmino()
	v015auction.RegisterLegacyAminoCodec(v15Codec)
	v015committee.RegisterLegacyAminoCodec(v15Codec)
	v015kavadist.RegisterLegacyAminoCodec(v15Codec)
	v036distr.RegisterLegacyAminoCodec(v15Codec)
	v038upgrade.RegisterLegacyAminoCodec(v15Codec)
	v036params.RegisterLegacyAminoCodec(v15Codec)

	v16Codec := clientCtx.Codec

	// Migrate x/auction
	if appState[v015auction.ModuleName] != nil {
		// unmarshal relative source genesis application state
		var genState v015auction.GenesisState
		v15Codec.MustUnmarshalJSON(appState[v015auction.ModuleName], &genState)

		// replace migrated genstate with previous genstate
		appState[v015auction.ModuleName] = v16Codec.MustMarshalJSON(v016auction.Migrate(genState))
	}

	// Migrate x/committee
	if appState[v015committee.ModuleName] != nil {
		// unmarshal relative source genesis application state
		var genState v015committee.GenesisState
		v15Codec.MustUnmarshalJSON(appState[v015committee.ModuleName], &genState)

		// replace migrated genstate with previous genstate
		appState[v015committee.ModuleName] = v16Codec.MustMarshalJSON(v016committee.Migrate(genState))
	}

	return appState
}
