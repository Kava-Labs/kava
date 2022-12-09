package types_test

import (
	"bytes"
	"encoding/json"
	fmt "fmt"
	"reflect"
	"testing"

	"github.com/kava-labs/kava/x/kavamint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"
)

const secondsPerYear = 31536000

func newValidParams(t *testing.T) types.Params {
	// 50%
	poolRate := sdk.MustNewDecFromStr("0.5")

	// 10%
	stakingRate := sdk.MustNewDecFromStr("0.1")

	params := types.NewParams(poolRate, stakingRate)
	require.NoError(t, params.Validate())

	return params
}

func TestParams_Default(t *testing.T) {
	defaultParams := types.DefaultParams()

	require.NoError(t, defaultParams.Validate(), "default parameters must be valid")

	assert.Equal(t, types.DefaultCommunityPoolInflation, defaultParams.CommunityPoolInflation, "expected default pool inflation to match exported default value")
	assert.Equal(t, defaultParams.CommunityPoolInflation, sdk.ZeroDec(), "expected default pool inflation to be zero")

	assert.Equal(t, types.DefaultStakingRewardsApy, defaultParams.StakingRewardsApy, "expected default staking inflation to match exported default value")
	assert.Equal(t, defaultParams.CommunityPoolInflation, sdk.ZeroDec(), "expected default staking inflation to be zero")
}

func TestParams_MaxInflationRate_ApproxRootDoesNotPanic(t *testing.T) {
	require.Equal(t, sdk.NewDec(100), types.MaxMintingRate) // 10,000%, should never be a reason to exceed this value
	maxYearlyRate := types.MaxMintingRate

	require.NotPanics(t, func() {
		expectedMaxRate, err := maxYearlyRate.ApproxRoot(secondsPerYear)
		require.NoError(t, err)
		expectedMaxRate = expectedMaxRate.Sub(sdk.OneDec())
	})
}

func TestParams_MaxInflationRate_DoesNotOverflow(t *testing.T) {
	maxRate := types.MaxMintingRate // use the max minting rate
	totalSupply := sdk.NewDec(1e14) // 100 trillion starting supply
	years := uint64(25)             // calculate over 50 years

	perSecondMaxRate, err := maxRate.ApproxRoot(secondsPerYear)
	require.NoError(t, err)
	perSecondMaxRate = perSecondMaxRate.Sub(sdk.OneDec())

	var finalSupply sdk.Int

	require.NotPanics(t, func() {
		compoundedRate := perSecondMaxRate.Power(years * secondsPerYear)
		finalSupply = totalSupply.Mul(compoundedRate).RoundInt()

	})

	require.Less(t, finalSupply.BigInt().BitLen(), 256)
}

func TestParams_ParamSetPairs_CommunityPoolInflation(t *testing.T) {
	assert.Equal(t, []byte("CommunityPoolInflation"), types.KeyCommunityPoolInflation)
	defaultParams := types.DefaultParams()

	var paramSetPair *paramstypes.ParamSetPair
	for _, pair := range defaultParams.ParamSetPairs() {
		if bytes.Equal(pair.Key, types.KeyCommunityPoolInflation) {
			paramSetPair = &pair
			break
		}
	}
	require.NotNil(t, paramSetPair)

	pairs, ok := paramSetPair.Value.(*sdk.Dec)
	require.True(t, ok)
	assert.Equal(t, pairs, &defaultParams.CommunityPoolInflation)

	assert.Nil(t, paramSetPair.ValidatorFn(*pairs))
	assert.EqualError(t, paramSetPair.ValidatorFn(struct{}{}), "invalid parameter type: struct {}")
}

