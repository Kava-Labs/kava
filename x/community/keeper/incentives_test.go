package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/community/keeper"
	types "github.com/kava-labs/kava/x/community/types"
	kavadisttypes "github.com/kava-labs/kava/x/kavadist/types"
)

// Test suite used for all incentive tests
type IncentivesTestSuite struct {
	suite.Suite

	App    app.TestApp
	Ctx    sdk.Context
	Keeper keeper.Keeper
}

// The default state used by each test
func (suite *IncentivesTestSuite) SetupTest() {
	app.SetSDKConfig()
	tApp := app.NewTestApp()
	suite.App = tApp
	suite.Ctx = suite.App.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})
	suite.Keeper = suite.App.GetCommunityKeeper()

	// Set up x/mint and x/kavadist gen state
	mintGen := minttypes.DefaultGenesisState()
	mintGen.Params.InflationMax = sdk.NewDecWithPrec(595, 3)
	mintGen.Params.InflationMin = sdk.NewDecWithPrec(595, 3)
	kavadistGen := kavadisttypes.DefaultGenesisState()
	kavadistGen.Params.Active = true
	appCodec := tApp.AppCodec()
	suite.App.InitializeFromGenesisStates(
		app.GenesisState{minttypes.ModuleName: appCodec.MustMarshalJSON(mintGen)},
		app.GenesisState{kavadisttypes.ModuleName: appCodec.MustMarshalJSON(kavadistGen)},
	)
}

func TestIncentivesTestSuite(t *testing.T) {
	suite.Run(t, new(IncentivesTestSuite))
}

func (suite *IncentivesTestSuite) TestShouldStartDisableInflationUpgrade() {
	shouldUpgrade := func() bool {
		return suite.Keeper.ShouldStartDisableInflationUpgrade(suite.Ctx)
	}

	setUpgradeTimeFromNow := func(t time.Duration) {
		suite.Keeper.SetParams(
			suite.Ctx,
			types.Params{UpgradeTimeDisableInflation: suite.Ctx.BlockTime().Add(t)},
		)
	}

	suite.Run("skips upgrade if community params does not exist", func() {
		suite.SetupTest()

		// remove param from store
		store := suite.Ctx.KVStore(suite.App.GetKVStoreKey(types.StoreKey))
		store.Delete(types.ParamsKey)

		_, found := suite.Keeper.GetParams(suite.Ctx)
		suite.False(found)
		suite.False(shouldUpgrade())
	})

	suite.Run("skips upgrade if upgrade time is set in the future", func() {
		suite.SetupTest()

		setUpgradeTimeFromNow(1 * time.Hour)
		suite.False(shouldUpgrade())
	})

	suite.Run("upgrades if params are set to the default", func() {
		suite.SetupTest()

		suite.Keeper.SetParams(suite.Ctx, types.DefaultParams())
		param, found := suite.Keeper.GetParams(suite.Ctx)
		suite.True(found)
		suite.Equal(time.Time{}, param.UpgradeTimeDisableInflation)
		suite.True(shouldUpgrade())
	})

	suite.Run("upgrades if blockTime is at or after upgrade time", func() {
		suite.SetupTest()

		setUpgradeTimeFromNow(0)
		suite.True(shouldUpgrade())
		setUpgradeTimeFromNow(-2 * time.Minute)
		suite.True(shouldUpgrade())
	})

	suite.Run("skips upgrade if already upgraded", func() {
		suite.SetupTest()

		setUpgradeTimeFromNow(-2 * time.Minute)
		suite.True(shouldUpgrade())
		suite.Keeper.StartDisableInflationUpgrade(suite.Ctx)
		suite.False(shouldUpgrade())
	})
}

func (suite *IncentivesTestSuite) TestStartDisableInflationUpgrade() {
	isUpgraded := func() bool {
		_, found := suite.Keeper.GetPreviousBlockTime(suite.Ctx)
		return found
	}
	setUpgradeTimeFromNow := func(t time.Duration) {
		suite.Keeper.SetParams(
			suite.Ctx,
			types.Params{UpgradeTimeDisableInflation: suite.Ctx.BlockTime().Add(t)},
		)
	}

	suite.Run("upgrade should set mint and kavadist inflation to 0", func() {
		suite.SetupTest()

		mintParams := suite.App.GetMintKeeper().GetParams(suite.Ctx)
		suite.Equal(sdk.NewDecWithPrec(595, 3), mintParams.InflationMax)
		kavadistParams := suite.App.GetKavadistKeeper().GetParams(suite.Ctx)
		suite.True(kavadistParams.Active)

		setUpgradeTimeFromNow(-2 * time.Minute)
		suite.Keeper.StartDisableInflationUpgrade(suite.Ctx)
		suite.True(isUpgraded())

		mintParams = suite.App.GetMintKeeper().GetParams(suite.Ctx)
		suite.Equal(sdk.ZeroDec(), mintParams.InflationMax)
		suite.Equal(sdk.ZeroDec(), mintParams.InflationMin)

		kavadistParams = suite.App.GetKavadistKeeper().GetParams(suite.Ctx)
		suite.False(kavadistParams.Active)
	})

	suite.Run("upgrade should set previous block time", func() {
		suite.SetupTest()

		setUpgradeTimeFromNow(-2 * time.Minute)
		suite.Keeper.StartDisableInflationUpgrade(suite.Ctx)
		suite.True(isUpgraded())
		prevBlockTime, found := suite.Keeper.GetPreviousBlockTime(suite.Ctx)
		suite.True(found)
		suite.Equal(suite.Ctx.BlockTime(), prevBlockTime)
	})
}
