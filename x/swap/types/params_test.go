package types_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/kava-labs/kava/x/swap/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestParams_UnmarshalJSON(t *testing.T) {
	pools := types.NewAllowedPools(
		types.NewAllowedPool("hard", "ukava"),
		types.NewAllowedPool("hard", "usdx"),
	)
	poolData, err := json.Marshal(pools)
	require.NoError(t, err)

	fee, err := sdk.NewDecFromStr("0.5")
	require.NoError(t, err)
	feeData, err := json.Marshal(fee)
	require.NoError(t, err)

	data := fmt.Sprintf(`{
	"allowed_pools": %s,
	"swap_fee": %s
}`, string(poolData), string(feeData))

	var params types.Params
	err = json.Unmarshal([]byte(data), &params)
	require.NoError(t, err)

	assert.Equal(t, pools, params.AllowedPools)
	assert.Equal(t, fee, params.SwapFee)
}

func TestParams_MarshalYAML(t *testing.T) {
	pools := types.NewAllowedPools(
		types.NewAllowedPool("hard", "ukava"),
		types.NewAllowedPool("hard", "usdx"),
	)
	fee, err := sdk.NewDecFromStr("0.5")
	require.NoError(t, err)

	p := types.Params{
		AllowedPools: pools,
		SwapFee:      fee,
	}

	data, err := yaml.Marshal(p)
	require.NoError(t, err)

	var params map[string]interface{}
	err = yaml.Unmarshal(data, &params)
	require.NoError(t, err)

	_, ok := params["allowed_pools"]
	require.True(t, ok)
	_, ok = params["swap_fee"]
	require.True(t, ok)
}

func TestParams_Default(t *testing.T) {
	defaultParams := types.DefaultParams()

	require.NoError(t, defaultParams.Validate())

	assert.Equal(t, types.DefaultAllowedPools, defaultParams.AllowedPools)
	assert.Equal(t, types.DefaultSwapFee, defaultParams.SwapFee)

	assert.Equal(t, 0, len(defaultParams.AllowedPools))
	assert.Equal(t, sdk.ZeroDec(), defaultParams.SwapFee)
}

func TestParams_ParamSetPairs_AllowedPools(t *testing.T) {
	assert.Equal(t, []byte("AllowedPools"), types.KeyAllowedPools)
	defaultParams := types.DefaultParams()

	var paramSetPair *paramstypes.ParamSetPair
	for _, pair := range defaultParams.ParamSetPairs() {
		if bytes.Equal(pair.Key, types.KeyAllowedPools) {
			paramSetPair = &pair
			break
		}
	}
	require.NotNil(t, paramSetPair)

	pairs, ok := paramSetPair.Value.(*types.AllowedPools)
	require.True(t, ok)
	assert.Equal(t, pairs, &defaultParams.AllowedPools)

	assert.Nil(t, paramSetPair.ValidatorFn(*pairs))
	assert.EqualError(t, paramSetPair.ValidatorFn(struct{}{}), "invalid parameter type: struct {}")
}

func TestParams_ParamSetPairs_SwapFee(t *testing.T) {
	assert.Equal(t, []byte("SwapFee"), types.KeySwapFee)
	defaultParams := types.DefaultParams()

	var paramSetPair *paramstypes.ParamSetPair
	for _, pair := range defaultParams.ParamSetPairs() {
		if bytes.Equal(pair.Key, types.KeySwapFee) {
			paramSetPair = &pair
			break
		}
	}
	require.NotNil(t, paramSetPair)

	swapFee, ok := paramSetPair.Value.(*sdk.Dec)
	require.True(t, ok)
	assert.Equal(t, swapFee, &defaultParams.SwapFee)

	assert.Nil(t, paramSetPair.ValidatorFn(*swapFee))
	assert.EqualError(t, paramSetPair.ValidatorFn(struct{}{}), "invalid parameter type: struct {}")
}

