package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/evmutil/keeper"
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

func (suite *ParamsTestSuite) TestHistoricParamsQuery() {
	// setup a params store that lacks allowed_native_denoms param (as was the case in v1)
	oldParamStore := suite.App.GetParamsKeeper().Subspace("test_subspace_for_evmutil")
	oldParamStore.WithKeyTable(types.ParamKeyTable())
	oldParamStore.Set(suite.Ctx, types.KeyEnabledConversionPairs, types.ConversionPairs{})

	suite.True(oldParamStore.Has(suite.Ctx, types.KeyEnabledConversionPairs))
	suite.False(oldParamStore.Has(suite.Ctx, types.KeyAllowedNativeDenoms))

	oldStateKeeper := keeper.NewKeeper(
		suite.App.AppCodec(),
		sdk.NewKVStoreKey(types.StoreKey),
		oldParamStore,
		suite.App.GetBankKeeper(),
		suite.App.GetAccountKeeper(),
	)

	// prior to making GetParams() use GetParamSetIfExists, this would panic.
	suite.NotPanics(func() {
		_ = oldStateKeeper.GetParams(suite.Ctx)
	})
}
