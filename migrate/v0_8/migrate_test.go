package v0_8

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	v032tendermint "github.com/kava-labs/kava/migrate/v0_8/tendermint/v0_32"
	v033tendermint "github.com/kava-labs/kava/migrate/v0_8/tendermint/v0_33"
	"github.com/stretchr/testify/require"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/kava-labs/kava/app"
)

func TestMain(m *testing.M) {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)
	app.SetBip44CoinType(config)

	os.Exit(m.Run())
}

func TestMigrate_Auth_BaseAccount(t *testing.T) {
	bz, err := ioutil.ReadFile(filepath.Join("testdata", "auth-base-old.json"))
	require.NoError(t, err)
	oldAppState := genutil.AppMap{"auth": bz}

	newAppState := MigrateAppState(oldAppState)

	bz, err = ioutil.ReadFile(filepath.Join("testdata", "auth-base-new.json"))
	require.NoError(t, err)
	require.JSONEq(t, string(bz), string(newAppState["auth"]))
}
func TestMigrate_Auth_MultiSigAccount(t *testing.T) {
	bz, err := ioutil.ReadFile(filepath.Join("testdata", "auth-base-multisig-old.json"))
	require.NoError(t, err)
	oldAppState := genutil.AppMap{"auth": bz}

	newAppState := MigrateAppState(oldAppState)

	bz, err = ioutil.ReadFile(filepath.Join("testdata", "auth-base-multisig-new.json"))
	require.NoError(t, err)
	require.JSONEq(t, string(bz), string(newAppState["auth"]))
}

func TestMigrate_Auth_ValidatorVestingAccount(t *testing.T) {
	bz, err := ioutil.ReadFile(filepath.Join("testdata", "auth-valvesting-old.json"))
	require.NoError(t, err)
	oldAppState := genutil.AppMap{"auth": bz}

	newAppState := MigrateAppState(oldAppState)

	bz, err = ioutil.ReadFile(filepath.Join("testdata", "auth-valvesting-new.json"))
	require.NoError(t, err)
	require.JSONEq(t, string(bz), string(newAppState["auth"]))
}

func TestMigrate_Auth_ModuleAccount(t *testing.T) {
	bz, err := ioutil.ReadFile(filepath.Join("testdata", "auth-module-old.json"))
	require.NoError(t, err)
	oldAppState := genutil.AppMap{"auth": bz}

	newAppState := MigrateAppState(oldAppState)

	bz, err = ioutil.ReadFile(filepath.Join("testdata", "auth-module-new.json"))
	require.NoError(t, err)
	require.JSONEq(t, string(bz), string(newAppState["auth"]))
}

func TestMigrate_Auth_PeriodicVestingAccount(t *testing.T) {
	bz, err := ioutil.ReadFile(filepath.Join("testdata", "auth-periodic-old.json"))
	require.NoError(t, err)
	oldAppState := genutil.AppMap{"auth": bz}

	newAppState := MigrateAppState(oldAppState)

	bz, err = ioutil.ReadFile(filepath.Join("testdata", "auth-periodic-new.json"))
	require.NoError(t, err)
	require.JSONEq(t, string(bz), string(newAppState["auth"]))
}

func TestMigrate_Distribution(t *testing.T) {
	bz, err := ioutil.ReadFile(filepath.Join("testdata", "distribution-old.json"))
	require.NoError(t, err)
	oldAppState := genutil.AppMap{"distribution": bz}

	newAppState := MigrateSDK(oldAppState)

	bz, err = ioutil.ReadFile(filepath.Join("testdata", "distribution-new.json"))
	require.NoError(t, err)
	require.JSONEq(t, string(bz), string(newAppState["distribution"]))
}

func TestMigrate_Staking(t *testing.T) {
	bz, err := ioutil.ReadFile(filepath.Join("testdata", "staking-old.json"))
	require.NoError(t, err)
	oldAppState := genutil.AppMap{"staking": bz}

	newAppState := MigrateSDK(oldAppState)

	bz, err = ioutil.ReadFile(filepath.Join("testdata", "staking-new.json"))
	require.NoError(t, err)
	require.JSONEq(t, string(bz), string(newAppState["staking"]))
}