func TestParams_Validation(t *testing.T) {
	testCases := []struct {
		name        string
		key         []byte
		testFn      func(params *types.Params)
		expectedErr string
	}{
		{
			name: "invalid denom",
			key:  types.KeyAllowedPools,
			testFn: func(params *types.Params) {
				params.AllowedPools = types.NewAllowedPools(types.NewAllowedPool("UKAVA", "ukava"))
			},
			expectedErr: "invalid denom: UKAVA",
		},
		{
			name: "duplicate pools",
			key:  types.KeyAllowedPools,
			testFn: func(params *types.Params) {
				params.AllowedPools = types.NewAllowedPools(types.NewAllowedPool("ukava", "ukava"))
			},
			expectedErr: "pool cannot have two tokens of the same type, received 'ukava' and 'ukava'",
		},
		{
			name: "nil swap fee",
			key:  types.KeySwapFee,
			testFn: func(params *types.Params) {
				params.SwapFee = sdk.Dec{}
			},
			expectedErr: "invalid swap fee: <nil>",
		},
		{
			name: "negative swap fee",
			key:  types.KeySwapFee,
			testFn: func(params *types.Params) {
				params.SwapFee = sdk.NewDec(-1)
			},
			expectedErr: "invalid swap fee: -1.000000000000000000",
		},
		{
			name: "swap fee greater than 1",
			key:  types.KeySwapFee,
			testFn: func(params *types.Params) {
				params.SwapFee = sdk.MustNewDecFromStr("1.000000000000000001")
			},
			expectedErr: "invalid swap fee: 1.000000000000000001",
		},
		{
			name: "0 swap fee",
			key:  types.KeySwapFee,
			testFn: func(params *types.Params) {
				params.SwapFee = sdk.ZeroDec()
			},
			expectedErr: "",
		},
		{
			name: "1 swap fee",
			key:  types.KeySwapFee,
			testFn: func(params *types.Params) {
				params.SwapFee = sdk.OneDec()
			},
			expectedErr: "invalid swap fee: 1.000000000000000000",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			params := types.DefaultParams()
			tc.testFn(&params)

			err := params.Validate()

			if tc.expectedErr == "" {
				assert.Nil(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedErr)
			}

			var paramSetPair *paramstypes.ParamSetPair
			for _, pair := range params.ParamSetPairs() {
				if bytes.Equal(pair.Key, tc.key) {
					paramSetPair = &pair
					break
				}
			}
			require.NotNil(t, paramSetPair)
			value := reflect.ValueOf(paramSetPair.Value).Elem().Interface()

			// assert validation error is same as param set validation
			assert.Equal(t, err, paramSetPair.ValidatorFn(value))
		})
	}
}

func TestParams_String(t *testing.T) {
	params := types.NewParams(
		types.NewAllowedPools(
			types.NewAllowedPool("hard", "ukava"),
			types.NewAllowedPool("ukava", "usdx"),
		),
		sdk.MustNewDecFromStr("0.5"),
	)
	require.NoError(t, params.Validate())

	output := params.String()
	assert.Contains(t, output, types.PoolID("hard", "ukava"))
	assert.Contains(t, output, types.PoolID("ukava", "usdx"))
	assert.Contains(t, output, "0.5")
}

func TestAllowedPool_Validation(t *testing.T) {
	testCases := []struct {
		name        string
		allowedPool types.AllowedPool
		expectedErr string
	}{
		{
			name:        "blank token a",
			allowedPool: types.NewAllowedPool("", "ukava"),
			expectedErr: "invalid denom: ",
		},
		{
			name:        "blank token b",
			allowedPool: types.NewAllowedPool("ukava", ""),
			expectedErr: "invalid denom: ",
		},
		{
			name:        "invalid token a",
			allowedPool: types.NewAllowedPool("1ukava", "ukava"),
			expectedErr: "invalid denom: 1ukava",
		},
		{
			name:        "invalid token b",
			allowedPool: types.NewAllowedPool("ukava", "1ukava"),
			expectedErr: "invalid denom: 1ukava",
		},
		{
			name:        "no uppercase letters token a",
			allowedPool: types.NewAllowedPool("uKava", "ukava"),
			expectedErr: "invalid denom: uKava",
		},
		{
			name:        "no uppercase letters token b",
			allowedPool: types.NewAllowedPool("ukava", "UKAVA"),
			expectedErr: "invalid denom: UKAVA",
		},
		{
			name:        "matching tokens",
			allowedPool: types.NewAllowedPool("ukava", "ukava"),
			expectedErr: "pool cannot have two tokens of the same type, received 'ukava' and 'ukava'",
		},
		{
			name:        "invalid token order",
			allowedPool: types.NewAllowedPool("usdx", "ukava"),
			expectedErr: "invalid token order: 'ukava' must come before 'usdx'",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.allowedPool.Validate()
			assert.EqualError(t, err, tc.expectedErr)
		})
	}
}

