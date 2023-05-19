package types_test

import (
	bytes "bytes"
	"testing"

	"github.com/stretchr/testify/suite"
	"sigs.k8s.io/yaml"

	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/evmutil/testutil"
	"github.com/kava-labs/kava/x/evmutil/types"
)

type ParamsTestSuite struct {
	suite.Suite
}

func (suite *ParamsTestSuite) SetupTest() {
	app.SetSDKConfig()
}

func (suite *ParamsTestSuite) TestDefault() {
	defaultParams := types.DefaultParams()
	suite.Require().NoError(defaultParams.Validate())
}

func (suite *ParamsTestSuite) TestMarshalYAML() {
	conversionPairs := types.NewConversionPairs(
		types.NewConversionPair(
			testutil.MustNewInternalEVMAddressFromString("0x0000000000000000000000000000000000000001"),
			"usdc",
		),
	)
	allowedNativeDenoms := types.NewAllowedNativeCoinERC20Tokens(
		types.NewAllowedNativeCoinERC20Token("denom", "Sdk Denom!", "DENOM", 6),
	)

	p := types.NewParams(
		conversionPairs,
		allowedNativeDenoms,
	)

	data, err := yaml.Marshal(p)
	suite.Require().NoError(err)

	var params map[string]interface{}
	err = yaml.Unmarshal(data, &params)
	suite.Require().NoError(err)
	_, ok := params["enabled_conversion_pairs"]
	suite.Require().True(ok, "enabled_conversion_pairs should exist in yaml")
	_, ok = params["allowed_native_denoms"]
	suite.Require().True(ok, "allowed_native_denoms should exist in yaml")
}

func (suite *ParamsTestSuite) TestParamSetPairs_EnabledConversionPairs() {
	suite.Require().Equal([]byte("EnabledConversionPairs"), types.KeyEnabledConversionPairs)
	defaultParams := types.DefaultParams()

	var paramSetPair *paramstypes.ParamSetPair
	for _, pair := range defaultParams.ParamSetPairs() {
		if bytes.Equal(pair.Key, types.KeyEnabledConversionPairs) {
			paramSetPair = &pair
			break
		}
	}
	suite.Require().NotNil(paramSetPair)

	pairs, ok := paramSetPair.Value.(*types.ConversionPairs)
	suite.Require().True(ok)
	suite.Require().Equal(pairs, &defaultParams.EnabledConversionPairs)

	suite.Require().Nil(paramSetPair.ValidatorFn(*pairs))
	suite.Require().EqualError(paramSetPair.ValidatorFn(struct{}{}), "invalid parameter type: struct {}")
}

func (suite *ParamsTestSuite) TestParamSetPairs_AllowedNativeDenoms() {
	suite.Require().Equal([]byte("AllowedNativeDenoms"), types.KeyAllowedNativeDenoms)
	defaultParams := types.DefaultParams()

	var paramSetPair *paramstypes.ParamSetPair
	for _, pair := range defaultParams.ParamSetPairs() {
		if bytes.Equal(pair.Key, types.KeyAllowedNativeDenoms) {
			paramSetPair = &pair
			break
		}
	}
	suite.Require().NotNil(paramSetPair)

	pairs, ok := paramSetPair.Value.(*types.AllowedNativeCoinERC20Tokens)
	suite.Require().True(ok)
	suite.Require().Equal(pairs, &defaultParams.AllowedNativeDenoms)

	suite.Require().Nil(paramSetPair.ValidatorFn(*pairs))
	suite.Require().EqualError(paramSetPair.ValidatorFn(struct{}{}), "invalid parameter type: struct {}")
}

func (suite *ParamsTestSuite) TestParams_Validate() {
	validConversionPairs := types.NewConversionPairs(
		types.NewConversionPair(
			testutil.MustNewInternalEVMAddressFromString("0x0000000000000000000000000000000000000001"),
			"usdc",
		),
	)
	invalidConversionPairs := types.NewConversionPairs(
		types.NewConversionPair(
			testutil.MustNewInternalEVMAddressFromString("0x000000000000000000000000000000000000000A"),
			"kava",
		),
		types.NewConversionPair(
			testutil.MustNewInternalEVMAddressFromString("0x000000000000000000000000000000000000000B"),
			"kava", // duplicate denom!
		),
	)
	validAllowedNativeDenoms := types.NewAllowedNativeCoinERC20Tokens(
		types.NewAllowedNativeCoinERC20Token("hard", "EVM Hard", "HARD", 6),
	)
	invalidAllowedNativeDenoms := types.NewAllowedNativeCoinERC20Tokens(
		types.NewAllowedNativeCoinERC20Token("", "Invalid Token", "NOPE", 0), // empty sdk denom
	)

	testCases := []struct {
		name   string
		params types.Params
		expErr string
	}{
		{
			name:   "valid - empty",
			params: types.NewParams(types.NewConversionPairs(), types.NewAllowedNativeCoinERC20Tokens()),
			expErr: "",
		},
		{
			name:   "valid - with data",
			params: types.NewParams(validConversionPairs, validAllowedNativeDenoms),
			expErr: "",
		},
		{
			name:   "invalid - invalid conversion pair",
			params: types.NewParams(invalidConversionPairs, validAllowedNativeDenoms),
			expErr: "found duplicate",
		},
		{
			name:   "invalid - invalid allowed native denoms",
			params: types.NewParams(validConversionPairs, invalidAllowedNativeDenoms),
			expErr: "invalid token",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			err := tc.params.Validate()
			if tc.expErr != "" {
				suite.ErrorContains(err, tc.expErr, "Expected validation error")
			} else {
				suite.NoError(err, "Expected no validation error")
			}
		})
	}
}

func TestParamsTestSuite(t *testing.T) {
	suite.Run(t, new(ParamsTestSuite))
}
