package testutil

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/community"
	"github.com/kava-labs/kava/x/community/keeper"
	"github.com/kava-labs/kava/x/community/types"
	kavadisttypes "github.com/kava-labs/kava/x/kavadist/types"
)

type testFunc func(sdk.Context, keeper.Keeper)

// Test suite used for all abci inflation tests
type disableInflationTestSuite struct {
	suite.Suite

	App    app.TestApp
	Ctx    sdk.Context
	Keeper keeper.Keeper

	genesisMintState     *minttypes.GenesisState
	genesisKavadistState *kavadisttypes.GenesisState

	testFunc testFunc
}

func NewDisableInflationTestSuite(tf testFunc) *disableInflationTestSuite {
	suite := &disableInflationTestSuite{}
	suite.testFunc = tf
	return suite
}

// The default state used by each test
func (suite *disableInflationTestSuite) SetupTest() {
	app.SetSDKConfig()
	tApp := app.NewTestApp()
	suite.App = tApp
	suite.Ctx = suite.App.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})
	suite.Keeper = suite.App.GetCommunityKeeper()

	// Set up x/mint and x/kavadist gen state
	mintGen := minttypes.DefaultGenesisState()
	mintGen.Params.InflationMax = sdk.NewDecWithPrec(595, 3)
	mintGen.Params.InflationMin = sdk.NewDecWithPrec(595, 3)
	suite.genesisMintState = mintGen

	kavadistGen := kavadisttypes.DefaultGenesisState()
	kavadistGen.Params.Active = true
	suite.genesisKavadistState = kavadistGen

	appCodec := tApp.AppCodec()
	suite.App.InitializeFromGenesisStates(
		app.GenesisState{minttypes.ModuleName: appCodec.MustMarshalJSON(mintGen)},
		app.GenesisState{kavadisttypes.ModuleName: appCodec.MustMarshalJSON(kavadistGen)},
	)
}

func (suite *disableInflationTestSuite) TestDisableInflation() {
	validateState := func(upgraded bool, expectedDisableTime time.Time, msg string) {
		params, found := suite.Keeper.GetParams(suite.Ctx)
		suite.Require().True(found)
		mintParams := suite.App.GetMintKeeper().GetParams(suite.Ctx)
		kavadistParams := suite.App.GetKavadistKeeper().GetParams(suite.Ctx)

		disableTimeMsg := "expected inflation disable time to match"
		expectedMintState := suite.genesisMintState
		expectedKavadistState := suite.genesisKavadistState
		msgSuffix := "before upgrade"

		// The state expected after upgrade time is reached
		if upgraded {
			// Disable upgrade time is reset when run.
			//
			// This allows the time to be set and run again if required.
			// In addition, with zero time not upgrading, achieves idempotence
			// without extra logic or state.
			expectedDisableTime = time.Time{}
			disableTimeMsg = "expected inflation disable time to be reset"

			expectedMintState.Params.InflationMin = sdk.ZeroDec()
			expectedMintState.Params.InflationMax = sdk.ZeroDec()

			expectedKavadistState.Params.Active = false
			msgSuffix = "after upgrade"
		}

		suite.Require().Equal(expectedMintState.Params.InflationMin, mintParams.InflationMin, msg+": expected mint inflation min to match state "+msgSuffix)
		suite.Require().Equal(expectedMintState.Params.InflationMax, mintParams.InflationMax, msg+": expected mint inflation max to match state "+msgSuffix)
		suite.Require().Equal(expectedKavadistState.Params.Active, kavadistParams.Active, msg+":expected kavadist active flag match state "+msgSuffix)
		suite.Require().Equal(expectedDisableTime, params.UpgradeTimeDisableInflation, msg+": "+disableTimeMsg)
	}

	blockTime := suite.Ctx.BlockTime()
	testCases := []struct {
		name          string
		upgradeTime   time.Time
		shouldUpgrade bool
	}{
		{"zero upgrade time -- should not upgrade", time.Time{}, false},
		{"upgrade time in future -- should not upgrade", blockTime.Add(1 * time.Second), false},
		{"upgrade time in past -- should upgrade", blockTime.Add(-1 * time.Second), true},
		{"upgrade time equal to block time -- should upgrade", blockTime, true},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			// ensure state is as we expect before running upgrade or updating time
			validateState(false, time.Time{}, "initial state")

			// set inflation disable time
			params, found := suite.Keeper.GetParams(suite.Ctx)
			suite.Require().True(found)
			params.UpgradeTimeDisableInflation = tc.upgradeTime
			suite.Keeper.SetParams(suite.Ctx, params)

			// run test function
			suite.testFunc(suite.Ctx, suite.Keeper)

			// run assertions to ensure upgrade did or did not run
			validateState(tc.shouldUpgrade, tc.upgradeTime, "first begin blocker run")

			// test idempotence only if upgrade should have been ran
			if tc.shouldUpgrade {
				// reset mint and kavadist state to their initial values
				suite.App.GetMintKeeper().SetParams(suite.Ctx, suite.genesisMintState.Params)
				suite.App.GetKavadistKeeper().SetParams(suite.Ctx, suite.genesisKavadistState.Params)

				// run begin blocker again
				community.BeginBlocker(suite.Ctx, suite.Keeper)

				// ensure begin blocker is impodent and never runs twice
				validateState(false, time.Time{}, "second begin blocker run")
			}
		})
	}
}

func (suite *disableInflationTestSuite) TestPanicsOnMissingParameters() {
	suite.SetupTest()

	store := suite.Ctx.KVStore(suite.App.GetKVStoreKey(types.StoreKey))
	store.Delete(types.ParamsKey)

	suite.PanicsWithValue("invalid state: module parameters not found", func() {
		suite.testFunc(suite.Ctx, suite.Keeper)
	})
}
