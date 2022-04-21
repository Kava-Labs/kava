package v0_17

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cosmos/cosmos-sdk/client"

	tmjson "github.com/tendermint/tendermint/libs/json"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/kava-labs/kava/app"
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

	assert.LessOrEqual(t, actualGenDoc.ConsensusParams.Evidence.MaxBytes, actualGenDoc.ConsensusParams.Block.MaxBytes)
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
