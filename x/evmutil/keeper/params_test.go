package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/x/evmutil/testutil"
	"github.com/kava-labs/kava/x/evmutil/types"
)

type ParamsTestSuite struct {
	testutil.Suite
}

func TestParamsSuite(t *testing.T) {
	suite.Run(t, new(ParamsTestSuite))
}

func (suite *ParamsTestSuite) TestEnabledConversionPair() {
	pairAddr := testutil.MustNewInternalEVMAddressFromString("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2")
	expPair := types.ConversionPair{
		KavaERC20Address: pairAddr.Bytes(),
		Denom:            "weth",
	}
	params := types.DefaultParams()
	params.EnabledConversionPairs = []types.ConversionPair{expPair}
	suite.Keeper.SetParams(suite.Ctx, params)

	actualPair, err := suite.Keeper.GetEnabledConversionPairFromERC20Address(
		suite.Ctx,
		pairAddr,
	)
	suite.Require().NoError(err)
	suite.Require().Equal(expPair, actualPair)
}
