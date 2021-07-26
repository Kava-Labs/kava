package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/kava/app"
	cdpkeeper "github.com/kava-labs/kava/x/cdp/keeper"
	cdptypes "github.com/kava-labs/kava/x/cdp/types"
	"github.com/kava-labs/kava/x/incentive"
	"github.com/kava-labs/kava/x/incentive/keeper"
	"github.com/kava-labs/kava/x/incentive/testutil"
	"github.com/kava-labs/kava/x/incentive/types"
	"github.com/kava-labs/kava/x/kavadist"
)

type USDXIntegrationTests struct {
	testutil.IntegrationTester

	genesisTime time.Time
	addrs       []sdk.AccAddress
}

func TestUSDXIntegration(t *testing.T) {
	suite.Run(t, new(USDXIntegrationTests))
}

// SetupTest is run automatically before each suite test
func (suite *USDXIntegrationTests) SetupTest() {

	_, suite.addrs = app.GeneratePrivKeyAddressPairs(5)

	suite.genesisTime = time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC)
}

func (suite *USDXIntegrationTests) ProposeAndVoteOnNewRewardPeriods(committeeID uint64, voter sdk.AccAddress, newPeriods types.RewardPeriods) {
	suite.ProposeAndVoteOnNewParams(
		voter,
		committeeID,
		[]paramtypes.ParamChange{{
			Subspace: incentive.ModuleName,
			Key:      string(incentive.KeyUSDXMintingRewardPeriods),
			Value:    string(incentive.ModuleCdc.MustMarshalJSON(newPeriods)),
		}})
}

func (suite *USDXIntegrationTests) TestSingleUserAccumulatesRewardsAfterSyncing() {
	userA := suite.addrs[0]

	authBulder := app.NewAuthGenesisBuilder().
		WithSimpleModuleAccount(kavadist.ModuleName, cs(c(types.USDXMintingRewardDenom, 1e18))). // Fill kavadist with enough coins to pay out any reward
		WithSimpleAccount(userA, cs(c("bnb", 1e12)))                                             // give the user some coins

	incentBuilder := testutil.NewIncentiveGenesisBuilder().
		WithGenesisTime(suite.genesisTime).
		WithMultipliers(types.Multipliers{
			types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0")), // keep payout at 1.0 to make maths easier
		}).
		WithSimpleUSDXRewardPeriod("bnb-a", c(types.USDXMintingRewardDenom, 1e6))

	suite.StartChain(
		suite.genesisTime,
		NewPricefeedGenStateMultiFromTime(suite.genesisTime),
		NewCDPGenStateMulti(),
		authBulder.BuildMarshalled(),
		incentBuilder.BuildMarshalled(),
	)

	// User creates a CDP to begin earning rewards.
	suite.NoError(
		suite.DeliverMsgCreateCDP(userA, c("bnb", 1e10), c(cdptypes.DefaultStableDenom, 1e9), "bnb-a"),
	)

	// Let time pass to accumulate interest on the deposit
	// Use one long block instead of many to reduce any rounding errors, and speed up tests.
	suite.NextBlockAfter(1e6 * time.Second) // about 12 days

	// User repays and borrows just to sync their CDP
	suite.NoError(
		suite.DeliverCDPMsgRepay(userA, "bnb-a", c(cdptypes.DefaultStableDenom, 1)),
	)
	suite.NoError(
		suite.DeliverCDPMsgBorrow(userA, "bnb-a", c(cdptypes.DefaultStableDenom, 1)),
	)

	// Accumulate more rewards.
	// The user still has the same percentage of all CDP debt (100%) so their rewards should be the same as in the previous block.
	suite.NextBlockAfter(1e6 * time.Second) // about 12 days

	// User claims all their rewards
	suite.NoError(
		suite.DeliverIncentiveMsg(types.NewMsgClaimUSDXMintingReward(userA, "large")),
	)

	// The users has always had 100% of cdp debt, so they should receive all rewards for the previous two blocks.
	// Total rewards for each block is block duration * rewards per second
	accuracy := 1e-18 // using a very high accuracy to flag future small calculation changes
	suite.BalanceInEpsilon(userA, cs(c("bnb", 1e12-1e10), c(cdptypes.DefaultStableDenom, 1e9), c(types.USDXMintingRewardDenom, 2*1e6*1e6)), accuracy)
}

