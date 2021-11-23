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
	original := GetTestDataJSON("appstate-gov-v15.json")
	expected := GetTestDataJSON("appstate-gov-v16.json")
	actual := MustMigrateAppStateJSON(original)
	assert.JSONEq(t, expected, actual)
}

func TestCosmosMigrate_Bank(t *testing.T) {
	// The bank json tests migrating the bank, auth, and supply modules
	// The json contains the kava ValidatorVestingAccount account and tests for
	// both the correct proto migration & moving account coins and supply to the bank module.
	original := GetTestDataJSON("appstate-bank-v15.json")
	expected := GetTestDataJSON("appstate-bank-v16.json")
	actual := MustMigrateAppStateJSON(original)
	assert.JSONEq(t, expected, actual)
}

func TestCosmosMigrate_Distribution(t *testing.T) {
	original := GetTestDataJSON("appstate-distribution-v15.json")
	expected := GetTestDataJSON("appstate-distribution-v16.json")
	actual := MustMigrateAppStateJSON(original)
	assert.JSONEq(t, expected, actual)
}

func TestCosmosMigrate_Staking(t *testing.T) {
	original := GetTestDataJSON("appstate-staking-v15.json")
	expected := GetTestDataJSON("appstate-staking-v16.json")
	actual := MustMigrateAppStateJSON(original)
	assert.JSONEq(t, expected, actual)
}

func TestCosmosMigrate_Modules(t *testing.T) {
	original := GetTestDataJSON("appstate-cosmos-v15.json")
	expected := GetTestDataJSON("appstate-cosmos-v16.json")
	actual := MustMigrateAppStateJSON(original)
	assert.JSONEq(t, expected, actual)
}

// MustMigrateAppStateJSON migrate v15 app state json to v16
func MustMigrateAppStateJSON(appStateJson string) string {
	var appState genutiltypes.AppMap
	if err := json.Unmarshal([]byte(appStateJson), &appState); err != nil {
		panic(err)
	}
	newGenState := migrateCosmosAppState(appState, NewClientContext())
	actual, err := json.Marshal(newGenState)
	if err != nil {
		panic(err)
	}
	return string(actual)
}

func GetTestDataJSON(filename string) string {
	file := filepath.Join("testdata", filename)
	data, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func NewClientContext() client.Context {
	config := app.MakeEncodingConfig()
	return client.Context{}.
		WithCodec(config.Marshaler).
		WithLegacyAmino(config.Amino)
}
