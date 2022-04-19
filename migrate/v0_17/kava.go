package v0_17

import (
	"github.com/cosmos/cosmos-sdk/client"

	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	v017auction "github.com/kava-labs/kava/x/auction/types"
)

func migrateKavaAppState(appState genutiltypes.AppMap, clientCtx client.Context) {

	v17Codec := clientCtx.Codec

	// Migrate x/auction
	if appState[v017auction.ModuleName] != nil {
		// Since v0.16 genesis state  is serialized to JSON
		// using backwards-compatible protobuf definitions
		// Thus we only can directly unmarshal the v0.16 genesis state
		// without using legacy codecs such as Amino
		var genState v017auction.GenesisState
		v17Codec.MustUnmarshalJSON(appState[v017auction.ModuleName], &genState)

		// V17 Auction migration changes
		// - Replace singular default auction `Duration` param with
		// directional duration params
		genState.DefaultForwardBidDuration = v017auction.DefaultForwardBidDuration
		genState.DefaultReverseBidDuration = v017auction.DefaultReverseBidDuration

		// replace previous genesis state with migrated genesis state
		appState[v017auction.ModuleName] = v17Codec.MustMarshalJSON(v017auction.Migrate(genState))
	}
}
