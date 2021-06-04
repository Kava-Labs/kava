package types_test

import (
	"testing"

	"github.com/kava-labs/kava/x/swap/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenesis_Default(t *testing.T) {
	defaultGenesis := types.DefaultGenesisState()

	require.NoError(t, defaultGenesis.Validate())

	defaultParams := types.DefaultParams()
	assert.Equal(t, defaultParams, defaultGenesis.Params)
}

func TestGenesis_Empty(t *testing.T) {
	emptyGenesis := types.GenesisState{}
	assert.True(t, emptyGenesis.IsEmpty())

	emptyGenesis = types.GenesisState{
		Params: types.Params{},
	}
	assert.True(t, emptyGenesis.IsEmpty())
}

func TestGenesis_NotEmpty(t *testing.T) {
	nonEmptyGenesis := types.GenesisState{
		Params: types.Params{
			Pairs:   types.NewPairs(types.NewPair("ukava", "hard")),
			SwapFee: sdk.ZeroDec(),
		},
	}
	assert.False(t, nonEmptyGenesis.IsEmpty())
}
