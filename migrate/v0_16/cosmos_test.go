package v0_16

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/stretchr/testify/assert"

	"github.com/kava-labs/kava/app"
)

func TestCosmosMigrate_Gov(t *testing.T) {
	// The gov json contains a gov app state with 3 different proposals.
	// Two of the proposals are from cosmos while the 3rd one is kavadist/CommunityPoolMultiSpendProposal.
	original := getTestDataJSON("appstate-gov-v15.json")
	expected := getTestDataJSON("appstate-gov-v16.json")
	actual := mustMigrateCosmosAppStateJSON(original)
	assert.JSONEq(t, expected, actual)
}

func TestCosmosMigrate_Bank(t *testing.T) {
	// The bank json tests migrating the bank, auth, and supply modules
	// The json contains the kava ValidatorVestingAccount account and tests for
	// both the correct proto migration & moving account coins and supply to the bank module.
	original := getTestDataJSON("appstate-bank-v15.json")
	expected := getTestDataJSON("appstate-bank-v16.json")
	actual := mustMigrateCosmosAppStateJSON(original)
	assert.JSONEq(t, expected, actual)
}

func TestCosmosMigrate_Distribution(t *testing.T) {
	original := getTestDataJSON("appstate-distribution-v15.json")
	expected := getTestDataJSON("appstate-distribution-v16.json")
	actual := mustMigrateCosmosAppStateJSON(original)
	assert.JSONEq(t, expected, actual)
}

func TestCosmosMigrate_Staking(t *testing.T) {
	original := getTestDataJSON("appstate-staking-v15.json")
	expected := getTestDataJSON("appstate-staking-v16.json")
	actual := mustMigrateCosmosAppStateJSON(original)
	assert.JSONEq(t, expected, actual)
}

func TestCosmosMigrate_Modules(t *testing.T) {
	original := getTestDataJSON("appstate-cosmos-v15.json")
	expected := getTestDataJSON("appstate-cosmos-v16.json")
	actual := mustMigrateCosmosAppStateJSON(original)
	assert.JSONEq(t, expected, actual)
}

// mustMigrateCosmosAppStateJSON migrate v15 app state json to v16
func mustMigrateCosmosAppStateJSON(appStateJson string) string {
	var appState genutiltypes.AppMap
	if err := json.Unmarshal([]byte(appStateJson), &appState); err != nil {
		panic(err)
	}
	appState = migrateV040(appState, newClientContext())
	newGenState := migrateV043(appState, newClientContext())
	actual, err := json.Marshal(newGenState)
	if err != nil {
		panic(err)
	}
	return string(actual)
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
		WithLegacyAmino(config.Amino)
}
