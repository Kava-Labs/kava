package community_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/community"
	"github.com/kava-labs/kava/x/community/keeper"
	kavadisttypes "github.com/kava-labs/kava/x/kavadist/types"
)

// Test suite used for all abci tests
type ABCITestSuite struct {
	suite.Suite

	App    app.TestApp
	Ctx    sdk.Context
	Keeper keeper.Keeper
}

// The default state used by each test
func (suite *ABCITestSuite) SetupTest() {
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

func TestABCITestSuite(t *testing.T) {
	suite.Run(t, new(ABCITestSuite))
}

func (suite *ABCITestSuite) TestBeginBlockerPayoutStakingRewards() {
	validateRewardsPaid := func(rewards int64) {
		bk := suite.App.GetBankKeeper()
		feeCollectorAcc := suite.App.GetAccountKeeper().GetModuleAccount(suite.Ctx, authtypes.FeeCollectorName)
		feeCollectorBal := bk.GetBalance(suite.Ctx, feeCollectorAcc.GetAddress(), "ukava")
		suite.Equal(rewards, feeCollectorBal.Amount.Int64())
	}

	suite.Run("does not pay rewards if not upgraded", func() {
		suite.SetupTest()
		params, _ := suite.Keeper.GetParams(suite.Ctx)
		params.UpgradeTimeDisableInflation = suite.Ctx.BlockTime().Add(time.Hour * 1)
		community.BeginBlocker(suite.Ctx, suite.Keeper)
		validateRewardsPaid(0)
	})

	suite.Run("pays rewards if upgraded", func() {
		suite.SetupTest()

		// pays out 0 rewards on upgrade block
		community.BeginBlocker(suite.Ctx, suite.Keeper)
		validateRewardsPaid(0)

		// pays out correct rewards on next block in 6.2 seconds
		ctx := suite.Ctx.WithBlockTime(suite.Ctx.BlockTime().Add(6_200 * time.Millisecond))
		community.BeginBlocker(ctx, suite.Keeper)
		validateRewardsPaid(4_465_146)

		// pays out correct rewards on next block in 8.74 seconds
		ctx = suite.Ctx.WithBlockTime(suite.Ctx.BlockTime().Add(8_740 * time.Millisecond))
		community.BeginBlocker(ctx, suite.Keeper)
		validateRewardsPaid(5_953_528)

		// pays out correct rewards on next block in 12.5 seconds
		ctx = suite.Ctx.WithBlockTime(suite.Ctx.BlockTime().Add(12_500 * time.Millisecond))
		community.BeginBlocker(ctx, suite.Keeper)
		validateRewardsPaid(8_930_292)
	})
}

func (suite *ABCITestSuite) TestBeginBlockerDisableInflationUpgrade() {
	validateUpgradedState := func() {
		mintParams := suite.App.GetMintKeeper().GetParams(suite.Ctx)
		suite.Equal(sdk.ZeroDec(), mintParams.InflationMax)
		suite.Equal(sdk.ZeroDec(), mintParams.InflationMin)

		kavadistParams := suite.App.GetKavadistKeeper().GetParams(suite.Ctx)
		suite.False(kavadistParams.Active)
	}

	suite.Run("starts disable inflation upgrade on vanilla chain", func() {
		suite.SetupTest()
		community.BeginBlocker(suite.Ctx, suite.Keeper)
		validateUpgradedState()
	})

	suite.Run("starts disable inflation upgrade when past upgrade time", func() {
		suite.SetupTest()

		suite.setUpgradeTimeFromNow(time.Hour * -1)
		community.BeginBlocker(suite.Ctx, suite.Keeper)
		validateUpgradedState()
	})

	suite.Run("don't upgrade if upgrade time is in the future", func() {
		suite.SetupTest()

		suite.setUpgradeTimeFromNow(time.Hour * 1)
		community.BeginBlocker(suite.Ctx, suite.Keeper)

		mintParams := suite.App.GetMintKeeper().GetParams(suite.Ctx)
		suite.NotEqual(sdk.ZeroDec(), mintParams.InflationMax)
		suite.NotEqual(sdk.ZeroDec(), mintParams.InflationMin)

		kavadistParams := suite.App.GetKavadistKeeper().GetParams(suite.Ctx)
		suite.True(kavadistParams.Active)
	})

	suite.Run("don't run upgrade if already upgraded", func() {
		suite.SetupTest()

		community.BeginBlocker(suite.Ctx, suite.Keeper)
		validateUpgradedState()

		kavadistParams := suite.App.GetKavadistKeeper().GetParams(suite.Ctx)
		kavadistParams.Active = true
		suite.App.GetKavadistKeeper().SetParams(suite.Ctx, kavadistParams)

		suite.Ctx = suite.Ctx.WithBlockTime(suite.Ctx.BlockTime().Add(time.Minute * 6))
		community.BeginBlocker(suite.Ctx, suite.Keeper)

		mintParams := suite.App.GetMintKeeper().GetParams(suite.Ctx)
		suite.Equal(sdk.ZeroDec(), mintParams.InflationMax)
		suite.Equal(sdk.ZeroDec(), mintParams.InflationMin)

		kavadistParams = suite.App.GetKavadistKeeper().GetParams(suite.Ctx)
		suite.True(kavadistParams.Active)
	})
}

func (suite *ABCITestSuite) setUpgradeTimeFromNow(t time.Duration) {
	params, found := suite.Keeper.GetParams(suite.Ctx)
	suite.True(found)
	params.UpgradeTimeDisableInflation = suite.Ctx.BlockTime().Add(t)
	suite.Keeper.SetParams(suite.Ctx, params)
}
