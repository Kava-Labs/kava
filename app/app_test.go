package app

import (
	"encoding/json"
	"os"
	"sort"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"
	db "github.com/tendermint/tm-db"
	ethermint "github.com/tharsis/ethermint/types"
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

	stateBytes, err := json.Marshal(NewDefaultGenesisState())
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
	assert.Equal(t, []tmtypes.GenesisValidator(nil), exportedApp.Validators) // no validators set in default genesis
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

func TestAsyncUpgrade_DefaultAccountType(t *testing.T) {
	_, addrs := GeneratePrivKeyAddressPairs(3)

	tApp := NewTestApp()
	tApp.InitializeFromGenesisStates(
		NewAuthBankGenesisBuilder().
			WithSimpleAccount(addrs[0], sdk.NewCoins(sdk.NewInt64Coin("ukava", 1e9))).
			BuildMarshalled(tApp.AppCodec()),
	)
	accountKeeper := tApp.GetAccountKeeper()
	bankKeeper := tApp.GetBankKeeper()

	// create a default account before upgrade height and check it's an eth account
	ctx := tApp.NewContext(false, tmproto.Header{Height: 1})

	bankKeeper.SendCoins(ctx, addrs[0], addrs[1], sdk.NewCoins(sdk.NewInt64Coin("ukava", 1e6)))
	acc := accountKeeper.GetAccount(ctx, addrs[1])
	_, ok := acc.(*ethermint.EthAccount)
	require.True(t, ok)

	// create a default account after upgrade height and check it's a base account
	ctx = tApp.NewContext(false, tmproto.Header{Height: FixDefaultAccountUpgradeHeight})

	bankKeeper.SendCoins(ctx, addrs[0], addrs[2], sdk.NewCoins(sdk.NewInt64Coin("ukava", 1e6)))
	acc = accountKeeper.GetAccount(ctx, addrs[2])
	_, ok = acc.(*authtypes.BaseAccount)
	require.True(t, ok)
}