func (suite *USDXIntegrationTests) TestSingleUserAccumulatesRewardsWithoutSyncing() {

	user := suite.addrs[0]
	initialCollateral := c("bnb", 1e9)

	authBuilder := app.NewAuthGenesisBuilder().
		WithSimpleModuleAccount(kavadist.ModuleName, cs(c(types.USDXMintingRewardDenom, 1e18))). // Fill kavadist with enough coins to pay out any reward
		WithSimpleAccount(user, cs(initialCollateral))

	collateralType := "bnb-a"

	incentBuilder := testutil.NewIncentiveGenesisBuilder().
		WithGenesisTime(suite.genesisTime).
		WithMultipliers(types.Multipliers{
			types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0")), // keep payout at 1.0 to make maths easier
		}).
		WithSimpleUSDXRewardPeriod(collateralType, c(types.USDXMintingRewardDenom, 1e6))

	suite.StartChain(
		suite.genesisTime,
		authBuilder.BuildMarshalled(),
		NewPricefeedGenStateMultiFromTime(suite.genesisTime),
		NewCDPGenStateMulti(),
		incentBuilder.BuildMarshalled(),
	)

	// Setup cdp state containing one CDP
	suite.NoError(
		suite.DeliverMsgCreateCDP(user, initialCollateral, c("usdx", 1e8), collateralType),
	)

	// Skip ahead a few blocks blocks to accumulate both interest and usdx reward for the cdp
	// Don't sync the CDP between the blocks
	suite.NextBlockAfter(1e6 * time.Second) // about 12 days
	suite.NextBlockAfter(1e6 * time.Second)
	suite.NextBlockAfter(1e6 * time.Second)

	suite.NoError(
		suite.DeliverIncentiveMsg(types.NewMsgClaimUSDXMintingReward(user, "large")),
	)

	// The users has always had 100% of cdp debt, so they should receive all rewards for the previous two blocks.
	// Total rewards for each block is block duration * rewards per second
	accuracy := 1e-18 // using a very high accuracy to flag future small calculation changes
	suite.BalanceInEpsilon(user, cs(c(cdptypes.DefaultStableDenom, 1e8), c(types.USDXMintingRewardDenom, 3*1e6*1e6)), accuracy)
}

func (suite *USDXIntegrationTests) TestReinstatingRewardParamsDoesNotTriggerOverPayments() {

	userA := suite.addrs[0]
	userB := suite.addrs[1]

	authBuilder := app.NewAuthGenesisBuilder().
		WithSimpleModuleAccount(kavadist.ModuleName, cs(c(types.USDXMintingRewardDenom, 1e18))). // Fill kavadist with enough coins to pay out any reward
		WithSimpleAccount(userA, cs(c("bnb", 1e10))).
		WithSimpleAccount(userB, cs(c("bnb", 1e10)))

	incentBuilder := testutil.NewIncentiveGenesisBuilder().
		WithGenesisTime(suite.genesisTime).
		WithMultipliers(types.Multipliers{
			types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0")), // keep payout at 1.0 to make maths easier
		}).
		WithSimpleUSDXRewardPeriod("bnb-a", c(types.USDXMintingRewardDenom, 1e6))

	suite.StartChain(
		suite.genesisTime,
		authBuilder.BuildMarshalled(),
		NewPricefeedGenStateMultiFromTime(suite.genesisTime),
		NewCDPGenStateMulti(),
		incentBuilder.BuildMarshalled(),
		NewCommitteeGenesisState(0, userA), // create a committtee to change params
	)

	// Accumulate some CDP rewards, requires creating a cdp so the total borrowed isn't 0.
	suite.NoError(
		suite.DeliverMsgCreateCDP(userA, c("bnb", 1e10), c("usdx", 1e9), "bnb-a"),
	)
	suite.NextBlockAfter(1e6 * time.Second)

	// Remove the USDX reward period
	suite.ProposeAndVoteOnNewRewardPeriods(0, userA, types.RewardPeriods{})
	// next block so proposal is enacted
	suite.NextBlockAfter(1 * time.Second)

	// Create a CDP when there is no reward periods. In a previous version the claim object would not be created, leading to the bug.
	// Withdraw the same amount of usdx as the first cdp currently has. This make the reward maths easier, as rewards will be split 50:50 between each cdp.
	firstCDP, f := suite.App.GetCDPKeeper().GetCdpByOwnerAndCollateralType(suite.Ctx, userA, "bnb-a")
	suite.True(f)
	firstCDPTotalPrincipal := firstCDP.GetTotalPrincipal()
	suite.NoError(
		suite.DeliverMsgCreateCDP(userB, c("bnb", 1e10), firstCDPTotalPrincipal, "bnb-a"),
	)

	// Add back the reward period
	suite.ProposeAndVoteOnNewRewardPeriods(0, userA,
		types.RewardPeriods{types.NewRewardPeriod(
			true,
			"bnb-a",
			suite.Ctx.BlockTime(), // start accumulating again from this block
			suite.genesisTime.Add(365*24*time.Hour),
			c(types.USDXMintingRewardDenom, 1e6),
		)},
	)
	// next block so proposal is enacted
	suite.NextBlockAfter(1 * time.Second)

	// Sync the cdp and claim by borrowing a bit
	// In a previous version this would create the cdp with incorrect indexes, leading to overpayment.
	suite.NoError(
		suite.DeliverCDPMsgBorrow(userB, "bnb-a", c(cdptypes.DefaultStableDenom, 1)),
	)

	// Claim rewards
	suite.NoError(
		suite.DeliverIncentiveMsg(types.NewMsgClaimUSDXMintingReward(userB, "large")),
	)

	// The cdp had half the total borrows for a 1s block. So should earn half the rewards for that block
	suite.BalanceInEpsilon(
		userB,
		cs(firstCDPTotalPrincipal.Add(c(cdptypes.DefaultStableDenom, 1)), c(types.USDXMintingRewardDenom, 0.5*1e6)),
		1e-18, // using very high accuracy to catch small changes to the calculations
	)
}

