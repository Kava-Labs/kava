package testutil

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	sdkmath "cosmossdk.io/math"
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
	genesisDistrState    *distrtypes.GenesisState

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

	distrGen := distrtypes.DefaultGenesisState()
	distrGen.Params.CommunityTax = sdk.MustNewDecFromStr("0.949500000000000000")
	suite.genesisDistrState = distrGen

	appCodec := tApp.AppCodec()
	suite.App.InitializeFromGenesisStates(
		app.GenesisState{minttypes.ModuleName: appCodec.MustMarshalJSON(mintGen)},
		app.GenesisState{kavadisttypes.ModuleName: appCodec.MustMarshalJSON(kavadistGen)},
		app.GenesisState{distrtypes.ModuleName: appCodec.MustMarshalJSON(distrGen)},
	)
}

func (suite *disableInflationTestSuite) TestDisableInflation() {
	validateState := func(upgraded bool, expectedDisableTime time.Time, originalStakingRewards sdkmath.LegacyDec, setStakingRewards sdkmath.LegacyDec, msg string) {
		params, found := suite.Keeper.GetParams(suite.Ctx)
		suite.Require().True(found)
		mintParams := suite.App.GetMintKeeper().GetParams(suite.Ctx)
		kavadistParams := suite.App.GetKavadistKeeper().GetParams(suite.Ctx)
		distrParams := suite.App.GetDistrKeeper().GetParams(suite.Ctx)

		disableTimeMsg := "expected inflation disable time to match"
		expectedMintState := suite.genesisMintState
		expectedKavadistState := suite.genesisKavadistState
		expectedDistrState := suite.genesisDistrState
		expectedStakingRewards := originalStakingRewards
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
			expectedStakingRewards = setStakingRewards

			expectedMintState.Params.InflationMin = sdk.ZeroDec()
			expectedMintState.Params.InflationMax = sdk.ZeroDec()

			expectedKavadistState.Params.Active = false

			expectedDistrState.Params.CommunityTax = sdk.ZeroDec()

			msgSuffix = "after upgrade"

			suite.Require().NoError(
				app.EventsContains(
					suite.Ctx.EventManager().Events(),
					sdk.NewEvent(
						types.EventTypeInflationStop,
						sdk.NewAttribute(
							types.AttributeKeyInflationDisableTime,
							suite.Ctx.BlockTime().Format(time.RFC3339),
						),
					),
				))
		}

		suite.Require().Equal(expectedMintState.Params.InflationMin, mintParams.InflationMin, msg+": expected mint inflation min to match state "+msgSuffix)
		suite.Require().Equal(expectedMintState.Params.InflationMax, mintParams.InflationMax, msg+": expected mint inflation max to match state "+msgSuffix)
		suite.Require().Equal(expectedKavadistState.Params.Active, kavadistParams.Active, msg+":expected kavadist active flag match state "+msgSuffix)
		suite.Require().Equal(expectedDistrState.Params.CommunityTax, distrParams.CommunityTax, msg+":expected x/distribution community tax to match state "+msgSuffix)
		suite.Require().Equal(expectedDisableTime, params.UpgradeTimeDisableInflation, msg+": "+disableTimeMsg)

		// we always check staking rewards per second matches the passed in expectation
		suite.Require().Equal(expectedStakingRewards, params.StakingRewardsPerSecond, msg+": "+"staking rewards per second to match "+msgSuffix)
		// we don't modify or zero out the initial rewards per second for upgrade time
		suite.Require().Equal(setStakingRewards, params.UpgradeTimeSetStakingRewardsPerSecond, msg+": "+"set staking rewards per second to match "+msgSuffix)
	}

	blockTime := suite.Ctx.BlockTime()
	testCases := []struct {
		name              string
		upgradeTime       time.Time
		setStakingRewards sdkmath.LegacyDec
		shouldUpgrade     bool
	}{
		{"zero upgrade time -- should not upgrade", time.Time{}, sdkmath.LegacyNewDec(1001), false},
		{"upgrade time in future -- should not upgrade", blockTime.Add(1 * time.Second), sdkmath.LegacyNewDec(1002), false},
		{"upgrade time in past -- should upgrade", blockTime.Add(-1 * time.Second), sdkmath.LegacyNewDec(1003), true},
		{"upgrade time equal to block time -- should upgrade", blockTime, sdkmath.LegacyNewDec(1004), true},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			params, found := suite.Keeper.GetParams(suite.Ctx)
			suite.Require().True(found)

			// these should not match in order to assure assertions test correct behavior
			suite.Require().NotEqual(params.StakingRewardsPerSecond, tc.setStakingRewards, "set staking rewards can not match initial staking rewards")

			// ensure state is as we expect before running upgrade or updating time
			validateState(false, time.Time{}, params.StakingRewardsPerSecond, params.UpgradeTimeSetStakingRewardsPerSecond, "initial state")

			// set inflation disable time
			params.UpgradeTimeDisableInflation = tc.upgradeTime
			// set upgrade time set staking rewards per second
			params.UpgradeTimeSetStakingRewardsPerSecond = tc.setStakingRewards
			suite.Keeper.SetParams(suite.Ctx, params)

			// run test function
			suite.testFunc(suite.Ctx, suite.Keeper)

			// run assertions to ensure upgrade did or did not run
			validateState(tc.shouldUpgrade, tc.upgradeTime, params.StakingRewardsPerSecond, tc.setStakingRewards, "first begin blocker run")

			// test idempotence only if upgrade should have been ran
			if tc.shouldUpgrade {
				// reset mint and kavadist state to their initial values
				suite.App.GetMintKeeper().SetParams(suite.Ctx, suite.genesisMintState.Params)
				suite.App.GetKavadistKeeper().SetParams(suite.Ctx, suite.genesisKavadistState.Params)

				// modify staking rewards per second to ensure they are not overridden again
				params, found := suite.Keeper.GetParams(suite.Ctx)
				suite.Require().True(found)
				params.StakingRewardsPerSecond = params.StakingRewardsPerSecond.Add(sdkmath.LegacyOneDec())
				suite.Keeper.SetParams(suite.Ctx, params)

				// run begin blocker again
				community.BeginBlocker(suite.Ctx, suite.Keeper)

				// ensure begin blocker is idempotent and never runs twice
				validateState(false, time.Time{}, params.StakingRewardsPerSecond, tc.setStakingRewards, "second begin blocker run")
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
