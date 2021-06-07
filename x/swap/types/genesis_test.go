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
			AllowedPools: types.NewAllowedPools(types.NewAllowedPool("ukava", "hard")),
			SwapFee:      sdk.ZeroDec(),
		},
	}
	assert.False(t, nonEmptyGenesis.IsEmpty())
}

func TestGenesis_Validate_SwapFee(t *testing.T) {
	type args struct {
		name      string
		swapFee   sdk.Dec
		expectErr bool
	}
	// More comprehensive swap fee tests are in prams_test.go
	testCases := []args{
		{
			"normal",
			sdk.MustNewDecFromStr("0.25"),
			false,
		},
		{
			"negative",
			sdk.MustNewDecFromStr("-0.5"),
			true,
		},
		{
			"greater than 1.0",
			sdk.MustNewDecFromStr("1.001"),
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			genesisState := types.GenesisState{
				Params: types.Params{
					AllowedPools: types.DefaultAllowedPools,
					SwapFee:      tc.swapFee,
				},
			}

			err := genesisState.Validate()
			if tc.expectErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestGenesis_Validate_AllowedPools(t *testing.T) {
	type args struct {
		name      string
		pairs     types.AllowedPools
		expectErr bool
	}
	// More comprehensive pair validation tests are in pair_test.go, params_test.go
	testCases := []args{
		{
			"normal",
			types.DefaultAllowedPools,
			false,
		},
		{
			"invalid",
			types.AllowedPools{
				{
					TokenA: "same",
					TokenB: "same",
				},
			},
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			genesisState := types.GenesisState{
				Params: types.Params{
					AllowedPools: tc.pairs,
					SwapFee:      types.DefaultSwapFee,
				},
			}

			err := genesisState.Validate()
			if tc.expectErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestGenesis_Equal(t *testing.T) {
	params := types.Params{
		types.NewAllowedPools(types.NewAllowedPool("ukava", "usdx")),
		sdk.MustNewDecFromStr("0.85"),
	}

	genesisA := types.GenesisState{params}
	genesisB := types.GenesisState{params}

	assert.True(t, genesisA.Equal(genesisB))
}

func TestGenesis_NotEqual(t *testing.T) {
	baseParams := types.Params{
		types.NewAllowedPools(types.NewAllowedPool("ukava", "usdx")),
		sdk.MustNewDecFromStr("0.85"),
	}

	// Base params
	genesisAParams := baseParams
	genesisA := types.GenesisState{genesisAParams}

	// Different swap fee
	genesisBParams := baseParams
	genesisBParams.SwapFee = sdk.MustNewDecFromStr("0.84")
	genesisB := types.GenesisState{genesisBParams}

	// Different pairs
	genesisCParams := baseParams
	genesisCParams.AllowedPools = types.NewAllowedPools(types.NewAllowedPool("ukava", "hard"))
	genesisC := types.GenesisState{genesisCParams}

	// A and B have different swap fees
	assert.False(t, genesisA.Equal(genesisB))
	// A and C have different pair token B denoms
	assert.False(t, genesisA.Equal(genesisC))
	// A and B and different swap fees and pair token B denoms
	assert.False(t, genesisA.Equal(genesisB))
}