// Test suite used for all keeper tests
type USDXRewardsTestSuite struct {
	suite.Suite

	keeper    keeper.Keeper
	cdpKeeper cdpkeeper.Keeper

	app app.TestApp
	ctx sdk.Context

	genesisTime time.Time
	addrs       []sdk.AccAddress
}

// SetupTest is run automatically before each suite test
func (suite *USDXRewardsTestSuite) SetupTest() {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)

	_, suite.addrs = app.GeneratePrivKeyAddressPairs(5)

	suite.genesisTime = time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC)
}

func (suite *USDXRewardsTestSuite) SetupApp() {
	suite.app = app.NewTestApp()

	suite.keeper = suite.app.GetIncentiveKeeper()
	suite.cdpKeeper = suite.app.GetCDPKeeper()

	suite.ctx = suite.app.NewContext(true, abci.Header{Height: 1, Time: suite.genesisTime})
}

func (suite *USDXRewardsTestSuite) SetupWithGenState(authBuilder app.AuthGenesisBuilder, incentBuilder testutil.IncentiveGenesisBuilder) {
	suite.SetupApp()

	suite.app.InitializeFromGenesisStatesWithTime(
		suite.genesisTime,
		authBuilder.BuildMarshalled(),
		NewPricefeedGenStateMultiFromTime(suite.genesisTime),
		NewCDPGenStateMulti(),
		incentBuilder.BuildMarshalled(),
	)
}

func (suite *USDXRewardsTestSuite) TestAccumulateUSDXMintingRewards() {
	type args struct {
		ctype                 string
		rewardsPerSecond      sdk.Coin
		initialTotalPrincipal sdk.Coin
		timeElapsed           int
		expectedRewardFactor  sdk.Dec
	}
	type test struct {
		name string
		args args
	}
	testCases := []test{
		{
			"7 seconds",
			args{
				ctype:                 "bnb-a",
				rewardsPerSecond:      c("ukava", 122354),
				initialTotalPrincipal: c("usdx", 1000000000000),
				timeElapsed:           7,
				expectedRewardFactor:  d("0.000000856478000000"),
			},
		},
		{
			"1 day",
			args{
				ctype:                 "bnb-a",
				rewardsPerSecond:      c("ukava", 122354),
				initialTotalPrincipal: c("usdx", 1000000000000),
				timeElapsed:           86400,
				expectedRewardFactor:  d("0.0105713856"),
			},
		},
		{
			"0 seconds",
			args{
				ctype:                 "bnb-a",
				rewardsPerSecond:      c("ukava", 122354),
				initialTotalPrincipal: c("usdx", 1000000000000),
				timeElapsed:           0,
				expectedRewardFactor:  d("0.0"),
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			incentBuilder := testutil.NewIncentiveGenesisBuilder().WithGenesisTime(suite.genesisTime).WithSimpleUSDXRewardPeriod(tc.args.ctype, tc.args.rewardsPerSecond)

			suite.SetupWithGenState(app.NewAuthGenesisBuilder(), incentBuilder)

			// setup cdp state
			suite.cdpKeeper.SetTotalPrincipal(suite.ctx, tc.args.ctype, cdptypes.DefaultStableDenom, tc.args.initialTotalPrincipal.Amount)

			updatedBlockTime := suite.ctx.BlockTime().Add(time.Duration(int(time.Second) * tc.args.timeElapsed))
			suite.ctx = suite.ctx.WithBlockTime(updatedBlockTime)
			rewardPeriod, found := suite.keeper.GetUSDXMintingRewardPeriod(suite.ctx, tc.args.ctype)
			suite.Require().True(found)
			suite.keeper.AccumulateUSDXMintingRewards(suite.ctx, rewardPeriod)

			rewardFactor, _ := suite.keeper.GetUSDXMintingRewardFactor(suite.ctx, tc.args.ctype)
			suite.Require().Equal(tc.args.expectedRewardFactor, rewardFactor)
		})
	}
}

