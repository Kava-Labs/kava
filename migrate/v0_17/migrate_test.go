package v0_17

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	bridgetypes "github.com/kava-labs/kava-bridge/x/bridge/types"
	"github.com/kava-labs/kava/app"
	"github.com/stretchr/testify/assert"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authz "github.com/cosmos/cosmos-sdk/x/authz"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	tmjson "github.com/tendermint/tendermint/libs/json"
	tmtypes "github.com/tendermint/tendermint/types"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
	feemarkettypes "github.com/tharsis/ethermint/x/feemarket/types"

	auctiontypes "github.com/kava-labs/kava/x/auction/types"
	evmutiltypes "github.com/kava-labs/kava/x/evmutil/types"
	incentivetypes "github.com/kava-labs/kava/x/incentive/types"
)

func TestMigrateGenesisDoc(t *testing.T) {
	expected := getTestDataJSON("genesis-v17.json")
	genDoc, err := tmtypes.GenesisDocFromFile(filepath.Join("testdata", "genesis-v16.json"))
	assert.NoError(t, err)

	actualGenDoc, err := Migrate(genDoc, newClientContext())
	assert.NoError(t, err)

	actualJson, err := tmjson.Marshal(actualGenDoc)
	assert.NoError(t, err)

	assert.JSONEq(t, expected, string(actualJson))
}

func TestMigrateEvmUtil(t *testing.T) {
	appMap, ctx := migrateToV17AndGetAppMap(t)
	var genstate evmutiltypes.GenesisState
	err := ctx.Codec.UnmarshalJSON(appMap[evmutiltypes.ModuleName], &genstate)
	assert.NoError(t, err)
	assert.Len(t, genstate.Accounts, 0)
}

func TestMigrateEvm(t *testing.T) {
	appMap, ctx := migrateToV17AndGetAppMap(t)
	var genstate evmtypes.GenesisState
	err := ctx.Codec.UnmarshalJSON(appMap[evmtypes.ModuleName], &genstate)
	assert.NoError(t, err)
	assert.Len(t, genstate.Accounts, 0)
	assert.Equal(t, genstate.Params, evmtypes.Params{
		EvmDenom:     "akava",
		EnableCreate: true,
		EnableCall:   true,
		ChainConfig:  evmtypes.DefaultChainConfig(),
		ExtraEIPs:    []int64{},
	})
}

func TestMigrateAuction(t *testing.T) {
	appMap, ctx := migrateToV17AndGetAppMap(t)
	var genstate auctiontypes.GenesisState
	err := ctx.Codec.UnmarshalJSON(appMap[auctiontypes.ModuleName], &genstate)
	assert.NoError(t, err)
	assert.Len(t, genstate.Auctions, 3)
}

func TestMigrateFeeMarket(t *testing.T) {
	appMap, ctx := migrateToV17AndGetAppMap(t)
	var genstate feemarkettypes.GenesisState
	err := ctx.Codec.UnmarshalJSON(appMap[feemarkettypes.ModuleName], &genstate)
	assert.NoError(t, err)
	assert.Equal(t, genstate, *feemarkettypes.DefaultGenesisState())
}

func TestMigrateAuthz(t *testing.T) {
	appMap, ctx := migrateToV17AndGetAppMap(t)
	var genstate authz.GenesisState
	err := ctx.Codec.UnmarshalJSON(appMap[authz.ModuleName], &genstate)
	assert.NoError(t, err)
	assert.Equal(t, genstate, authz.GenesisState{
		Authorization: []authz.GrantAuthorization{},
	})
}

func TestMigrateBridge(t *testing.T) {
	appMap, ctx := migrateToV17AndGetAppMap(t)
	var genstate bridgetypes.GenesisState
	err := ctx.Codec.UnmarshalJSON(appMap[bridgetypes.ModuleName], &genstate)
	assert.NoError(t, err)

	assert.Len(t, genstate.ERC20BridgePairs, 0)
	assert.Equal(t, genstate.NextWithdrawSequence, sdk.OneInt())
	assert.Equal(t, genstate.Params, bridgetypes.Params{
		BridgeEnabled:          false,
		EnabledERC20Tokens:     bridgetypes.EnabledERC20Tokens{},
		Relayer:                sdk.AccAddress{},
		EnabledConversionPairs: bridgetypes.ConversionPairs{},
	})
}

func TestMigrateIncentive(t *testing.T) {
	appMap, ctx := migrateToV17AndGetAppMap(t)
	var genstate incentivetypes.GenesisState
	err := ctx.Codec.UnmarshalJSON(appMap[incentivetypes.ModuleName], &genstate)
	assert.NoError(t, err)
	assert.Len(t, genstate.SavingsClaims, 0)
	assert.Len(t, genstate.SavingsRewardState.AccumulationTimes, 0)
	assert.Len(t, genstate.SavingsRewardState.MultiRewardIndexes, 0)
	assert.Len(t, genstate.Params.SavingsRewardPeriods, 0)
}

func migrateToV17AndGetAppMap(t *testing.T) (genutiltypes.AppMap, client.Context) {
	genDoc, err := tmtypes.GenesisDocFromFile(filepath.Join("testdata", "genesis-v16.json"))
	assert.NoError(t, err)

	ctx := newClientContext()
	actualGenDoc, err := Migrate(genDoc, ctx)
	assert.NoError(t, err)

	var appMap genutiltypes.AppMap
	err = tmjson.Unmarshal(actualGenDoc.AppState, &appMap)
	assert.NoError(t, err)

	return appMap, ctx
}

func getTestDataJSON(filename string) string {
	file := filepath.Join("testdata", filename)
	data, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func newClientContext() client.Context {
	config := app.MakeEncodingConfig()
	return client.Context{}.
		WithCodec(config.Marshaler).
		WithLegacyAmino(config.Amino).
		WithInterfaceRegistry(config.InterfaceRegistry)
}
