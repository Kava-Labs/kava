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
	v015bep3 "github.com/kava-labs/kava/x/bep3/legacy/v0_15"
	v016bep3 "github.com/kava-labs/kava/x/bep3/legacy/v0_16"
	v015cdp "github.com/kava-labs/kava/x/cdp/legacy/v0_15"
	v016cdp "github.com/kava-labs/kava/x/cdp/legacy/v0_16"
	v015committee "github.com/kava-labs/kava/x/committee/legacy/v0_15"
	v016committee "github.com/kava-labs/kava/x/committee/legacy/v0_16"
	v015hard "github.com/kava-labs/kava/x/hard/legacy/v0_15"
	v016hard "github.com/kava-labs/kava/x/hard/legacy/v0_16"
	v015incentive "github.com/kava-labs/kava/x/incentive/legacy/v0_15"
	v016incentive "github.com/kava-labs/kava/x/incentive/legacy/v0_16"
	v015issuance "github.com/kava-labs/kava/x/issuance/legacy/v0_15"
	v016issuance "github.com/kava-labs/kava/x/issuance/legacy/v0_16"
	v015kavadist "github.com/kava-labs/kava/x/kavadist/legacy/v0_15"
	v016kavadist "github.com/kava-labs/kava/x/kavadist/legacy/v0_16"
	v015pricefeed "github.com/kava-labs/kava/x/pricefeed/legacy/v0_15"
	v016pricefeed "github.com/kava-labs/kava/x/pricefeed/legacy/v0_16"
	v015swap "github.com/kava-labs/kava/x/swap/legacy/v0_15"
	v016swap "github.com/kava-labs/kava/x/swap/legacy/v0_16"
	v015validatorvesting "github.com/kava-labs/kava/x/validator-vesting/legacy/v0_15"
)

func migrateKavaAppState(appState genutiltypes.AppMap, clientCtx client.Context) {
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
		if appState[v015pricefeed.ModuleName] == nil {
			panic("pricefeed app state is missing, committee migration requires pricefeed app state")
		}

		var pricefeedGenState v015pricefeed.GenesisState
		v15Codec.MustUnmarshalJSON(appState[v015pricefeed.ModuleName], &pricefeedGenState)

		// unmarshal relative source genesis application state
		var genState v015committee.GenesisState
		v15Codec.MustUnmarshalJSON(appState[v015committee.ModuleName], &genState)

		// replace migrated genstate with previous genstate
		appState[v015committee.ModuleName] = v16Codec.MustMarshalJSON(v016committee.Migrate(genState, pricefeedGenState))
	}

	// Migrate x/bep3
	if appState[v015bep3.ModuleName] != nil {
		// unmarshal relative source genesis application state
		var genState v015bep3.GenesisState
		v15Codec.MustUnmarshalJSON(appState[v015bep3.ModuleName], &genState)

		// replace migrated genstate with previous genstate
		appState[v015bep3.ModuleName] = v16Codec.MustMarshalJSON(v016bep3.Migrate(genState))
	}

	// Migrate x/swap
	if appState[v015swap.ModuleName] != nil {
		// unmarshal relative source genesis application state
		var genState v015swap.GenesisState
		v15Codec.MustUnmarshalJSON(appState[v015swap.ModuleName], &genState)

		// replace migrated genstate with previous genstate
		appState[v015swap.ModuleName] = v16Codec.MustMarshalJSON(v016swap.Migrate(genState))
	}

	// Migrate x/kavadist
	if appState[v015kavadist.ModuleName] != nil {
		// unmarshal relative source genesis application state
		var genState v015kavadist.GenesisState
		v15Codec.MustUnmarshalJSON(appState[v015kavadist.ModuleName], &genState)

		// replace migrated genstate with previous genstate
		appState[v015kavadist.ModuleName] = v16Codec.MustMarshalJSON(v016kavadist.Migrate(genState))
	}

	// Migrate x/cdp
	if appState[v015cdp.ModuleName] != nil {
		// unmarshal relative source genesis application state
		var genState v015cdp.GenesisState
		v15Codec.MustUnmarshalJSON(appState[v015cdp.ModuleName], &genState)

		// replace migrated genstate with previous genstate
		appState[v015cdp.ModuleName] = v16Codec.MustMarshalJSON(v016cdp.Migrate(genState))
	}

	// Migrate x/issuance
	if appState[v015issuance.ModuleName] != nil {
		// unmarshal relative source genesis application state
		var genState v015issuance.GenesisState
		v15Codec.MustUnmarshalJSON(appState[v015issuance.ModuleName], &genState)

		// replace migrated genstate with previous genstate
		appState[v015issuance.ModuleName] = v16Codec.MustMarshalJSON(v016issuance.Migrate(genState))
	}

	// Migrate x/pricefeed
	if appState[v015pricefeed.ModuleName] != nil {
		// unmarshal relative source genesis application state
		var genState v015pricefeed.GenesisState
		v15Codec.MustUnmarshalJSON(appState[v015pricefeed.ModuleName], &genState)

		// replace migrated genstate with previous genstate
		appState[v015pricefeed.ModuleName] = v16Codec.MustMarshalJSON(v016pricefeed.Migrate(genState))
	}

	// Migrate x/hard
	if appState[v015hard.ModuleName] != nil {
		// unmarshal relative source genesis application state
		var genState v015hard.GenesisState
		v15Codec.MustUnmarshalJSON(appState[v015hard.ModuleName], &genState)

		// replace migrated genstate with previous genstate
		appState[v015hard.ModuleName] = v16Codec.MustMarshalJSON(v016hard.Migrate(genState))
	}

	// Migrate x/incentive
	if appState[v015incentive.ModuleName] != nil {
		// unmarshal relative source genesis application state
		var genState v015incentive.GenesisState
		v15Codec.MustUnmarshalJSON(appState[v015incentive.ModuleName], &genState)

		// replace migrated genstate with previous genstate
		appState[v015incentive.ModuleName] = v16Codec.MustMarshalJSON(v016incentive.Migrate(genState))
	}

	// Remove x/validator-vesting
	delete(appState, v015validatorvesting.ModuleName)
}
