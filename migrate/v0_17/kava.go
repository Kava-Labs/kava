package v0_17

import (
	"github.com/cosmos/cosmos-sdk/client"

	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
	feemarkettypes "github.com/tharsis/ethermint/x/feemarket/types"

	evmutiltypes "github.com/kava-labs/kava/x/evmutil/types"
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

	// x/feemarket
	feemarketState := feemarkettypes.DefaultGenesisState()
	appState[feemarkettypes.ModuleName] = codec.MustMarshalJSON(feemarketState)
}
