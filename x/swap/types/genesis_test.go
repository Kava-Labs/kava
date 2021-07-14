package types_test

import (
	"encoding/json"
	"testing"

	"github.com/kava-labs/kava/x/swap/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
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

	genesisA := types.GenesisState{params, types.DefaultPoolRecords, types.DefaultShareRecords}
	genesisB := types.GenesisState{params, types.DefaultPoolRecords, types.DefaultShareRecords}

	assert.True(t, genesisA.Equal(genesisB))
}

func TestGenesis_NotEqual(t *testing.T) {
	baseParams := types.Params{
		types.NewAllowedPools(types.NewAllowedPool("ukava", "usdx")),
		sdk.MustNewDecFromStr("0.85"),
	}

	// Base params
	genesisAParams := baseParams
	genesisA := types.GenesisState{genesisAParams, types.DefaultPoolRecords, types.DefaultShareRecords}

	// Different swap fee
	genesisBParams := baseParams
	genesisBParams.SwapFee = sdk.MustNewDecFromStr("0.84")
	genesisB := types.GenesisState{genesisBParams, types.DefaultPoolRecords, types.DefaultShareRecords}

	// Different pairs
	genesisCParams := baseParams
	genesisCParams.AllowedPools = types.NewAllowedPools(types.NewAllowedPool("ukava", "hard"))
	genesisC := types.GenesisState{genesisCParams, types.DefaultPoolRecords, types.DefaultShareRecords}

	// A and B have different swap fees
	assert.False(t, genesisA.Equal(genesisB))
	// A and C have different pair token B denoms
	assert.False(t, genesisA.Equal(genesisC))
	// A and B and different swap fees and pair token B denoms
	assert.False(t, genesisA.Equal(genesisB))
}

func TestGenesis_JSONEncoding(t *testing.T) {
	raw := `{
    "params": {
			"allowed_pools": [
			  {
			    "token_a": "ukava",
					"token_b": "usdx"
				},
			  {
			    "token_a": "hard",
					"token_b": "busd"
				}
			],
			"swap_fee": "0.003000000000000000"
		},
		"pool_records": [
		  {
			  "pool_id": "ukava/usdx",
			  "reserves_a": { "denom": "ukava", "amount": "1000000" },
			  "reserves_b": { "denom": "usdx", "amount": "5000000" },
			  "total_shares": "3000000"
			},
		  {
			  "pool_id": "hard/usdx",
			  "reserves_a": { "denom": "ukava", "amount": "1000000" },
			  "reserves_b": { "denom": "usdx", "amount": "2000000" },
			  "total_shares": "2000000"
			}
		],
		"share_records": [
		  {
		    "depositor": "kava1mq9qxlhze029lm0frzw2xr6hem8c3k9ts54w0w",
		    "pool_id": "ukava/usdx",
		    "shares_owned": "100000"
			},
		  {
		    "depositor": "kava1esagqd83rhqdtpy5sxhklaxgn58k2m3s3mnpea",
		    "pool_id": "hard/usdx",
		    "shares_owned": "200000"
			}
		]
	}`

	var state types.GenesisState
	err := json.Unmarshal([]byte(raw), &state)
	require.NoError(t, err)

	assert.Equal(t, 2, len(state.Params.AllowedPools))
	assert.Equal(t, sdk.MustNewDecFromStr("0.003"), state.Params.SwapFee)
	assert.Equal(t, 2, len(state.PoolRecords))
	assert.Equal(t, 2, len(state.ShareRecords))
}

func TestGenesis_YAMLEncoding(t *testing.T) {
	expected := `params:
  allowed_pools:
  - token_a: ukava
    token_b: usdx
  - token_a: hard
    token_b: busd
  swap_fee: "0.003000000000000000"
pool_records:
- pool_id: ukava/usdx
  reserves_a:
    denom: ukava
    amount: "1000000"
  reserves_b:
    denom: usdx
    amount: "5000000"
  total_shares: "3000000"
- pool_id: hard/usdx
  reserves_a:
    denom: hard
    amount: "1000000"
  reserves_b:
    denom: usdx
    amount: "2000000"
  total_shares: "1500000"
share_records:
- depositor: kava1mq9qxlhze029lm0frzw2xr6hem8c3k9ts54w0w
  pool_id: ukava/usdx
  shares_owned: "100000"
- depositor: kava1esagqd83rhqdtpy5sxhklaxgn58k2m3s3mnpea
  pool_id: hard/usdx
  shares_owned: "200000"
`

	depositor_1, err := sdk.AccAddressFromBech32("kava1mq9qxlhze029lm0frzw2xr6hem8c3k9ts54w0w")
	require.NoError(t, err)
	depositor_2, err := sdk.AccAddressFromBech32("kava1esagqd83rhqdtpy5sxhklaxgn58k2m3s3mnpea")
	require.NoError(t, err)

	state := types.NewGenesisState(
		types.NewParams(
			types.NewAllowedPools(
				types.NewAllowedPool("ukava", "usdx"),
				types.NewAllowedPool("hard", "busd"),
			),
			sdk.MustNewDecFromStr("0.003"),
		),
		types.PoolRecords{
			types.NewPoolRecord(sdk.NewCoins(ukava(1e6), usdx(5e6)), i(3e6)),
			types.NewPoolRecord(sdk.NewCoins(hard(1e6), usdx(2e6)), i(15e5)),
		},
		types.ShareRecords{
			types.NewShareRecord(depositor_1, "ukava/usdx", i(1e5)),
			types.NewShareRecord(depositor_2, "hard/usdx", i(2e5)),
		},
	)

	data, err := yaml.Marshal(state)
	require.NoError(t, err)

	assert.Equal(t, expected, string(data))
}