func TestParams_ParamSetPairs_StakingRewardsApy(t *testing.T) {
	assert.Equal(t, []byte("StakingRewardsApy"), types.KeyStakingRewardsApy)
	defaultParams := types.DefaultParams()

	var paramSetPair *paramstypes.ParamSetPair
	for _, pair := range defaultParams.ParamSetPairs() {
		if bytes.Equal(pair.Key, types.KeyStakingRewardsApy) {
			paramSetPair = &pair
			break
		}
	}
	require.NotNil(t, paramSetPair)

	pairs, ok := paramSetPair.Value.(*sdk.Dec)
	require.True(t, ok)
	assert.Equal(t, pairs, &defaultParams.StakingRewardsApy)

	assert.Nil(t, paramSetPair.ValidatorFn(*pairs))
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
			name: "nil community pool inflation",
			key:  types.KeyCommunityPoolInflation,
			testFn: func(params *types.Params) {
				params.CommunityPoolInflation = sdk.Dec{}
			},
			expectedErr: "invalid rate: <nil>",
		},
		{
			name: "negative community pool inflation",
			key:  types.KeyCommunityPoolInflation,
			testFn: func(params *types.Params) {
				params.CommunityPoolInflation = sdk.MustNewDecFromStr("-0.000000000011111111")
			},
			expectedErr: "invalid rate: -0.000000000011111111",
		},
		{
			name: "0 community pool inflation",
			key:  types.KeyCommunityPoolInflation,
			testFn: func(params *types.Params) {
				params.CommunityPoolInflation = sdk.ZeroDec()
			},
			expectedErr: "", // ok
		},
		{
			name: "community pool inflation 1e-18 less than max rate",
			key:  types.KeyCommunityPoolInflation,
			testFn: func(params *types.Params) {
				params.CommunityPoolInflation = types.MaxMintingRate.Sub(sdk.NewDecWithPrec(1, 18))
			},
			expectedErr: "", // ok
		},
		{
			name: "community pool inflation equal to max rate",
			key:  types.KeyCommunityPoolInflation,
			testFn: func(params *types.Params) {
				params.CommunityPoolInflation = types.MaxMintingRate
			},
			expectedErr: "", // ok
		},
		{
			name: "community pool inflation 1e-18 greater than max rate",
			key:  types.KeyCommunityPoolInflation,
			testFn: func(params *types.Params) {
				params.CommunityPoolInflation = types.MaxMintingRate.Add(sdk.NewDecWithPrec(1, 18))
			},
			expectedErr: "invalid rate: 100.000000000000000001",
		},
		{
			name: "nil staking inflation",
			key:  types.KeyStakingRewardsApy,
			testFn: func(params *types.Params) {
				params.StakingRewardsApy = sdk.Dec{}
			},
			expectedErr: "invalid rate: <nil>",
		},
		{
			name: "negative staking inflation",
			key:  types.KeyStakingRewardsApy,
			testFn: func(params *types.Params) {
				params.StakingRewardsApy = sdk.MustNewDecFromStr("-0.000000002222222222")
			},
			expectedErr: "invalid rate: -0.000000002222222222",
		},
		{
			name: "0 staking inflation",
			key:  types.KeyStakingRewardsApy,
			testFn: func(params *types.Params) {
				params.StakingRewardsApy = sdk.ZeroDec()
			},
			expectedErr: "", // ok
		},
		{
			name: "staking inflation 1e-18 less than max rate",
			key:  types.KeyStakingRewardsApy,
			testFn: func(params *types.Params) {
				params.StakingRewardsApy = types.MaxMintingRate.Sub(sdk.NewDecWithPrec(1, 18))
			},
			expectedErr: "", // ok
		},
		{
			name: "staking inflation equal to max rate",
			key:  types.KeyStakingRewardsApy,
			testFn: func(params *types.Params) {
				params.StakingRewardsApy = types.MaxMintingRate
			},
			expectedErr: "", // ok
		},
		{
			name: "staking inflation 1e-18 greater than max rate",
			key:  types.KeyStakingRewardsApy,
			testFn: func(params *types.Params) {
				params.StakingRewardsApy = types.MaxMintingRate.Add(sdk.NewDecWithPrec(1, 18))
			},
			expectedErr: "invalid rate: 100.000000000000000001",
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
	params := newValidParams(t)

	output := params.String()
	assert.Contains(t, output, fmt.Sprintf("CommunityPoolInflation: %s", params.CommunityPoolInflation.String()))
	assert.Contains(t, output, fmt.Sprintf("StakingRewardsApy: %s", params.StakingRewardsApy.String()))
}

func TestParams_UnmarshalJSON(t *testing.T) {
	params := newValidParams(t)

	poolRateData, err := json.Marshal(params.CommunityPoolInflation)
	require.NoError(t, err)

	stakingRateData, err := json.Marshal(params.StakingRewardsApy)
	require.NoError(t, err)

	data := fmt.Sprintf(`{
	"community_pool_inflation": %s,
	"staking_rewards_apy": %s
}`, string(poolRateData), string(stakingRateData))

	var parsedParams types.Params
	err = json.Unmarshal([]byte(data), &parsedParams)
	require.NoError(t, err)

	assert.Equal(t, params.CommunityPoolInflation, parsedParams.CommunityPoolInflation)
	assert.Equal(t, params.StakingRewardsApy, parsedParams.StakingRewardsApy)
}

func TestParams_MarshalYAML(t *testing.T) {
	p := newValidParams(t)

	data, err := yaml.Marshal(p)
	require.NoError(t, err)

	var params map[string]interface{}
	err = yaml.Unmarshal(data, &params)
	require.NoError(t, err)

	_, ok := params["community_pool_inflation"]
	require.True(t, ok)
	_, ok = params["staking_rewards_apy"]
	require.True(t, ok)
}