// ensure no regression in case insentive token matching if
// sdk.ValidateDenom ever allows upper case letters
func TestAllowedPool_TokenMatch(t *testing.T) {
	allowedPool := types.NewAllowedPool("UKAVA", "ukava")
	err := allowedPool.Validate()
	assert.Error(t, err)

	allowedPool = types.NewAllowedPool("hard", "haRd")
	err = allowedPool.Validate()
	assert.Error(t, err)

	allowedPool = types.NewAllowedPool("Usdx", "uSdX")
	err = allowedPool.Validate()
	assert.Error(t, err)
}

func TestAllowedPool_String(t *testing.T) {
	allowedPool := types.NewAllowedPool("hard", "ukava")
	require.NoError(t, allowedPool.Validate())

	output := `AllowedPool:
  Name: hard:ukava
	Token A: hard
	Token B: ukava
`
	assert.Equal(t, output, allowedPool.String())
}

func TestAllowedPool_Name(t *testing.T) {
	testCases := []struct {
		tokens string
		name   string
	}{
		{
			tokens: "atoken btoken",
			name:   "atoken:btoken",
		},
		{
			tokens: "aaa aaaa",
			name:   "aaa:aaaa",
		},
		{
			tokens: "aaaa aaab",
			name:   "aaaa:aaab",
		},
		{
			tokens: "a001 a002",
			name:   "a001:a002",
		},
		{
			tokens: "hard ukava",
			name:   "hard:ukava",
		},
		{
			tokens: "bnb hard",
			name:   "bnb:hard",
		},
		{
			tokens: "bnb xrpb",
			name:   "bnb:xrpb",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.tokens, func(t *testing.T) {
			tokens := strings.Split(tc.tokens, " ")
			require.Equal(t, 2, len(tokens))

			allowedPool := types.NewAllowedPool(tokens[0], tokens[1])
			require.NoError(t, allowedPool.Validate())

			assert.Equal(t, tc.name, allowedPool.Name())
		})
	}
}

func TestAllowedPools_Validate(t *testing.T) {
	testCases := []struct {
		name         string
		allowedPools types.AllowedPools
		expectedErr  string
	}{
		{
			name: "invalid pool",
			allowedPools: types.NewAllowedPools(
				types.NewAllowedPool("hard", "ukava"),
				types.NewAllowedPool("HARD", "UKAVA"),
			),
			expectedErr: "invalid denom: HARD",
		},
		{
			name: "duplicate pool",
			allowedPools: types.NewAllowedPools(
				types.NewAllowedPool("hard", "ukava"),
				types.NewAllowedPool("hard", "ukava"),
			),
			expectedErr: "duplicate pool: hard:ukava",
		},
		{
			name: "duplicate pools",
			allowedPools: types.NewAllowedPools(
				types.NewAllowedPool("hard", "ukava"),
				types.NewAllowedPool("bnb", "usdx"),
				types.NewAllowedPool("btcb", "xrpb"),
				types.NewAllowedPool("bnb", "usdx"),
			),
			expectedErr: "duplicate pool: bnb:usdx",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.allowedPools.Validate()
			assert.EqualError(t, err, tc.expectedErr)
		})
	}
}
