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
	v017bep3 "github.com/kava-labs/kava/x/bep3/legacy/v0_17"
	bep3types "github.com/kava-labs/kava/x/bep3/types"
	cdptypes "github.com/kava-labs/kava/x/cdp/types"
	committeetypes "github.com/kava-labs/kava/x/committee/types"
	incentivetypes "github.com/kava-labs/kava/x/incentive/types"
	savingstypes "github.com/kava-labs/kava/x/savings/types"
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
	evmChainConfig := evmtypes.DefaultChainConfig()
	evmChainConfig.LondonBlock = nil
	evmChainConfig.ArrowGlacierBlock = nil
	evmChainConfig.MergeForkBlock = nil

	evmGenState := &evmtypes.GenesisState{
		Accounts: []evmtypes.GenesisAccount{},
		Params: evmtypes.Params{
			EvmDenom:     "akava",
			EnableCreate: true,
			EnableCall:   true,
			ChainConfig:  evmChainConfig,
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
	// disable fee market and use minimum-gas-price instead of dynamic base fee
	feemarketState.Params.NoBaseFee = true
	appState[feemarkettypes.ModuleName] = codec.MustMarshalJSON(feemarketState)

	// x/authz
	authzState := authz.DefaultGenesisState()
	appState[authz.ModuleName] = codec.MustMarshalJSON(authzState)

	// x/cdp
	if appState[cdptypes.ModuleName] != nil {
		var genState cdptypes.GenesisState
		codec.MustUnmarshalJSON(appState[cdptypes.ModuleName], &genState)

		genState.Params.GlobalDebtLimit = sdk.NewCoin("usdx", sdk.NewInt(393000000000000))
		encodedState := codec.MustMarshalJSON(&genState)

		appState[cdptypes.ModuleName] = encodedState
	}

	// x/auction
	if appState[auctiontypes.ModuleName] != nil {
		var v16GenState v016auction.GenesisState
		v16Codec.MustUnmarshalJSON(appState[auctiontypes.ModuleName], &v16GenState)

		migratedState := v017auction.Migrate(v16GenState)
		encodedState := codec.MustMarshalJSON(migratedState)

		appState[auctiontypes.ModuleName] = encodedState
	}

	// x/incentive
	if appState[incentivetypes.ModuleName] != nil {
		var incentiveState incentivetypes.GenesisState
		codec.MustUnmarshalJSON(appState[incentivetypes.ModuleName], &incentiveState)

		appState[incentivetypes.ModuleName] = codec.MustMarshalJSON(&incentiveState)
	}

	// x/savings
	savingsState := savingstypes.DefaultGenesisState()
	appState[savingstypes.ModuleName] = codec.MustMarshalJSON(&savingsState)

	// x/bep3
	if appState[bep3types.ModuleName] != nil {
		var v16GenState bep3types.GenesisState
		codec.MustUnmarshalJSON(appState[bep3types.ModuleName], &v16GenState)

		migratedState := v017bep3.Migrate(v16GenState)

		appState[bep3types.ModuleName] = codec.MustMarshalJSON(migratedState)
	}

	// x/committee
	if appState[committeetypes.ModuleName] != nil {
		var genState committeetypes.GenesisState
		codec.MustUnmarshalJSON(appState[committeetypes.ModuleName], &genState)

		migratedState := migrateCommitteePermissions(genState)

		appState[committeetypes.ModuleName] = codec.MustMarshalJSON(&migratedState)
	}
}
