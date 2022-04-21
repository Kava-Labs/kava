package v0_17

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"

	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
	feemarkettypes "github.com/tharsis/ethermint/x/feemarket/types"

	evmutiltypes "github.com/kava-labs/kava/x/evmutil/types"

	v016auction "github.com/kava-labs/kava/x/auction/legacy/v0_16"
	v017auction "github.com/kava-labs/kava/x/auction/legacy/v0_17"
)

func migrateAppState(appState genutiltypes.AppMap, clientCtx client.Context) {
	v17codec := clientCtx.Codec

	// x/emvutil
	evmUtilGenState := evmutiltypes.NewGenesisState([]evmutiltypes.Account{})
	appState[evmutiltypes.ModuleName] = v17codec.MustMarshalJSON(evmUtilGenState)

	// x/evm
	evmGenState := &evmtypes.GenesisState{
		Accounts: []evmtypes.GenesisAccount{},
		Params: evmtypes.Params{
			EvmDenom:     "akava",
			EnableCreate: true,
			EnableCall:   true,
			ChainConfig:  evmtypes.DefaultChainConfig(),
			ExtraEIPs:    nil,
		},
	}
	appState[evmtypes.ModuleName] = v17codec.MustMarshalJSON(evmGenState)

	// x/feemarket
	feemarketState := feemarkettypes.DefaultGenesisState()
	appState[feemarkettypes.ModuleName] = v17codec.MustMarshalJSON(feemarketState)

	// Migrate x/auction
	// Replace Auction state param BidDuration with
	// ReverseBidDuration and
	// ForwardBidDuration
	v16AuctionCodec := codec.NewProtoCodec(v016auction.InterfaceRegistry())
	if appState[v016auction.ModuleName] != nil {
		var genState v016auction.GenesisState
		v16AuctionCodec.MustUnmarshalJSON(appState[v016auction.ModuleName], &genState)

		// replace previous genesis state with migrated genesis state
		appState[v016auction.ModuleName] = v17codec.MustMarshalJSON(v017auction.Migrate(genState))

	}
}