func TestMigrate_Slashing(t *testing.T) {
	bz, err := ioutil.ReadFile(filepath.Join("testdata", "slashing-old.json"))
	require.NoError(t, err)
	oldAppState := genutil.AppMap{"slashing": bz}

	newAppState := MigrateSDK(oldAppState)

	bz, err = ioutil.ReadFile(filepath.Join("testdata", "slashing-new.json"))
	require.NoError(t, err)
	require.JSONEq(t, string(bz), string(newAppState["slashing"]))
}

func TestMigrate_Evidence(t *testing.T) {
	bz, err := ioutil.ReadFile(filepath.Join("testdata", "slashing-old.json"))
	require.NoError(t, err)
	oldAppState := genutil.AppMap{"slashing": bz}

	newAppState := MigrateSDK(oldAppState)

	bz, err = ioutil.ReadFile(filepath.Join("testdata", "evidence-new.json"))
	require.NoError(t, err)
	require.JSONEq(t, string(bz), string(newAppState["evidence"]))
}

func TestMigrate_Tendermint(t *testing.T) {
	oldGenDoc, err := v032tendermint.GenesisDocFromFile(filepath.Join("testdata", "tendermint-old.json"))
	require.NoError(t, err)

	newGenDoc := v033tendermint.Migrate(*oldGenDoc)

	expectedGenDoc, err := tmtypes.GenesisDocFromFile(filepath.Join("testdata", "tendermint-new.json"))
	require.NoError(t, err)
	require.Equal(t, *expectedGenDoc, newGenDoc)
}

func TestMigrate(t *testing.T) {
	oldGenDoc, err := v032tendermint.GenesisDocFromFile(filepath.Join("testdata", "all-old.json"))
	require.NoError(t, err)

	newGenDoc := Migrate(*oldGenDoc)

	expectedGenDoc, err := tmtypes.GenesisDocFromFile(filepath.Join("testdata", "all-new.json"))
	require.NoError(t, err)
	// check each field seperately to aid debugging
	require.Equal(t, expectedGenDoc.AppHash, newGenDoc.AppHash)
	require.JSONEq(t, string(expectedGenDoc.AppState), string(newGenDoc.AppState))
	require.Equal(t, expectedGenDoc.ChainID, newGenDoc.ChainID)
	require.Equal(t, expectedGenDoc.ConsensusParams, newGenDoc.ConsensusParams)
	require.Equal(t, expectedGenDoc.GenesisTime, newGenDoc.GenesisTime)
	require.Equal(t, expectedGenDoc.Validators, newGenDoc.Validators)

	var newAppState genutil.AppMap
	require.NoError(t,
		app.MakeCodec().UnmarshalJSON(newGenDoc.AppState, &newAppState),
	)
	require.NoError(t,
		app.ModuleBasics.ValidateGenesis(newAppState),
	)
}

func TestMigrate_Full(t *testing.T) {
	// 1) load an exported kava-2 state
	oldGenDoc, err := v032tendermint.GenesisDocFromFile(filepath.Join("testdata", "kava-2.json"))
	require.NoError(t, err)
	tApp := app.NewTestApp() // also sets the bech32 prefix on sdk.Config
	cdc := app.MakeCodec()

	// 2) migrate
	newGenDoc := Migrate(*oldGenDoc)

	// 3) check new genesis is valid
	var newAppState genutil.AppMap
	require.NoError(t,
		cdc.UnmarshalJSON(newGenDoc.AppState, &newAppState),
	)
	require.NoError(t,
		app.ModuleBasics.ValidateGenesis(newAppState),
	)
	require.NotPanics(t, func() {
		// this runs both InitGenesis for all modules (which panic on errors) and runs all invariants
		tApp.InitializeFromGenesisStates(app.GenesisState(newAppState))
	})
}
