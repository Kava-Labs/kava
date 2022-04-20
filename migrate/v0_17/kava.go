package v0_17

import (
	"github.com/cosmos/cosmos-sdk/client"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authz "github.com/cosmos/cosmos-sdk/x/authz"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
	feemarkettypes "github.com/tharsis/ethermint/x/feemarket/types"

	evmutiltypes "github.com/kava-labs/kava/x/evmutil/types"

	bridgetypes "github.com/kava-labs/kava-bridge/x/bridge/types"
)

func migrateAppState(appState genutiltypes.AppMap, clientCtx client.Context) {
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
		// Start with no ERC20 tokens that can be bridged, no relayer set.
		bridgetypes.NewParams(
			false, // bridge disabled
			bridgetypes.EnabledERC20Tokens{},
			nil, // no relayer
			bridgetypes.ConversionPairs{},
		),
		// No ERC20 tokens have been bridged yet.
		bridgetypes.ERC20BridgePairs{},
		// NextWithdrawSequence starts at 1
		sdk.OneInt(),
	)
	appState[bridgetypes.ModuleName] = codec.MustMarshalJSON(&bridgeGenState)

	// x/feemarket
	feemarketState := feemarkettypes.DefaultGenesisState()
	appState[feemarkettypes.ModuleName] = codec.MustMarshalJSON(feemarketState)

	// x/authz
	authzState := authz.DefaultGenesisState()
	appState[authz.ModuleName] = codec.MustMarshalJSON(authzState)
}
