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

	p := types.NewParams(
		conversionPairs,
	)

	data, err := yaml.Marshal(p)
	suite.Require().NoError(err)

	var params map[string]interface{}
	err = yaml.Unmarshal(data, &params)
	suite.Require().NoError(err)
	_, ok := params["enabled_conversion_pairs"]
	suite.Require().True(ok, "enabled_conversion_pairs should exist in yaml")
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

func TestParamsTestSuite(t *testing.T) {
	suite.Run(t, new(ParamsTestSuite))
}
