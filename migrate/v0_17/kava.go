package v0_17

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authz "github.com/cosmos/cosmos-sdk/x/authz"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
	feemarkettypes "github.com/tharsis/ethermint/x/feemarket/types"

	evmutiltypes "github.com/kava-labs/kava/x/evmutil/types"

	bridgetypes "github.com/kava-labs/kava-bridge/x/bridge/types"
	v016auction "github.com/kava-labs/kava/x/auction/legacy/v0_16"
	v017auction "github.com/kava-labs/kava/x/auction/legacy/v0_17"
	auctiontypes "github.com/kava-labs/kava/x/auction/types"
)

func migrateAppState(appState genutiltypes.AppMap, clientCtx client.Context) {
	interfaceRegistry := types.NewInterfaceRegistry()
	v016auction.RegisterInterfaces(interfaceRegistry)
	v16Codec := codec.NewProtoCodec(interfaceRegistry)

	codec := clientCtx.Codec

	// x/emvutil
	evmUtilGenState := evmutiltypes.NewGenesisState([]evmutiltypes.Account{})
	appState[evmutiltypes.ModuleName] = codec.MustMarshalJSON(evmUtilGenState)

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
	appState[evmtypes.ModuleName] = codec.MustMarshalJSON(evmGenState)

	// x/bridge
	bridgeGenState := bridgetypes.NewGenesisState(
		bridgetypes.NewParams(
			false,                            // Bridge disabled
			bridgetypes.EnabledERC20Tokens{}, // No bridge ERC20 tokens
			nil,                              // No relayer
			bridgetypes.ConversionPairs{},    // No conversion pairs
		),
		bridgetypes.ERC20BridgePairs{}, // Empty state as there has been no ERC20 contracts deployed
		sdk.OneInt(),                   // NextWithdrawSequence starts at 1
	)
	appState[bridgetypes.ModuleName] = codec.MustMarshalJSON(&bridgeGenState)

	// x/feemarket
	feemarketState := feemarkettypes.DefaultGenesisState()
	appState[feemarkettypes.ModuleName] = codec.MustMarshalJSON(feemarketState)

	// x/authz
	authzState := authz.DefaultGenesisState()
	appState[authz.ModuleName] = codec.MustMarshalJSON(authzState)

	// x/auction
	if appState[auctiontypes.ModuleName] != nil {
		var v16GenState v016auction.GenesisState
		v16Codec.MustUnmarshalJSON(appState[auctiontypes.ModuleName], &v16GenState)

		migratedState := v017auction.Migrate(v16GenState)
		encodedState := codec.MustMarshalJSON(migratedState)

		appState[auctiontypes.ModuleName] = encodedState
	}
}
