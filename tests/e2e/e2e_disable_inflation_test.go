package e2e_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	"github.com/kava-labs/kava/tests/util"
	communitytypes "github.com/kava-labs/kava/x/community/types"
	kavadisttypes "github.com/kava-labs/kava/x/kavadist/types"
)

func (suite *IntegrationTestSuite) TestDisableInflationOnUpgrade() {
	beforeInflationDisableCtx := util.CtxAtHeight(1)
	afterInflationDisableCtx := util.CtxAtHeight(2)

	// Before balances
	kavaDistBalBefore, err := suite.Kava.Kavadist.Balance(beforeInflationDisableCtx, &kavadisttypes.QueryBalanceRequest{})
	suite.NoError(err)
	distrBalBefore, err := suite.Kava.Distribution.CommunityPool(beforeInflationDisableCtx, &distrtypes.QueryCommunityPoolRequest{})
	suite.NoError(err)
	distrBalCoinsBefore, distrBalDustBefore := distrBalBefore.Pool.TruncateDecimal()
	beforeCommPoolBalance, err := suite.Kava.Community.Balance(beforeInflationDisableCtx, &communitytypes.QueryBalanceRequest{})
	suite.NoError(err)

	// Before parameters
	suite.Run("x/distribution and x/kavadist parameters before upgrade", func() {
		kavaDistParamsBefore, err := suite.Kava.Kavadist.Params(beforeInflationDisableCtx, &kavadisttypes.QueryParamsRequest{})
		suite.NoError(err)
		mintParamsBefore, err := suite.Kava.Mint.Params(beforeInflationDisableCtx, &minttypes.QueryParamsRequest{})
		suite.NoError(err)

		suite.Require().True(
			kavaDistParamsBefore.Params.Active,
			"x/kavadist should be active before upgrade",
		)
		suite.Require().True(
			mintParamsBefore.Params.InflationMax.IsPositive(),
			"x/mint inflation max should be positive before upgrade",
		)
		suite.Require().True(
			mintParamsBefore.Params.InflationMin.IsPositive(),
			"x/mint inflation min should be positive before upgrade",
		)
	})

	// After parameters
	suite.Run("x/distribution and x/kavadist parameters after upgrade", func() {
		kavaDistParamsAfter, err := suite.Kava.Kavadist.Params(afterInflationDisableCtx, &kavadisttypes.QueryParamsRequest{})
		suite.NoError(err)
		mintParamsAfter, err := suite.Kava.Mint.Params(afterInflationDisableCtx, &minttypes.QueryParamsRequest{})
		suite.NoError(err)

		suite.Require().False(
			kavaDistParamsAfter.Params.Active,
			"x/kavadist should be inactive after upgrade",
		)
		suite.Require().True(
			mintParamsAfter.Params.InflationMax.IsZero(),
			"x/mint inflation max should be zero after upgrade",
		)
		suite.Require().True(
			mintParamsAfter.Params.InflationMin.IsZero(),
			"x/mint inflation min should be zero after upgrade",
		)
	})

	suite.Run("x/distribution and x/kavadist balances after upgrade", func() {
		// After balances
		kavaDistBalAfter, err := suite.Kava.Kavadist.Balance(afterInflationDisableCtx, &kavadisttypes.QueryBalanceRequest{})
		suite.NoError(err)
		distrBalAfter, err := suite.Kava.Distribution.CommunityPool(afterInflationDisableCtx, &distrtypes.QueryCommunityPoolRequest{})
		suite.NoError(err)
		afterCommPoolBalance, err := suite.Kava.Community.Balance(afterInflationDisableCtx, &communitytypes.QueryBalanceRequest{})
		suite.NoError(err)

		// expect empty balances after (ignoring dust in x/distribution)
		suite.Equal(sdk.NewCoins(), kavaDistBalAfter.Coins)
		distrCoinsAfter, distrBalDustAfter := distrBalAfter.Pool.TruncateDecimal()
		suite.Equal(sdk.NewCoins(), distrCoinsAfter)

		// x/kavadist and x/distribution community pools should be moved to x/community
		suite.Equal(
			beforeCommPoolBalance.Coins.
				Add(kavaDistBalBefore.Coins...).
				Add(distrBalCoinsBefore...),
			afterCommPoolBalance.Coins,
		)

		// x/distribution dust should stay in x/distribution
		suite.Equal(distrBalDustBefore, distrBalDustAfter)
	})
}

func (suite *IntegrationTestSuite) TestDisableInflationOnNewChain() {

}