func (suite *USDXRewardsTestSuite) TestSynchronizeUSDXMintingReward() {
	type args struct {
		ctype                string
		rewardsPerSecond     sdk.Coin
		initialCollateral    sdk.Coin
		initialPrincipal     sdk.Coin
		blockTimes           []int
		expectedRewardFactor sdk.Dec
		expectedRewards      sdk.Coin
	}
	type test struct {
		name string
		args args
	}

	testCases := []test{
		{
			"10 blocks",
			args{
				ctype:                "bnb-a",
				rewardsPerSecond:     c("ukava", 122354),
				initialCollateral:    c("bnb", 1000000000000),
				initialPrincipal:     c("usdx", 10000000000),
				blockTimes:           []int{10, 10, 10, 10, 10, 10, 10, 10, 10, 10},
				expectedRewardFactor: d("0.001223540000000000"),
				expectedRewards:      c("ukava", 12235400),
			},
		},
		{
			"10 blocks - long block time",
			args{
				ctype:                "bnb-a",
				rewardsPerSecond:     c("ukava", 122354),
				initialCollateral:    c("bnb", 1000000000000),
				initialPrincipal:     c("usdx", 10000000000),
				blockTimes:           []int{86400, 86400, 86400, 86400, 86400, 86400, 86400, 86400, 86400, 86400},
				expectedRewardFactor: d("10.57138560000000000"),
				expectedRewards:      c("ukava", 105713856000),
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			authBuilder := app.NewAuthGenesisBuilder().WithSimpleAccount(suite.addrs[0], cs(tc.args.initialCollateral))
			incentBuilder := testutil.NewIncentiveGenesisBuilder().WithGenesisTime(suite.genesisTime).WithSimpleUSDXRewardPeriod(tc.args.ctype, tc.args.rewardsPerSecond)

			suite.SetupWithGenState(authBuilder, incentBuilder)

			// setup cdp state
			err := suite.cdpKeeper.AddCdp(suite.ctx, suite.addrs[0], tc.args.initialCollateral, tc.args.initialPrincipal, tc.args.ctype)
			suite.Require().NoError(err)

			claim, found := suite.keeper.GetUSDXMintingClaim(suite.ctx, suite.addrs[0])
			suite.Require().True(found)
			suite.Require().Equal(sdk.ZeroDec(), claim.RewardIndexes[0].RewardFactor)

			var timeElapsed int
			previousBlockTime := suite.ctx.BlockTime()
			for _, t := range tc.args.blockTimes {
				timeElapsed += t
				updatedBlockTime := previousBlockTime.Add(time.Duration(int(time.Second) * t))
				previousBlockTime = updatedBlockTime
				blockCtx := suite.ctx.WithBlockTime(updatedBlockTime)
				rewardPeriod, found := suite.keeper.GetUSDXMintingRewardPeriod(blockCtx, tc.args.ctype)
				suite.Require().True(found)
				suite.keeper.AccumulateUSDXMintingRewards(blockCtx, rewardPeriod)
			}
			updatedBlockTime := suite.ctx.BlockTime().Add(time.Duration(int(time.Second) * timeElapsed))
			suite.ctx = suite.ctx.WithBlockTime(updatedBlockTime)
			cdp, found := suite.cdpKeeper.GetCdpByOwnerAndCollateralType(suite.ctx, suite.addrs[0], tc.args.ctype)
			suite.Require().True(found)
			suite.Require().NotPanics(func() {
				suite.keeper.SynchronizeUSDXMintingReward(suite.ctx, cdp)
			})

			rewardFactor, _ := suite.keeper.GetUSDXMintingRewardFactor(suite.ctx, tc.args.ctype)
			suite.Require().Equal(tc.args.expectedRewardFactor, rewardFactor)

			claim, found = suite.keeper.GetUSDXMintingClaim(suite.ctx, suite.addrs[0])
			suite.Require().True(found)
			suite.Require().Equal(tc.args.expectedRewardFactor, claim.RewardIndexes[0].RewardFactor)
			suite.Require().Equal(tc.args.expectedRewards, claim.Reward)
		})
	}
}

