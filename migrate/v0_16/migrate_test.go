package v0_16

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	tmjson "github.com/tendermint/tendermint/libs/json"
	tmtypes "github.com/tendermint/tendermint/types"
)

func TestMigrateGenesisDoc(t *testing.T) {
	// The gov json contains a gov app state with 3 different proposals.
	// Two of the proposals are from cosmos while the 3rd one is kavadist/CommunityPoolMultiSpendProposal.
	expected := getTestDataJSON("genesis-v16.json")
	genDoc, err := tmtypes.GenesisDocFromFile(filepath.Join("testdata", "genesis-v15.json"))
	assert.NoError(t, err)
	actualGenDoc, err := Migrate(genDoc, newClientContext())
	assert.NoError(t, err)
	actualJson, err := tmjson.Marshal(actualGenDoc)
	assert.NoError(t, err)
	assert.JSONEq(t, expected, string(actualJson))
}
