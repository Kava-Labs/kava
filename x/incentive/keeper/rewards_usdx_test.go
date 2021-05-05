package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	cdpkeeper "github.com/kava-labs/kava/x/cdp/keeper"
	cdptypes "github.com/kava-labs/kava/x/cdp/types"
	"github.com/kava-labs/kava/x/incentive/keeper"
)

// Test suite used for all keeper tests
type USDXRewardsTestSuite struct {
	suite.Suite

	keeper    keeper.Keeper
	cdpKeeper cdpkeeper.Keeper

	app   app.TestApp
	ctx   sdk.Context
	addrs []sdk.AccAddress
}

// SetupTest is run automatically before each suite test
func (suite *USDXRewardsTestSuite) SetupTest() {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)

	_, suite.addrs = app.GeneratePrivKeyAddressPairs(5)
}

func (suite *USDXRewardsTestSuite) SetupApp() {
	suite.app = app.NewTestApp()

	suite.keeper = suite.app.GetIncentiveKeeper()
	suite.cdpKeeper = suite.app.GetCDPKeeper()

	suite.ctx = suite.app.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
}

func (suite *USDXRewardsTestSuite) SetupWithGenState(authBuilder AuthGenesisBuilder, incentBuilder incentiveGenesisBuilder) {
	suite.SetupApp()

	suite.app.InitializeFromGenesisStates(
		authBuilder.BuildMarshalled(),
		NewPricefeedGenStateMulti(),
		NewCDPGenStateMulti(),
		incentBuilder.buildMarshalled(),
	)
}

func (suite *USDXRewardsTestSuite) TestAccumulateUSDXMintingRewards() {
	type args struct {
		ctype                 string
		rewardsPerSecond      sdk.Coin
		initialTime           time.Time
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
				initialTime:           time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
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
				initialTime:           time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
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
				initialTime:           time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				initialTotalPrincipal: c("usdx", 1000000000000),
				timeElapsed:           0,
				expectedRewardFactor:  d("0.0"),
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			incentBuilder := newIncentiveGenesisBuilder().withGenesisTime(tc.args.initialTime).withSimpleUSDXRewardPeriod(tc.args.ctype, tc.args.rewardsPerSecond)

			suite.SetupWithGenState(NewAuthGenesisBuilder(), incentBuilder)
			suite.ctx = suite.ctx.WithBlockTime(tc.args.initialTime)

			// setup cdp state
			suite.cdpKeeper.SetTotalPrincipal(suite.ctx, tc.args.ctype, cdptypes.DefaultStableDenom, tc.args.initialTotalPrincipal.Amount)

			updatedBlockTime := suite.ctx.BlockTime().Add(time.Duration(int(time.Second) * tc.args.timeElapsed))
			suite.ctx = suite.ctx.WithBlockTime(updatedBlockTime)
			rewardPeriod, found := suite.keeper.GetUSDXMintingRewardPeriod(suite.ctx, tc.args.ctype)
			suite.Require().True(found)
			err := suite.keeper.AccumulateUSDXMintingRewards(suite.ctx, rewardPeriod)
			suite.Require().NoError(err)

			rewardFactor, _ := suite.keeper.GetUSDXMintingRewardFactor(suite.ctx, tc.args.ctype)
			suite.Require().Equal(tc.args.expectedRewardFactor, rewardFactor)
		})
	}
}

func (suite *USDXRewardsTestSuite) TestSynchronizeUSDXMintingReward() {
	type args struct {
		ctype                string
		rewardsPerSecond     sdk.Coin
		initialTime          time.Time
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
				initialTime:          time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
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
				initialTime:          time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
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
			authBuilder := NewAuthGenesisBuilder().WithSimpleAccount(suite.addrs[0], cs(tc.args.initialCollateral))
			incentBuilder := newIncentiveGenesisBuilder().withGenesisTime(tc.args.initialTime).withSimpleUSDXRewardPeriod(tc.args.ctype, tc.args.rewardsPerSecond)

			suite.SetupWithGenState(authBuilder, incentBuilder)
			suite.ctx = suite.ctx.WithBlockTime(tc.args.initialTime)

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
				err := suite.keeper.AccumulateUSDXMintingRewards(blockCtx, rewardPeriod)
				suite.Require().NoError(err)
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
		initialTime          time.Time
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
				initialTime:          time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
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
				initialTime:          time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
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
			authBuilder := NewAuthGenesisBuilder().WithSimpleAccount(suite.addrs[0], cs(tc.args.initialCollateral))
			incentBuilder := newIncentiveGenesisBuilder().withGenesisTime(tc.args.initialTime).withSimpleUSDXRewardPeriod(tc.args.ctype, tc.args.rewardsPerSecond)

			suite.SetupWithGenState(authBuilder, incentBuilder)
			suite.ctx = suite.ctx.WithBlockTime(tc.args.initialTime)

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
				err := suite.keeper.AccumulateUSDXMintingRewards(blockCtx, rewardPeriod)
				suite.Require().NoError(err)
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
