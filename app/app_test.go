package app

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"
	db "github.com/tendermint/tm-db"
)

func TestNewApp(t *testing.T) {
	SetSDKConfig()
	NewApp(
		log.NewTMLogger(log.NewSyncWriter(os.Stdout)),
		db.NewMemDB(),
		DefaultNodeHome,
		nil,
		MakeEncodingConfig(),
		DefaultOptions,
	)
}

func TestExport(t *testing.T) {
	SetSDKConfig()
	db := db.NewMemDB()
	app := NewApp(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, DefaultNodeHome, nil, MakeEncodingConfig(), DefaultOptions)

	genesisState := GenesisStateWithSingleValidator(&TestApp{App: *app}, NewDefaultGenesisState())

	stateBytes, err := json.Marshal(genesisState)
	require.NoError(t, err)

	initRequest := abci.RequestInitChain{
		Time:            time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC),
		ChainId:         "kavatest_1-1",
		InitialHeight:   1,
		ConsensusParams: tmtypes.TM2PB.ConsensusParams(tmtypes.DefaultConsensusParams()),
		Validators:      nil,
		AppStateBytes:   stateBytes,
	}
	app.InitChain(initRequest)
	app.Commit()

	exportedApp, err := app.ExportAppStateAndValidators(false, []string{})
	require.NoError(t, err)

	// Assume each module is exported correctly, so only check modules in genesis are present in export
	initialModules, err := unmarshalJSONKeys(initRequest.AppStateBytes)
	require.NoError(t, err)
	exportedModules, err := unmarshalJSONKeys(exportedApp.AppState)
	require.NoError(t, err)
	assert.ElementsMatch(t, initialModules, exportedModules)

	assert.Equal(t, initRequest.InitialHeight+1, exportedApp.Height) // app.Commit() increments height
	assert.Equal(t, initRequest.ConsensusParams, exportedApp.ConsensusParams)
	assert.Len(t, exportedApp.Validators, 1) // no validators set in default genesis
}

// TestLegacyMsgAreAminoRegistered checks if all known msg types are registered on the app's amino codec.
// It doesn't check if they are registered on the module codecs used for signature checking.
func TestLegacyMsgAreAminoRegistered(t *testing.T) {
	tApp := NewTestApp()

	lcdc := tApp.LegacyAmino()

	// Use the proto codec as the canonical list of msg types.
	protoCodec := tApp.AppCodec().(*codec.ProtoCodec)
	protoRegisteredMsgs := protoCodec.InterfaceRegistry().ListImplementations(sdk.MsgInterfaceProtoName)

	for i, msgName := range protoRegisteredMsgs {
		// Skip msgs from dependencies that were never amino registered.
		if msgName == sdk.MsgTypeURL(&evmtypes.MsgEthereumTx{}) ||
			msgName == sdk.MsgTypeURL(&vestingtypes.MsgCreatePeriodicVestingAccount{}) {
			continue
		}

		// Create an encoded json msg, then unmarshal it to instantiate the msg type.
		jsonMsg := []byte(fmt.Sprintf(`{"@type": "%s"}`, msgName))

		var msg sdk.Msg
		err := protoCodec.UnmarshalInterfaceJSON(jsonMsg, &msg)
		require.NoError(t, err)

		// Only check legacy msgs for amino registration.
		// Only legacy msg can be signed with amino.
		_, ok := msg.(legacytx.LegacyMsg)
		if !ok {
			continue
		}

		// Check the msg is registered in amino by checking a repeat registration call panics.
		panicValue, ok := catchPanic(func() {
			lcdc.RegisterConcrete(interface{}(msg), fmt.Sprintf("aUniqueRegistrationName%d", i), nil)
		})
		assert.True(t, ok, "registration did not panic, msg %s is not registered in amino", msgName)
		if ok {
			require.IsTypef(t, "", panicValue, "msg %s amino registration panicked with unexpected type", msgName)
			aminoErrMsgPrefix := "TypeInfo already exists for"
			require.Containsf(t, panicValue, aminoErrMsgPrefix, "msg %s amino registration panicked for unexpected reason", msgName)
		}
	}
}

// catchPanic returns the panic value of the passed function. The second return indicates if the function panicked.
func catchPanic(f func()) (panicValue interface{}, didPanic bool) {
	didPanic = true

	defer func() {
		panicValue = recover()
	}()

	f()
	didPanic = false
	return
}

// unmarshalJSONKeys extracts keys from the top level of a json blob.
func unmarshalJSONKeys(jsonBytes []byte) ([]string, error) {
	var jsonMap map[string]json.RawMessage
	err := json.Unmarshal(jsonBytes, &jsonMap)
	if err != nil {
		return nil, err
	}

	keys := make([]string, 0, len(jsonMap))
	for k := range jsonMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys, nil
}
