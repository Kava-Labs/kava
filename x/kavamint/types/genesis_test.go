package types_test

import (
	"encoding/json"
	"testing"

	time "time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"

	"github.com/kava-labs/kava/x/kavamint/types"
)

func TestGenesis_Default(t *testing.T) {
	defaultGenesis := types.DefaultGenesisState()

	require.NoError(t, defaultGenesis.Validate())

	defaultParams := types.DefaultParams()
	assert.Equal(t, defaultParams, defaultGenesis.Params)
}

func TestGenesis_Validation(t *testing.T) {
	testCases := []struct {
		name        string
		gs          *types.GenesisState
		expectedErr string
	}{
		{
			"valid - default genesis is valid",
			types.DefaultGenesisState(),
			"",
		},
		{
			"valid - valid genesis",
			types.NewGenesisState(
				newValidParams(t),
				time.Now(),
			),
			"",
		},
		{
			"valid - no inflation",
			types.NewGenesisState(
				types.NewParams(sdk.ZeroDec(), sdk.ZeroDec()),
				time.Now(),
			),
			"",
		},
		{
			"valid - no time set",
			types.NewGenesisState(
				newValidParams(t),
				time.Time{},
			),
			"",
		},
		{
			"invalid - community inflation param too big",
			types.NewGenesisState(
				types.NewParams(types.MaxMintingRate.Add(sdk.NewDecWithPrec(1, 18)), sdk.ZeroDec()),
				time.Now(),
			),
			"invalid rate: 100.000000000000000001",
		},
		{
			"invalid - staking reward inflation param too big",
			types.NewGenesisState(
				// inflation is larger than is allowed!
				types.NewParams(sdk.ZeroDec(), types.MaxMintingRate.Add(sdk.NewDecWithPrec(1, 18))),
				time.Now(),
			),
			"invalid rate: 100.000000000000000001",
		},
		{
			"invalid - negative community inflation param",
			types.NewGenesisState(
				types.NewParams(sdk.OneDec().MulInt64(-1), sdk.ZeroDec()),
				time.Now(),
			),
			"invalid rate: -1.000000000000000000",
		},
		{
			"invalid - negative staking inflation param",
			types.NewGenesisState(
				types.NewParams(sdk.ZeroDec(), sdk.OneDec().MulInt64(-1)),
				time.Now(),
			),
			"invalid rate: -1.000000000000000000",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.gs.Validate()

			if tc.expectedErr == "" {
				assert.Nil(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestGenesis_JSONEncoding(t *testing.T) {
	raw := `{
    "params": {
			"community_pool_inflation": "0.000000000000000001",
			"staking_rewards_apy": "0.000000000000000002"
		},
		"previous_block_time": "2022-12-01T11:30:55.000000001Z"
	}`

	var state types.GenesisState
	err := json.Unmarshal([]byte(raw), &state)
	require.NoError(t, err)

	assert.Equal(t, sdk.MustNewDecFromStr("0.000000000000000001"), state.Params.CommunityPoolInflation)
	assert.Equal(t, sdk.MustNewDecFromStr("0.000000000000000002"), state.Params.StakingRewardsApy)

	prevBlockTime, err := time.Parse(time.RFC3339, "2022-12-01T11:30:55.000000001Z")
	require.NoError(t, err)

	assert.Equal(t, prevBlockTime, state.PreviousBlockTime)
}

func TestGenesis_YAMLEncoding(t *testing.T) {
	expected := `params:
  community_pool_inflation: "0.000000000000000001"
  staking_rewards_apy: "0.000000000000000002"
previous_block_time: "2022-12-01T11:30:55.000000001Z"
`
	prevBlockTime, err := time.Parse(time.RFC3339, "2022-12-01T11:30:55.000000001Z")
	require.NoError(t, err)

	state := types.NewGenesisState(
		types.NewParams(sdk.MustNewDecFromStr("0.000000000000000001"), sdk.MustNewDecFromStr("0.000000000000000002")),
		prevBlockTime,
	)

	data, err := yaml.Marshal(state)
	require.NoError(t, err)

	assert.Equal(t, expected, string(data))
}
