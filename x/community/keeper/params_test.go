package keeper_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/community/keeper"
	"github.com/kava-labs/kava/x/community/types"
)

// Test suite used for all store tests
type StoreTestSuite struct {
	suite.Suite

	App    app.TestApp
	Ctx    sdk.Context
	Keeper keeper.Keeper
}

// The default state used by each test
func (suite *StoreTestSuite) SetupTest() {
	app.SetSDKConfig()
	suite.App = app.NewTestApp()
	suite.Ctx = suite.App.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})
	suite.Keeper = suite.App.GetCommunityKeeper()
}

func TestStoreTestSuite(t *testing.T) {
	suite.Run(t, new(StoreTestSuite))
}

func (suite *StoreTestSuite) TestGetSetParams() {
	suite.Run("get params returns not found on empty store", func() {
		_, found := suite.Keeper.GetParams(suite.Ctx)
		suite.Require().False(found)
	})

	suite.Run("set params cannot store invalid params", func() {
		invalid := types.Params{UpgradeTimeDisableInflation: time.Date(-1, 1, 1, 0, 0, 0, 0, time.UTC)}
		suite.Panics(func() {
			suite.Keeper.SetParams(suite.Ctx, invalid)
		})
	})

	suite.Run("get params returns stored params", func() {
		suite.Keeper.SetParams(suite.Ctx, types.DefaultParams())

		storedParams, found := suite.Keeper.GetParams(suite.Ctx)
		suite.True(found)
		suite.Equal(types.DefaultParams(), storedParams)
	})

	suite.Run("set overwrite previous value", func() {
		suite.Keeper.SetParams(suite.Ctx, types.DefaultParams())

		params := types.NewParams(
			time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC),
			sdkmath.LegacyNewDec(1000),
			sdkmath.LegacyNewDec(1000),
		)
		suite.Keeper.SetParams(suite.Ctx, params)

		storedParams, found := suite.Keeper.GetParams(suite.Ctx)
		suite.True(found)
		suite.NotEqual(params, types.DefaultParams())
		suite.Equal(params, storedParams)
	})
}
