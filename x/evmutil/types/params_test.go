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
	allowedCosmosDenoms := types.NewAllowedCosmosCoinERC20Tokens(
		types.NewAllowedCosmosCoinERC20Token("denom", "Sdk Denom!", "DENOM", 6),
	)

	p := types.NewParams(
		conversionPairs,
		allowedCosmosDenoms,
	)

	data, err := yaml.Marshal(p)
	suite.Require().NoError(err)

	var params map[string]interface{}
	err = yaml.Unmarshal(data, &params)
	suite.Require().NoError(err)
	_, ok := params["enabled_conversion_pairs"]
	suite.Require().True(ok, "enabled_conversion_pairs should exist in yaml")
	_, ok = params["allowed_cosmos_denoms"]
	suite.Require().True(ok, "allowed_cosmos_denoms should exist in yaml")
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

func (suite *ParamsTestSuite) TestParamSetPairs_AllowedCosmosDenoms() {
	suite.Require().Equal([]byte("AllowedCosmosDenoms"), types.KeyAllowedCosmosDenoms)
	defaultParams := types.DefaultParams()

	var paramSetPair *paramstypes.ParamSetPair
	for _, pair := range defaultParams.ParamSetPairs() {
		if bytes.Equal(pair.Key, types.KeyAllowedCosmosDenoms) {
			paramSetPair = &pair
			break
		}
	}
	suite.Require().NotNil(paramSetPair)

	pairs, ok := paramSetPair.Value.(*types.AllowedCosmosCoinERC20Tokens)
	suite.Require().True(ok)
	suite.Require().Equal(pairs, &defaultParams.AllowedCosmosDenoms)

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
	validAllowedCosmosDenoms := types.NewAllowedCosmosCoinERC20Tokens(
		types.NewAllowedCosmosCoinERC20Token("hard", "EVM Hard", "HARD", 6),
	)
	invalidAllowedCosmosDenoms := types.NewAllowedCosmosCoinERC20Tokens(
		types.NewAllowedCosmosCoinERC20Token("", "Invalid Token", "NOPE", 0), // empty sdk denom
	)

	testCases := []struct {
		name   string
		params types.Params
		expErr string
	}{
		{
			name:   "valid - empty",
			params: types.NewParams(types.NewConversionPairs(), types.NewAllowedCosmosCoinERC20Tokens()),
			expErr: "",
		},
		{
			name:   "valid - with data",
			params: types.NewParams(validConversionPairs, validAllowedCosmosDenoms),
			expErr: "",
		},
		{
			name:   "invalid - invalid conversion pair",
			params: types.NewParams(invalidConversionPairs, validAllowedCosmosDenoms),
			expErr: "found duplicate",
		},
		{
			name:   "invalid - invalid allowed cosmos denoms",
			params: types.NewParams(validConversionPairs, invalidAllowedCosmosDenoms),
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
