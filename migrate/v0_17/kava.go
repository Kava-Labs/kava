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
	hardtypes "github.com/kava-labs/kava/x/hard/types"
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

	if appState[committeetypes.ModuleName] != nil {
		var genState committeetypes.GenesisState
		codec.MustUnmarshalJSON(appState[committeetypes.ModuleName], &genState)

		migratedState := MigrateCommitteePermissions(genState)

		appState[committeetypes.ModuleName] = codec.MustMarshalJSON(&migratedState)
	}
}

func MigrateCommitteePermissions(genState committeetypes.GenesisState) committeetypes.GenesisState {

	var newCommittees committeetypes.Committees
	for _, committee := range genState.GetCommittees() {
		switch committee.GetDescription() {
		case "Hard Governance Committee":
			committee = fixHardPermissions(committee)
		case "Kava Stability Committee":
			committee = fixStabilityPermissions(committee)
		}
		newCommittees = append(newCommittees, committee)
	}

	packedCommittees, err := committeetypes.PackCommittees(newCommittees)
	if err != nil {
		panic(err)
	}
	genState.Committees = packedCommittees
	return genState
}

func fixStabilityPermissions(committee committeetypes.Committee) committeetypes.Committee {
	permissions := committee.GetPermissions()

	// get first params change permission in committee
	var perm *committeetypes.ParamsChangePermission
	var permIndex int
	var found bool
	for permIndex = range permissions {
		p, ok := permissions[permIndex].(*committeetypes.ParamsChangePermission)
		if ok {
			perm = p
			found = true
			break
		}
	}
	if !found {
		panic("ParamsChangePermission not found")
	}

	appendSubparamRequirement(
		perm.AllowedParamsChanges,
		hardtypes.ModuleName, string(hardtypes.KeyMoneyMarkets),
		committeetypes.SubparamRequirement{
			Key: "denom",
			Val: "ibc/799FDD409719A1122586A629AE8FCA17380351A51C1F47A80A1B8E7F2A491098",
			AllowedSubparamAttrChanges: []string{
				"borrow_limit",
				"interest_rate_model",
				"keeper_reward_percentage",
				"reserve_factor",
				"spot_market_id",
			},
		},
	)

	appendSubparamRequirement(
		perm.AllowedParamsChanges,
		cdptypes.ModuleName, string(cdptypes.KeyCollateralParams),
		committeetypes.SubparamRequirement{
			Key: "type",
			Val: "ust-a",
			AllowedSubparamAttrChanges: []string{
				"auction_size",
				"check_collateralization_index_count",
				"debt_limit",
				"keeper_reward_percentage",
				"stability_fee",
			},
		},
	)

	perm.AllowedParamsChanges.Delete(auctiontypes.ModuleName, "BidDuration")

	perm.AllowedParamsChanges.Set(committeetypes.AllowedParamsChange{
		Subspace: auctiontypes.ModuleName,
		Key:      string(auctiontypes.KeyForwardBidDuration),
	})
	perm.AllowedParamsChanges.Set(committeetypes.AllowedParamsChange{
		Subspace: auctiontypes.ModuleName,
		Key:      string(auctiontypes.KeyReverseBidDuration),
	})

	// update committee
	permissions[permIndex] = perm
	committee.SetPermissions(permissions)

	return committee
}

func fixHardPermissions(committee committeetypes.Committee) committeetypes.Committee {
	permissions := committee.GetPermissions()

	// get first params change permission in committee
	var perm *committeetypes.ParamsChangePermission
	var permIndex int
	var found bool
	for permIndex = range permissions {
		p, ok := permissions[permIndex].(*committeetypes.ParamsChangePermission)
		if ok {
			perm = p
			found = true
			break
		}
	}
	if !found {
		panic("ParamsChangePermission not found")
	}

	appendSubparamRequirement(
		perm.AllowedParamsChanges,
		hardtypes.ModuleName, string(hardtypes.KeyMoneyMarkets),
		committeetypes.SubparamRequirement{
			Key: "denom",
			Val: "ibc/799FDD409719A1122586A629AE8FCA17380351A51C1F47A80A1B8E7F2A491098",
			AllowedSubparamAttrChanges: []string{
				"borrow_limit",
				"interest_rate_model",
				"keeper_reward_percentage",
				"reserve_factor",
				"spot_market_id",
			},
		},
	)

	// update committee
	permissions[permIndex] = perm
	committee.SetPermissions(permissions)

	return committee
}

func appendSubparamRequirement(allowed committeetypes.AllowedParamsChanges, subspace, key string, requirement committeetypes.SubparamRequirement) {
	apc, found := allowed.Get(subspace, key)
	if !found {
		panic("AllowedParamsChange not found")
	}
	apc.MultiSubparamsRequirements = append(apc.MultiSubparamsRequirements, requirement)
	allowed.Set(apc)
}
