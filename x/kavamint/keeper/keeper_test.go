package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/kavamint/testutil"
	"github.com/kava-labs/kava/x/kavamint/types"
	"github.com/stretchr/testify/suite"
)

type keeperTestSuite struct {
	testutil.KavamintTestSuite
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(keeperTestSuite))
}

func (suite keeperTestSuite) TestParams_Persistance() {
	keeper := suite.Keeper

	params := types.NewParams(
		sdk.MustNewDecFromStr("0.000000000000000001"),
		sdk.MustNewDecFromStr("0.000000000000000002"),
	)
	keeper.SetParams(suite.Ctx, params)
	suite.Equal(keeper.GetParams(suite.Ctx), params)

	oldParams := params
	params = types.NewParams(
		sdk.MustNewDecFromStr("0.000000000000000011"),
		sdk.MustNewDecFromStr("0.000000000000000022"),
	)
	keeper.SetParams(suite.Ctx, params)
	suite.NotEqual(keeper.GetParams(suite.Ctx), oldParams)
	suite.Equal(keeper.GetParams(suite.Ctx), params)
}

func (suite keeperTestSuite) TestPreviousBlockTime_Persistance() {
	keeper := suite.Keeper
	zeroTime := time.Time{}

	keeper.SetPreviousBlockTime(suite.Ctx, zeroTime)
	suite.Equal(keeper.GetPreviousBlockTime(suite.Ctx), zeroTime)

	newTime := suite.Ctx.BlockTime()
	suite.Require().False(newTime.IsZero())

	keeper.SetPreviousBlockTime(suite.Ctx, newTime)
	suite.Equal(keeper.GetPreviousBlockTime(suite.Ctx), newTime)
}
