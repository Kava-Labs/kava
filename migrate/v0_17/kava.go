package v0_17

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"

	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	v016auction "github.com/kava-labs/kava/x/auction/legacy/v0_16"
	v017auction "github.com/kava-labs/kava/x/auction/legacy/v0_17"
)

func migrateKavaAppState(appState genutiltypes.AppMap, clientCtx client.Context) {

	v16AuctionCodec := codec.NewProtoCodec(v016auction.InterfaceRegistry())
	v17Codec := clientCtx.Codec

	// Migrate x/auction
	// Replace Auction state param BidDuration with
	// ReverseBidDuration and
	// ForwardBidDuration
	if appState[v016auction.ModuleName] != nil {
		var genState v016auction.GenesisState
		v16AuctionCodec.MustUnmarshalJSON(appState[v016auction.ModuleName], &genState)

		// replace previous genesis state with migrated genesis state
		appState[v016auction.ModuleName] = v17Codec.MustMarshalJSON(v017auction.Migrate(genState))

	}
}