func (suite *USDXRewardsTestSuite) TestSimulateUSDXMintingRewardSynchronization() {
	type args struct {
		ctype                string
		rewardsPerSecond     sdk.Coin
		initialCollateral    sdk.Coin
		initialPrincipal     sdk.Coin
		blockTimes           []int
		expectedRewardFactor sdk.Dec
		expectedRewards      sdk.Coin
	}
	type test struct {
		name string
		args args
	}

	testCases := []test{
		{
			"10 blocks",
			args{
				ctype:                "bnb-a",
				rewardsPerSecond:     c("ukava", 122354),
				initialCollateral:    c("bnb", 1000000000000),
				initialPrincipal:     c("usdx", 10000000000),
				blockTimes:           []int{10, 10, 10, 10, 10, 10, 10, 10, 10, 10},
				expectedRewardFactor: d("0.001223540000000000"),
				expectedRewards:      c("ukava", 12235400),
			},
		},
		{
			"10 blocks - long block time",
			args{
				ctype:                "bnb-a",
				rewardsPerSecond:     c("ukava", 122354),
				initialCollateral:    c("bnb", 1000000000000),
				initialPrincipal:     c("usdx", 10000000000),
				blockTimes:           []int{86400, 86400, 86400, 86400, 86400, 86400, 86400, 86400, 86400, 86400},
				expectedRewardFactor: d("10.57138560000000000"),
				expectedRewards:      c("ukava", 105713856000),
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			authBuilder := app.NewAuthGenesisBuilder().WithSimpleAccount(suite.addrs[0], cs(tc.args.initialCollateral))
			incentBuilder := testutil.NewIncentiveGenesisBuilder().WithGenesisTime(suite.genesisTime).WithSimpleUSDXRewardPeriod(tc.args.ctype, tc.args.rewardsPerSecond)

			suite.SetupWithGenState(authBuilder, incentBuilder)

			// setup cdp state
			err := suite.cdpKeeper.AddCdp(suite.ctx, suite.addrs[0], tc.args.initialCollateral, tc.args.initialPrincipal, tc.args.ctype)
			suite.Require().NoError(err)

			claim, found := suite.keeper.GetUSDXMintingClaim(suite.ctx, suite.addrs[0])
			suite.Require().True(found)
			suite.Require().Equal(sdk.ZeroDec(), claim.RewardIndexes[0].RewardFactor)

			var timeElapsed int
			previousBlockTime := suite.ctx.BlockTime()
			for _, t := range tc.args.blockTimes {
				timeElapsed += t
				updatedBlockTime := previousBlockTime.Add(time.Duration(int(time.Second) * t))
				previousBlockTime = updatedBlockTime
				blockCtx := suite.ctx.WithBlockTime(updatedBlockTime)
				rewardPeriod, found := suite.keeper.GetUSDXMintingRewardPeriod(blockCtx, tc.args.ctype)
				suite.Require().True(found)
				suite.keeper.AccumulateUSDXMintingRewards(blockCtx, rewardPeriod)
			}
			updatedBlockTime := suite.ctx.BlockTime().Add(time.Duration(int(time.Second) * timeElapsed))
			suite.ctx = suite.ctx.WithBlockTime(updatedBlockTime)

			claim, found = suite.keeper.GetUSDXMintingClaim(suite.ctx, suite.addrs[0])
			suite.Require().True(found)
			suite.Require().Equal(claim.RewardIndexes[0].RewardFactor, sdk.ZeroDec())
			suite.Require().Equal(claim.Reward, sdk.NewCoin("ukava", sdk.ZeroInt()))

			updatedClaim := suite.keeper.SimulateUSDXMintingSynchronization(suite.ctx, claim)
			suite.Require().Equal(tc.args.expectedRewardFactor, updatedClaim.RewardIndexes[0].RewardFactor)
			suite.Require().Equal(tc.args.expectedRewards, updatedClaim.Reward)
		})
	}
}

func TestUSDXRewardsTestSuite(t *testing.T) {
	suite.Run(t, new(USDXRewardsTestSuite))
}
