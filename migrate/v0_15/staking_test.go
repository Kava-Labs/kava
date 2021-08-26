package v0_15

import (
	"path/filepath"
	"testing"

	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/kava-labs/kava/app"
	"github.com/stretchr/testify/require"
	tmtypes "github.com/tendermint/tendermint/types"
)

func TestStaking_Full(t *testing.T) {
	t.Skip() // skip to avoid having to commit a large genesis file to the repo

	genDoc, err := tmtypes.GenesisDocFromFile(filepath.Join("testdata", "genesis.json"))
	require.NoError(t, err)

	cdc := makeV014Codec()

	var oldState genutil.AppMap
	cdc.MustUnmarshalJSON(genDoc.AppState, &oldState)

	var oldStakingGenState staking.GenesisState
	cdc.MustUnmarshalJSON(oldState[staking.ModuleName], &oldStakingGenState)

	var oldSlashingGenState slashing.GenesisState
	cdc.MustUnmarshalJSON(oldState[slashing.ModuleName], &oldSlashingGenState)

	var oldDistributionGenState distribution.GenesisState
	cdc.MustUnmarshalJSON(oldState[distribution.ModuleName], &oldDistributionGenState)

	newStakingGenState, newDistributionGenState, newSlashingGenState := Staking(app.MakeCodec(), oldStakingGenState, oldDistributionGenState, oldSlashingGenState, ValidatorKeysDir)
	require.NoError(t, staking.ValidateGenesis(newStakingGenState))
	require.NoError(t, distribution.ValidateGenesis(newDistributionGenState))
	require.NoError(t, slashing.ValidateGenesis(newSlashingGenState))

}
