package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/committee"
	committeekeeper "github.com/kava-labs/kava/x/committee/keeper"
	"github.com/kava-labs/kava/x/hard"
	hardkeeper "github.com/kava-labs/kava/x/hard/keeper"
	"github.com/kava-labs/kava/x/incentive/keeper"
	"github.com/kava-labs/kava/x/incentive/types"
)

// Test suite used for all keeper tests
type SupplyRewardsTestSuite struct {
	suite.Suite

	keeper          keeper.Keeper
	hardKeeper      hardkeeper.Keeper
	committeeKeeper committeekeeper.Keeper
	app             app.TestApp
	ctx             sdk.Context
	addrs           []sdk.AccAddress
}

// SetupTest is run automatically before each suite test
func (suite *SupplyRewardsTestSuite) SetupTest() {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)

	_, suite.addrs = app.GeneratePrivKeyAddressPairs(5)
}

func (suite *SupplyRewardsTestSuite) SetupApp() {
	suite.app = app.NewTestApp()

	suite.keeper = suite.app.GetIncentiveKeeper()
	suite.hardKeeper = suite.app.GetHardKeeper()
	suite.committeeKeeper = suite.app.GetCommitteeKeeper()

	suite.ctx = suite.app.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
}

func (suite *SupplyRewardsTestSuite) SetupWithGenState() {
	suite.SetupApp()

	suite.app.InitializeFromGenesisStates(
		NewAuthGenState(suite.addrs, cs(c("ukava", 1_000_000_000))),
		NewPricefeedGenStateMulti(),
		NewHardGenStateMulti(),
		NewCommitteeGenesisState(suite.addrs[:2]), // TODO add committee members to suite
	)
}

func (suite *SupplyRewardsTestSuite) TestAccumulateHardSupplyRewards() {
	type args struct {
		deposit               sdk.Coin
		rewardsPerSecond      sdk.Coins
		initialTime           time.Time
		timeElapsed           int
		expectedRewardIndexes types.RewardIndexes
	}
	type test struct {
		name string
		args args
	}
	testCases := []test{
		{
			"single reward denom: 7 seconds",
			args{
				deposit:               c("bnb", 1000000000000),
				rewardsPerSecond:      cs(c("hard", 122354)),
				initialTime:           time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				timeElapsed:           7,
				expectedRewardIndexes: types.RewardIndexes{types.NewRewardIndex("hard", d("0.000000856478000000"))},
			},
		},
		{
			"single reward denom: 1 day",
			args{
				deposit:               c("bnb", 1000000000000),
				rewardsPerSecond:      cs(c("hard", 122354)),
				initialTime:           time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				timeElapsed:           86400,
				expectedRewardIndexes: types.RewardIndexes{types.NewRewardIndex("hard", d("0.010571385600000000"))},
			},
		},
		{
			"single reward denom: 0 seconds",
			args{
				deposit:               c("bnb", 1000000000000),
				rewardsPerSecond:      cs(c("hard", 122354)),
				initialTime:           time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				timeElapsed:           0,
				expectedRewardIndexes: types.RewardIndexes{types.NewRewardIndex("hard", d("0.0"))},
			},
		},
		{
			"multiple reward denoms: 7 seconds",
			args{
				deposit:          c("bnb", 1000000000000),
				rewardsPerSecond: cs(c("hard", 122354), c("ukava", 122354)),
				initialTime:      time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				timeElapsed:      7,
				expectedRewardIndexes: types.RewardIndexes{
					types.NewRewardIndex("hard", d("0.000000856478000000")),
					types.NewRewardIndex("ukava", d("0.000000856478000000")),
				},
			},
		},
		{
			"multiple reward denoms: 1 day",
			args{
				deposit:          c("bnb", 1000000000000),
				rewardsPerSecond: cs(c("hard", 122354), c("ukava", 122354)),
				initialTime:      time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				timeElapsed:      86400,
				expectedRewardIndexes: types.RewardIndexes{
					types.NewRewardIndex("hard", d("0.010571385600000000")),
					types.NewRewardIndex("ukava", d("0.010571385600000000")),
				},
			},
		},
		{
			"multiple reward denoms: 0 seconds",
			args{
				deposit:          c("bnb", 1000000000000),
				rewardsPerSecond: cs(c("hard", 122354), c("ukava", 122354)),
				initialTime:      time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				timeElapsed:      0,
				expectedRewardIndexes: types.RewardIndexes{
					types.NewRewardIndex("hard", d("0.0")),
					types.NewRewardIndex("ukava", d("0.0")),
				},
			},
		},
		{
			"multiple reward denoms with different rewards per second: 1 day",
			args{
				deposit:          c("bnb", 1000000000000),
				rewardsPerSecond: cs(c("hard", 122354), c("ukava", 555555)),
				initialTime:      time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				timeElapsed:      86400,
				expectedRewardIndexes: types.RewardIndexes{
					types.NewRewardIndex("hard", d("0.010571385600000000")),
					types.NewRewardIndex("ukava", d("0.047999952000000000")),
				},
			},
		},
		{
			"single reward denom, no rewards",
			args{
				deposit:               c("bnb", 1000000000000),
				rewardsPerSecond:      sdk.Coins{},
				initialTime:           time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				timeElapsed:           7,
				expectedRewardIndexes: types.RewardIndexes{},
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupWithGenState()
			suite.ctx = suite.ctx.WithBlockTime(tc.args.initialTime)

			// Set up incentive state
			params := types.NewParams(
				types.RewardPeriods{types.NewRewardPeriod(true, tc.args.deposit.Denom, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), c("hard", 1))},
				types.MultiRewardPeriods{types.NewMultiRewardPeriod(true, tc.args.deposit.Denom, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), tc.args.rewardsPerSecond)},
				types.MultiRewardPeriods{types.NewMultiRewardPeriod(true, tc.args.deposit.Denom, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), tc.args.rewardsPerSecond)},
				types.RewardPeriods{types.NewRewardPeriod(true, tc.args.deposit.Denom, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), c("hard", 1))},
				types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
				tc.args.initialTime.Add(time.Hour*24*365*5),
			)
			suite.keeper.SetParams(suite.ctx, params)
			suite.keeper.SetPreviousHardSupplyRewardAccrualTime(suite.ctx, tc.args.deposit.Denom, tc.args.initialTime)
			var rewardIndexes types.RewardIndexes
			for _, rewardCoin := range tc.args.rewardsPerSecond {
				rewardIndex := types.NewRewardIndex(rewardCoin.Denom, sdk.ZeroDec())
				rewardIndexes = append(rewardIndexes, rewardIndex)
			}
			if len(rewardIndexes) > 0 {
				suite.keeper.SetHardSupplyRewardIndexes(suite.ctx, tc.args.deposit.Denom, rewardIndexes)
			}

			// Set up hard state (interest factor for the relevant denom)
			suite.hardKeeper.SetSupplyInterestFactor(suite.ctx, tc.args.deposit.Denom, sdk.MustNewDecFromStr("1.0"))
			suite.hardKeeper.SetPreviousAccrualTime(suite.ctx, tc.args.deposit.Denom, tc.args.initialTime)

			// User deposits to increase total supplied amount
			userAddr := suite.addrs[3]
			err := suite.hardKeeper.Deposit(suite.ctx, userAddr, sdk.NewCoins(tc.args.deposit))
			suite.Require().NoError(err)

			// Set up chain context at future time
			runAtTime := suite.ctx.BlockTime().Add(time.Duration(int(time.Second) * tc.args.timeElapsed))
			runCtx := suite.ctx.WithBlockTime(runAtTime)

			// Run Hard begin blocker in order to update the denom's index factor
			hard.BeginBlocker(runCtx, suite.hardKeeper)

			// Accumulate hard supply rewards for the deposit denom
			multiRewardPeriod, found := suite.keeper.GetHardSupplyRewardPeriods(runCtx, tc.args.deposit.Denom)
			suite.Require().True(found)
			err = suite.keeper.AccumulateHardSupplyRewards(runCtx, multiRewardPeriod)
			suite.Require().NoError(err)

			// Check that each expected reward index matches the current stored reward index for the denom
			globalRewardIndexes, found := suite.keeper.GetHardSupplyRewardIndexes(runCtx, tc.args.deposit.Denom)
			if len(tc.args.rewardsPerSecond) > 0 {
				suite.Require().True(found)
				for _, expectedRewardIndex := range tc.args.expectedRewardIndexes {
					globalRewardIndex, found := globalRewardIndexes.GetRewardIndex(expectedRewardIndex.CollateralType)
					suite.Require().True(found)
					suite.Require().Equal(expectedRewardIndex, globalRewardIndex)
				}
			} else {
				suite.Require().False(found)
			}

		})
	}
}

func (suite *SupplyRewardsTestSuite) TestInitializeHardSupplyRewards() {

	type args struct {
		moneyMarketRewardDenoms          map[string][]string
		deposit                          sdk.Coins
		initialTime                      time.Time
		expectedClaimSupplyRewardIndexes types.MultiRewardIndexes
	}
	type test struct {
		name string
		args args
	}

	standardMoneyMarketRewardDenoms := map[string][]string{
		"bnb":  {"hard"},
		"btcb": {"hard", "ukava"},
		"xrp":  {},
	}

	testCases := []test{
		{
			"single deposit denom, single reward denom",
			args{
				moneyMarketRewardDenoms: standardMoneyMarketRewardDenoms,
				deposit:                 cs(c("bnb", 1000000000000)),
				initialTime:             time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				expectedClaimSupplyRewardIndexes: types.MultiRewardIndexes{
					types.NewMultiRewardIndex(
						"bnb",
						types.RewardIndexes{
							types.NewRewardIndex("hard", d("0.0")),
						},
					),
				},
			},
		},
		{
			"single deposit denom, multiple reward denoms",
			args{
				moneyMarketRewardDenoms: standardMoneyMarketRewardDenoms,
				deposit:                 cs(c("btcb", 1000000000000)),
				initialTime:             time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				expectedClaimSupplyRewardIndexes: types.MultiRewardIndexes{
					types.NewMultiRewardIndex(
						"btcb",
						types.RewardIndexes{
							types.NewRewardIndex("hard", d("0.0")),
							types.NewRewardIndex("ukava", d("0.0")),
						},
					),
				},
			},
		},
		{
			"single deposit denom, no reward denoms",
			args{
				moneyMarketRewardDenoms: standardMoneyMarketRewardDenoms,
				deposit:                 cs(c("xrp", 1000000000000)),
				initialTime:             time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				expectedClaimSupplyRewardIndexes: types.MultiRewardIndexes{
					types.NewMultiRewardIndex(
						"xrp",
						nil,
					),
				},
			},
		},
		{
			"multiple deposit denoms, multiple overlapping reward denoms",
			args{
				moneyMarketRewardDenoms: standardMoneyMarketRewardDenoms,
				deposit:                 cs(c("bnb", 1000000000000), c("btcb", 1000000000000)),
				initialTime:             time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				expectedClaimSupplyRewardIndexes: types.MultiRewardIndexes{
					types.NewMultiRewardIndex(
						"bnb",
						types.RewardIndexes{
							types.NewRewardIndex("hard", d("0.0")),
						},
					),
					types.NewMultiRewardIndex(
						"btcb",
						types.RewardIndexes{
							types.NewRewardIndex("hard", d("0.0")),
							types.NewRewardIndex("ukava", d("0.0")),
						},
					),
				},
			},
		},
		{
			"multiple deposit denoms, correct discrete reward denoms",
			args{
				moneyMarketRewardDenoms: standardMoneyMarketRewardDenoms,
				deposit:                 cs(c("bnb", 1000000000000), c("xrp", 1000000000000)),
				initialTime:             time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				expectedClaimSupplyRewardIndexes: types.MultiRewardIndexes{
					types.NewMultiRewardIndex(
						"bnb",
						types.RewardIndexes{
							types.NewRewardIndex("hard", d("0.0")),
						},
					),
					types.NewMultiRewardIndex(
						"xrp",
						nil,
					),
				},
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupWithGenState()
			suite.ctx = suite.ctx.WithBlockTime(tc.args.initialTime)

			userAddr := suite.addrs[3]

			// Prepare money market + reward params
			i := 0
			var multiRewardPeriods types.MultiRewardPeriods
			var rewardPeriods types.RewardPeriods
			for moneyMarketDenom, rewardDenoms := range tc.args.moneyMarketRewardDenoms {
				// Set up multi reward periods for supply/borrow indexes with dynamic money market denoms/reward denoms
				var rewardsPerSecond sdk.Coins
				for _, rewardDenom := range rewardDenoms {
					rewardsPerSecond = append(rewardsPerSecond, sdk.NewCoin(rewardDenom, sdk.OneInt()))
				}
				multiRewardPeriod := types.NewMultiRewardPeriod(true, moneyMarketDenom, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), rewardsPerSecond)
				multiRewardPeriods = append(multiRewardPeriods, multiRewardPeriod)

				// Set up generic reward periods for usdx minting/delegator indexes
				if i == 0 && len(rewardDenoms) > 0 {
					rewardPeriod := types.NewRewardPeriod(true, moneyMarketDenom, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), rewardsPerSecond[i])
					rewardPeriods = append(rewardPeriods, rewardPeriod)
					i++
				}
			}

			// Initialize and set incentive params
			params := types.NewParams(
				rewardPeriods, multiRewardPeriods, multiRewardPeriods, rewardPeriods,
				types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
				tc.args.initialTime.Add(time.Hour*24*365*5),
			)
			suite.keeper.SetParams(suite.ctx, params)

			// Set each money market's previous accrual time and supply reward indexes
			for moneyMarketDenom, rewardDenoms := range tc.args.moneyMarketRewardDenoms {
				var rewardIndexes types.RewardIndexes
				for _, rewardDenom := range rewardDenoms {
					rewardIndex := types.NewRewardIndex(rewardDenom, sdk.ZeroDec())
					rewardIndexes = append(rewardIndexes, rewardIndex)
				}
				suite.keeper.SetPreviousHardSupplyRewardAccrualTime(suite.ctx, moneyMarketDenom, tc.args.initialTime)
				if len(rewardIndexes) > 0 {
					suite.keeper.SetHardSupplyRewardIndexes(suite.ctx, moneyMarketDenom, rewardIndexes)
				}
			}

			// User deposits
			err := suite.hardKeeper.Deposit(suite.ctx, userAddr, tc.args.deposit)
			suite.Require().NoError(err)

			claim, foundClaim := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, userAddr)
			suite.Require().True(foundClaim)
			suite.Require().Equal(tc.args.expectedClaimSupplyRewardIndexes, claim.SupplyRewardIndexes)
		})
	}
}

func (suite *SupplyRewardsTestSuite) TestSynchronizeHardSupplyReward() {
	type args struct {
		incentiveSupplyRewardDenom   string
		deposit                      sdk.Coin
		rewardsPerSecond             sdk.Coins
		initialTime                  time.Time
		blockTimes                   []int
		expectedRewardIndexes        types.RewardIndexes
		expectedRewards              sdk.Coins
		updateRewardsViaCommmittee   bool
		updatedBaseDenom             string
		updatedRewardsPerSecond      sdk.Coins
		updatedExpectedRewardIndexes types.RewardIndexes
		updatedExpectedRewards       sdk.Coins
		updatedTimeDuration          int
	}
	type test struct {
		name string
		args args
	}

	testCases := []test{
		{
			"single reward denom: 10 blocks",
			args{
				incentiveSupplyRewardDenom: "bnb",
				deposit:                    c("bnb", 10000000000),
				rewardsPerSecond:           cs(c("hard", 122354)),
				initialTime:                time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				blockTimes:                 []int{10, 10, 10, 10, 10, 10, 10, 10, 10, 10},
				expectedRewardIndexes:      types.RewardIndexes{types.NewRewardIndex("hard", d("0.001223540000000000"))},
				expectedRewards:            cs(c("hard", 12235400)),
				updateRewardsViaCommmittee: false,
			},
		},
		{
			"single reward denom: 10 blocks - long block time",
			args{
				incentiveSupplyRewardDenom: "bnb",
				deposit:                    c("bnb", 10000000000),
				rewardsPerSecond:           cs(c("hard", 122354)),
				initialTime:                time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				blockTimes:                 []int{86400, 86400, 86400, 86400, 86400, 86400, 86400, 86400, 86400, 86400},
				expectedRewardIndexes:      types.RewardIndexes{types.NewRewardIndex("hard", d("10.571385600000000000"))},
				expectedRewards:            cs(c("hard", 105713856000)),
				updateRewardsViaCommmittee: false,
			},
		},
		{
			"single reward denom: user reward index updated when reward is zero",
			args{
				incentiveSupplyRewardDenom: "ukava",
				deposit:                    c("ukava", 1),
				rewardsPerSecond:           cs(c("hard", 122354)),
				initialTime:                time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				blockTimes:                 []int{10, 10, 10, 10, 10, 10, 10, 10, 10, 10},
				expectedRewardIndexes:      types.RewardIndexes{types.NewRewardIndex("hard", d("0.122353998776460010"))},
				expectedRewards:            cs(),
				updateRewardsViaCommmittee: false,
			},
		},
		{
			"multiple reward denoms: 10 blocks",
			args{
				incentiveSupplyRewardDenom: "bnb",
				deposit:                    c("bnb", 10000000000),
				rewardsPerSecond:           cs(c("hard", 122354), c("ukava", 122354)),
				initialTime:                time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				blockTimes:                 []int{10, 10, 10, 10, 10, 10, 10, 10, 10, 10},
				expectedRewardIndexes: types.RewardIndexes{
					types.NewRewardIndex("hard", d("0.001223540000000000")),
					types.NewRewardIndex("ukava", d("0.001223540000000000")),
				},
				expectedRewards:            cs(c("hard", 12235400), c("ukava", 12235400)),
				updateRewardsViaCommmittee: false,
			},
		},
		{
			"multiple reward denoms: 10 blocks - long block time",
			args{
				incentiveSupplyRewardDenom: "bnb",
				deposit:                    c("bnb", 10000000000),
				rewardsPerSecond:           cs(c("hard", 122354), c("ukava", 122354)),
				initialTime:                time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				blockTimes:                 []int{86400, 86400, 86400, 86400, 86400, 86400, 86400, 86400, 86400, 86400},
				expectedRewardIndexes: types.RewardIndexes{
					types.NewRewardIndex("hard", d("10.571385600000000000")),
					types.NewRewardIndex("ukava", d("10.571385600000000000")),
				},
				expectedRewards:            cs(c("hard", 105713856000), c("ukava", 105713856000)),
				updateRewardsViaCommmittee: false,
			},
		},
		{
			"multiple reward denoms with different rewards per second: 10 blocks",
			args{
				incentiveSupplyRewardDenom: "bnb",
				deposit:                    c("bnb", 10000000000),
				rewardsPerSecond:           cs(c("hard", 122354), c("ukava", 555555)),
				initialTime:                time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				blockTimes:                 []int{10, 10, 10, 10, 10, 10, 10, 10, 10, 10},
				expectedRewardIndexes: types.RewardIndexes{
					types.NewRewardIndex("hard", d("0.001223540000000000")),
					types.NewRewardIndex("ukava", d("0.005555550000000000")),
				},
				expectedRewards:            cs(c("hard", 12235400), c("ukava", 55555500)),
				updateRewardsViaCommmittee: false,
			},
		},
		{
			"denom is in incentive's hard supply reward params but it has no rewards; add reward",
			args{
				incentiveSupplyRewardDenom: "bnb",
				deposit:                    c("bnb", 10000000000),
				rewardsPerSecond:           sdk.Coins{},
				initialTime:                time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				blockTimes:                 []int{100},
				expectedRewardIndexes:      types.RewardIndexes{},
				expectedRewards:            sdk.Coins{},
				updateRewardsViaCommmittee: true,
				updatedBaseDenom:           "bnb",
				updatedRewardsPerSecond:    cs(c("hard", 100000)),
				updatedExpectedRewards:     cs(c("hard", 8640000000)),
				updatedExpectedRewardIndexes: types.RewardIndexes{
					types.NewRewardIndex("hard", d("0.864")),
				},
				updatedTimeDuration: 86400,
			},
		},
		{
			"denom is in incentive's hard supply reward params and has rewards; add new reward type",
			args{
				incentiveSupplyRewardDenom: "bnb",
				deposit:                    c("bnb", 10000000000),
				rewardsPerSecond:           cs(c("hard", 122354)),
				initialTime:                time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				blockTimes:                 []int{86400},
				expectedRewardIndexes: types.RewardIndexes{
					types.NewRewardIndex("hard", d("1.057138560000000000")),
				},
				expectedRewards:            cs(c("hard", 10571385600)),
				updateRewardsViaCommmittee: true,
				updatedBaseDenom:           "bnb",
				updatedRewardsPerSecond:    cs(c("hard", 122354), c("ukava", 100000)),
				updatedExpectedRewards:     cs(c("hard", 21142771200), c("ukava", 8640000000)),
				updatedExpectedRewardIndexes: types.RewardIndexes{
					types.NewRewardIndex("hard", d("2.114277120000000000")),
					types.NewRewardIndex("ukava", d("0.864000000000000000")),
				},
				updatedTimeDuration: 86400,
			},
		},
		{
			"denom is in hard's money market params but not in incentive's hard supply reward params; add reward",
			args{
				incentiveSupplyRewardDenom: "bnb",
				deposit:                    c("zzz", 10000000000),
				rewardsPerSecond:           sdk.Coins{},
				initialTime:                time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				blockTimes:                 []int{100},
				expectedRewardIndexes:      types.RewardIndexes{},
				expectedRewards:            sdk.Coins{},
				updateRewardsViaCommmittee: true,
				updatedBaseDenom:           "zzz",
				updatedRewardsPerSecond:    cs(c("hard", 100000)),
				updatedExpectedRewards:     cs(c("hard", 8640000000)),
				updatedExpectedRewardIndexes: types.RewardIndexes{
					types.NewRewardIndex("hard", d("0.864")),
				},
				updatedTimeDuration: 86400,
			},
		},
		{
			"denom incentive's hard supply reward params but it has no rewards; add multiple reward types",
			args{
				incentiveSupplyRewardDenom: "bnb",
				deposit:                    c("bnb", 10000000000),
				rewardsPerSecond:           sdk.Coins{},
				initialTime:                time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				blockTimes:                 []int{100},
				expectedRewardIndexes:      types.RewardIndexes{},
				expectedRewards:            sdk.Coins{},
				updateRewardsViaCommmittee: true,
				updatedBaseDenom:           "bnb",
				updatedRewardsPerSecond:    cs(c("hard", 100000), c("ukava", 100500), c("swap", 500)),
				updatedExpectedRewards:     cs(c("hard", 8640000000), c("ukava", 8683200000), c("swap", 43200000)),
				updatedExpectedRewardIndexes: types.RewardIndexes{
					types.NewRewardIndex("hard", d("0.864")),
					types.NewRewardIndex("ukava", d("0.86832")),
					types.NewRewardIndex("swap", d("0.00432")),
				},
				updatedTimeDuration: 86400,
			},
		},
		{
			"denom is in hard's money market params but not in incentive's hard supply reward params; add multiple reward types",
			args{
				incentiveSupplyRewardDenom: "bnb",
				deposit:                    c("zzz", 10000000000),
				rewardsPerSecond:           sdk.Coins{},
				initialTime:                time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				blockTimes:                 []int{100},
				expectedRewardIndexes:      types.RewardIndexes{},
				expectedRewards:            sdk.Coins{},
				updateRewardsViaCommmittee: true,
				updatedBaseDenom:           "zzz",
				updatedRewardsPerSecond:    cs(c("hard", 100000), c("ukava", 100500), c("swap", 500)),
				updatedExpectedRewards:     cs(c("hard", 8640000000), c("ukava", 8683200000), c("swap", 43200000)),
				updatedExpectedRewardIndexes: types.RewardIndexes{
					types.NewRewardIndex("hard", d("0.864")),
					types.NewRewardIndex("ukava", d("0.86832")),
					types.NewRewardIndex("swap", d("0.00432")),
				},
				updatedTimeDuration: 86400,
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupWithGenState()
			suite.ctx = suite.ctx.WithBlockTime(tc.args.initialTime)

			// Set up incentive state
			incentiveParams := types.NewParams(
				types.RewardPeriods{types.NewRewardPeriod(true, tc.args.incentiveSupplyRewardDenom, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), c("hard", 1))},
				types.MultiRewardPeriods{types.NewMultiRewardPeriod(true, tc.args.incentiveSupplyRewardDenom, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), tc.args.rewardsPerSecond)},
				types.MultiRewardPeriods{types.NewMultiRewardPeriod(true, tc.args.incentiveSupplyRewardDenom, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), tc.args.rewardsPerSecond)},
				types.RewardPeriods{types.NewRewardPeriod(true, tc.args.incentiveSupplyRewardDenom, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), c("hard", 1))},
				types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
				tc.args.initialTime.Add(time.Hour*24*365*5),
			)
			suite.keeper.SetParams(suite.ctx, incentiveParams)
			suite.keeper.SetPreviousHardSupplyRewardAccrualTime(suite.ctx, tc.args.incentiveSupplyRewardDenom, tc.args.initialTime)
			var rewardIndexes types.RewardIndexes
			for _, rewardCoin := range tc.args.rewardsPerSecond {
				rewardIndex := types.NewRewardIndex(rewardCoin.Denom, sdk.ZeroDec())
				rewardIndexes = append(rewardIndexes, rewardIndex)
			}
			if len(rewardIndexes) > 0 {
				suite.keeper.SetHardSupplyRewardIndexes(suite.ctx, tc.args.incentiveSupplyRewardDenom, rewardIndexes)
			}

			// Set up hard state (interest factor for the relevant denom)
			suite.hardKeeper.SetSupplyInterestFactor(suite.ctx, tc.args.incentiveSupplyRewardDenom, sdk.MustNewDecFromStr("1.0"))
			suite.hardKeeper.SetBorrowInterestFactor(suite.ctx, tc.args.incentiveSupplyRewardDenom, sdk.MustNewDecFromStr("1.0"))
			suite.hardKeeper.SetPreviousAccrualTime(suite.ctx, tc.args.incentiveSupplyRewardDenom, tc.args.initialTime)

			// Deposit a fixed amount from another user to dilute primary user's rewards per second.
			suite.Require().NoError(
				suite.hardKeeper.Deposit(suite.ctx, suite.addrs[2], cs(c("ukava", 100_000_000))),
			)

			// User deposits and borrows to increase total borrowed amount
			userAddr := suite.addrs[3]
			err := suite.hardKeeper.Deposit(suite.ctx, userAddr, sdk.NewCoins(tc.args.deposit))
			suite.Require().NoError(err)

			// Check that Hard hooks initialized a HardLiquidityProviderClaim with 0 reward indexes
			claim, found := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, userAddr)
			suite.Require().True(found)
			multiRewardIndex, _ := claim.SupplyRewardIndexes.GetRewardIndex(tc.args.deposit.Denom)
			for _, expectedRewardIndex := range tc.args.expectedRewardIndexes {
				currRewardIndex, found := multiRewardIndex.RewardIndexes.GetRewardIndex(expectedRewardIndex.CollateralType)
				suite.Require().True(found)
				suite.Require().Equal(sdk.ZeroDec(), currRewardIndex.RewardFactor)
			}

			// Run accumulator at several intervals
			var timeElapsed int
			previousBlockTime := suite.ctx.BlockTime()
			for _, t := range tc.args.blockTimes {
				timeElapsed += t
				updatedBlockTime := previousBlockTime.Add(time.Duration(int(time.Second) * t))
				previousBlockTime = updatedBlockTime
				blockCtx := suite.ctx.WithBlockTime(updatedBlockTime)

				// Run Hard begin blocker for each block ctx to update denom's interest factor
				hard.BeginBlocker(blockCtx, suite.hardKeeper)

				// Accumulate hard supply-side rewards
				multiRewardPeriod, found := suite.keeper.GetHardSupplyRewardPeriods(blockCtx, tc.args.deposit.Denom)
				if found {
					err := suite.keeper.AccumulateHardSupplyRewards(blockCtx, multiRewardPeriod)
					suite.Require().NoError(err)
				}
			}
			updatedBlockTime := suite.ctx.BlockTime().Add(time.Duration(int(time.Second) * timeElapsed))
			suite.ctx = suite.ctx.WithBlockTime(updatedBlockTime)

			// After we've accumulated, run synchronize
			deposit, found := suite.hardKeeper.GetDeposit(suite.ctx, userAddr)
			suite.Require().True(found)
			suite.Require().NotPanics(func() {
				suite.keeper.SynchronizeHardSupplyReward(suite.ctx, deposit)
			})

			// Check that the global reward index's reward factor and user's claim have been updated as expected
			claim, found = suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, userAddr)
			suite.Require().True(found)
			globalRewardIndexes, foundGlobalRewardIndexes := suite.keeper.GetHardSupplyRewardIndexes(suite.ctx, tc.args.deposit.Denom)
			if len(tc.args.rewardsPerSecond) > 0 {
				suite.Require().True(foundGlobalRewardIndexes)
				for _, expectedRewardIndex := range tc.args.expectedRewardIndexes {
					// Check that global reward index has been updated as expected
					globalRewardIndex, found := globalRewardIndexes.GetRewardIndex(expectedRewardIndex.CollateralType)
					suite.Require().True(found)
					suite.Require().Equal(expectedRewardIndex, globalRewardIndex)

					// Check that the user's claim's reward index matches the corresponding global reward index
					multiRewardIndex, found := claim.SupplyRewardIndexes.GetRewardIndex(tc.args.deposit.Denom)
					suite.Require().True(found)
					rewardIndex, found := multiRewardIndex.RewardIndexes.GetRewardIndex(expectedRewardIndex.CollateralType)
					suite.Require().True(found)
					suite.Require().Equal(expectedRewardIndex, rewardIndex)

					// Check that the user's claim holds the expected amount of reward coins
					suite.Require().Equal(
						tc.args.expectedRewards.AmountOf(expectedRewardIndex.CollateralType),
						claim.Reward.AmountOf(expectedRewardIndex.CollateralType),
					)
				}
			}

			// Only test cases with reward param updates continue past this point
			if !tc.args.updateRewardsViaCommmittee {
				return
			}

			// If are no initial rewards per second, add new rewards through a committee param change
			// 1. Construct incentive's new HardSupplyRewardPeriods param
			currIncentiveHardSupplyRewardPeriods := suite.keeper.GetParams(suite.ctx).HardSupplyRewardPeriods
			multiRewardPeriod, found := currIncentiveHardSupplyRewardPeriods.GetMultiRewardPeriod(tc.args.deposit.Denom)
			if found {
				// Deposit denom's reward period exists, but it doesn't have any rewards per second
				index, found := currIncentiveHardSupplyRewardPeriods.GetMultiRewardPeriodIndex(tc.args.deposit.Denom)
				suite.Require().True(found)
				multiRewardPeriod.RewardsPerSecond = tc.args.updatedRewardsPerSecond
				currIncentiveHardSupplyRewardPeriods[index] = multiRewardPeriod
			} else {
				// Deposit denom's reward period does not exist
				_, found := currIncentiveHardSupplyRewardPeriods.GetMultiRewardPeriodIndex(tc.args.deposit.Denom)
				suite.Require().False(found)
				newMultiRewardPeriod := types.NewMultiRewardPeriod(true, tc.args.deposit.Denom, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), tc.args.updatedRewardsPerSecond)
				currIncentiveHardSupplyRewardPeriods = append(currIncentiveHardSupplyRewardPeriods, newMultiRewardPeriod)
			}

			// 2. Construct the parameter change proposal to update HardSupplyRewardPeriods param
			pubProposal := params.NewParameterChangeProposal(
				"Update hard supply rewards", "Adds a new reward coin to the incentive module's hard supply rewards.",
				[]params.ParamChange{
					{
						Subspace: types.ModuleName,                         // target incentive module
						Key:      string(types.KeyHardSupplyRewardPeriods), // target hard supply rewards key
						Value:    string(suite.app.Codec().MustMarshalJSON(currIncentiveHardSupplyRewardPeriods)),
					},
				},
			)

			// 3. Ensure proposal is properly formed
			err = suite.committeeKeeper.ValidatePubProposal(suite.ctx, pubProposal)
			suite.Require().NoError(err)

			// 4. Committee creates proposal
			committeeMemberOne := suite.addrs[0]
			committeeMemberTwo := suite.addrs[1]
			proposalID, err := suite.committeeKeeper.SubmitProposal(suite.ctx, committeeMemberOne, 1, pubProposal)
			suite.Require().NoError(err)

			// 5. Committee votes and passes proposal
			err = suite.committeeKeeper.AddVote(suite.ctx, proposalID, committeeMemberOne)
			suite.Require().NoError(err)
			err = suite.committeeKeeper.AddVote(suite.ctx, proposalID, committeeMemberTwo)
			suite.Require().NoError(err)

			// 6. Check proposal passed
			proposalPasses, err := suite.committeeKeeper.GetProposalResult(suite.ctx, proposalID)
			suite.Require().NoError(err)
			suite.Require().True(proposalPasses)

			// 7. Run committee module's begin blocker to enact proposal
			suite.NotPanics(func() {
				committee.BeginBlocker(suite.ctx, abci.RequestBeginBlock{}, suite.committeeKeeper)
			})

			// We need to accumulate hard supply-side rewards again
			multiRewardPeriod, found = suite.keeper.GetHardSupplyRewardPeriods(suite.ctx, tc.args.deposit.Denom)
			suite.Require().True(found)

			// But new deposit denoms don't have their PreviousHardSupplyRewardAccrualTime set yet,
			// so we need to call the accumulation method once to set the initial reward accrual time
			if tc.args.deposit.Denom != tc.args.incentiveSupplyRewardDenom {
				err = suite.keeper.AccumulateHardSupplyRewards(suite.ctx, multiRewardPeriod)
				suite.Require().NoError(err)
			}

			// Now we can jump forward in time and accumulate rewards
			updatedBlockTime = previousBlockTime.Add(time.Duration(int(time.Second) * tc.args.updatedTimeDuration))
			suite.ctx = suite.ctx.WithBlockTime(updatedBlockTime)
			err = suite.keeper.AccumulateHardSupplyRewards(suite.ctx, multiRewardPeriod)
			suite.Require().NoError(err)

			// After we've accumulated, run synchronize
			deposit, found = suite.hardKeeper.GetDeposit(suite.ctx, userAddr)
			suite.Require().True(found)
			suite.Require().NotPanics(func() {
				suite.keeper.SynchronizeHardSupplyReward(suite.ctx, deposit)
			})

			// Check that the global reward index's reward factor and user's claim have been updated as expected
			globalRewardIndexes, found = suite.keeper.GetHardSupplyRewardIndexes(suite.ctx, tc.args.deposit.Denom)
			suite.Require().True(found)
			claim, found = suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, userAddr)
			suite.Require().True(found)
			for _, expectedRewardIndex := range tc.args.updatedExpectedRewardIndexes {
				// Check that global reward index has been updated as expected
				globalRewardIndex, found := globalRewardIndexes.GetRewardIndex(expectedRewardIndex.CollateralType)
				suite.Require().True(found)
				suite.Require().Equal(expectedRewardIndex, globalRewardIndex)

				// Check that the user's claim's reward index matches the corresponding global reward index
				multiRewardIndex, found := claim.SupplyRewardIndexes.GetRewardIndex(tc.args.deposit.Denom)
				suite.Require().True(found)
				rewardIndex, found := multiRewardIndex.RewardIndexes.GetRewardIndex(expectedRewardIndex.CollateralType)
				suite.Require().True(found)
				suite.Require().Equal(expectedRewardIndex, rewardIndex)

				// Check that the user's claim holds the expected amount of reward coins
				suite.Require().Equal(
					tc.args.updatedExpectedRewards.AmountOf(expectedRewardIndex.CollateralType),
					claim.Reward.AmountOf(expectedRewardIndex.CollateralType),
				)
			}
		})
	}
}

func (suite *SupplyRewardsTestSuite) TestUpdateHardSupplyIndexDenoms() {
	type depositModification struct {
		coins    sdk.Coins
		withdraw bool
	}

	type args struct {
		firstDeposit              sdk.Coins
		modification              depositModification
		rewardsPerSecond          sdk.Coins
		initialTime               time.Time
		expectedSupplyIndexDenoms []string
	}
	type test struct {
		name string
		args args
	}

	testCases := []test{
		{
			"single reward denom: update adds one supply reward index",
			args{
				firstDeposit:              cs(c("bnb", 10000000000)),
				modification:              depositModification{coins: cs(c("ukava", 10000000000))},
				rewardsPerSecond:          cs(c("hard", 122354)),
				initialTime:               time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				expectedSupplyIndexDenoms: []string{"bnb", "ukava"},
			},
		},
		{
			"single reward denom: update adds multiple supply reward indexes",
			args{
				firstDeposit:              cs(c("bnb", 10000000000)),
				modification:              depositModification{coins: cs(c("ukava", 10000000000), c("btcb", 10000000000), c("xrp", 10000000000))},
				rewardsPerSecond:          cs(c("hard", 122354)),
				initialTime:               time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				expectedSupplyIndexDenoms: []string{"bnb", "ukava", "btcb", "xrp"},
			},
		},
		{
			"single reward denom: update doesn't add duplicate supply reward index for same denom",
			args{
				firstDeposit:              cs(c("bnb", 10000000000)),
				modification:              depositModification{coins: cs(c("bnb", 5000000000))},
				rewardsPerSecond:          cs(c("hard", 122354)),
				initialTime:               time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				expectedSupplyIndexDenoms: []string{"bnb"},
			},
		},
		{
			"multiple reward denoms: update adds one supply reward index",
			args{
				firstDeposit:              cs(c("bnb", 10000000000)),
				modification:              depositModification{coins: cs(c("ukava", 10000000000))},
				rewardsPerSecond:          cs(c("hard", 122354), c("ukava", 122354)),
				initialTime:               time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				expectedSupplyIndexDenoms: []string{"bnb", "ukava"},
			},
		},
		{
			"multiple reward denoms: update adds multiple supply reward indexes",
			args{
				firstDeposit:              cs(c("bnb", 10000000000)),
				modification:              depositModification{coins: cs(c("ukava", 10000000000), c("btcb", 10000000000), c("xrp", 10000000000))},
				rewardsPerSecond:          cs(c("hard", 122354), c("ukava", 122354)),
				initialTime:               time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				expectedSupplyIndexDenoms: []string{"bnb", "ukava", "btcb", "xrp"},
			},
		},
		{
			"multiple reward denoms: update doesn't add duplicate supply reward index for same denom",
			args{
				firstDeposit:              cs(c("bnb", 10000000000)),
				modification:              depositModification{coins: cs(c("bnb", 5000000000))},
				rewardsPerSecond:          cs(c("hard", 122354), c("ukava", 122354)),
				initialTime:               time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				expectedSupplyIndexDenoms: []string{"bnb"},
			},
		},
		{
			"single reward denom: fully withdrawing a denom deletes the denom's supply reward index",
			args{
				firstDeposit:              cs(c("bnb", 1000000000)),
				modification:              depositModification{coins: cs(c("bnb", 1100000000)), withdraw: true},
				rewardsPerSecond:          cs(c("hard", 122354)),
				initialTime:               time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				expectedSupplyIndexDenoms: []string{},
			},
		},
		{
			"single reward denom: fully withdrawing a denom deletes only the denom's supply reward index",
			args{
				firstDeposit:              cs(c("bnb", 1000000000), c("ukava", 100000000)),
				modification:              depositModification{coins: cs(c("bnb", 1100000000)), withdraw: true},
				rewardsPerSecond:          cs(c("hard", 122354)),
				initialTime:               time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				expectedSupplyIndexDenoms: []string{"ukava"},
			},
		},
		{
			"multiple reward denoms: fully repaying a denom deletes the denom's supply reward index",
			args{
				firstDeposit:              cs(c("bnb", 1000000000)),
				modification:              depositModification{coins: cs(c("bnb", 1100000000)), withdraw: true},
				rewardsPerSecond:          cs(c("hard", 122354), c("ukava", 122354)),
				initialTime:               time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				expectedSupplyIndexDenoms: []string{},
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupWithGenState()
			suite.ctx = suite.ctx.WithBlockTime(tc.args.initialTime)

			// Set up generic reward periods
			var multiRewardPeriods types.MultiRewardPeriods
			var rewardPeriods types.RewardPeriods
			for i, denom := range tc.args.expectedSupplyIndexDenoms {
				// Create just one reward period for USDX Minting / Hard Delegator reward periods (otherwise params will panic on duplicate)
				if i == 0 {
					rewardPeriod := types.NewRewardPeriod(true, denom, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), tc.args.rewardsPerSecond[i])
					rewardPeriods = append(rewardPeriods, rewardPeriod)
				}
				multiRewardPeriod := types.NewMultiRewardPeriod(true, denom, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), tc.args.rewardsPerSecond)
				multiRewardPeriods = append(multiRewardPeriods, multiRewardPeriod)
			}

			// Setup incentive state
			params := types.NewParams(
				rewardPeriods, multiRewardPeriods, multiRewardPeriods, rewardPeriods,
				types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
				tc.args.initialTime.Add(time.Hour*24*365*5),
			)
			suite.keeper.SetParams(suite.ctx, params)

			// Set each denom's previous accrual time and supply reward factor
			var rewardIndexes types.RewardIndexes
			for _, rewardCoin := range tc.args.rewardsPerSecond {
				rewardIndex := types.NewRewardIndex(rewardCoin.Denom, sdk.ZeroDec())
				rewardIndexes = append(rewardIndexes, rewardIndex)
			}
			for _, denom := range tc.args.expectedSupplyIndexDenoms {
				suite.keeper.SetPreviousHardSupplyRewardAccrualTime(suite.ctx, denom, tc.args.initialTime)
				suite.keeper.SetHardSupplyRewardIndexes(suite.ctx, denom, rewardIndexes)
			}

			// User deposits (first time)
			userAddr := suite.addrs[3]
			err := suite.hardKeeper.Deposit(suite.ctx, userAddr, tc.args.firstDeposit)
			suite.Require().NoError(err)

			// Confirm that a claim was created and populated with the correct supply indexes
			claimAfterFirstDeposit, found := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, suite.addrs[3])
			suite.Require().True(found)
			for _, coin := range tc.args.firstDeposit {
				_, hasIndex := claimAfterFirstDeposit.HasSupplyRewardIndex(coin.Denom)
				suite.Require().True(hasIndex)
			}
			suite.Require().True(len(claimAfterFirstDeposit.SupplyRewardIndexes) == len(tc.args.firstDeposit))

			// User modifies their Deposit by withdrawing or depositing more
			if tc.args.modification.withdraw {
				err = suite.hardKeeper.Withdraw(suite.ctx, userAddr, tc.args.modification.coins)
			} else {
				err = suite.hardKeeper.Deposit(suite.ctx, userAddr, tc.args.modification.coins)
			}
			suite.Require().NoError(err)

			// Confirm that the claim contains all expected supply indexes
			claimAfterModification, found := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, suite.addrs[3])
			suite.Require().True(found)
			for _, denom := range tc.args.expectedSupplyIndexDenoms {
				_, hasIndex := claimAfterModification.HasSupplyRewardIndex(denom)
				suite.Require().True(hasIndex)
			}
			suite.Require().True(len(claimAfterModification.SupplyRewardIndexes) == len(tc.args.expectedSupplyIndexDenoms))
		})
	}
}

func (suite *SupplyRewardsTestSuite) TestSimulateHardSupplyRewardSynchronization() {
	type args struct {
		deposit               sdk.Coin
		rewardsPerSecond      sdk.Coins
		initialTime           time.Time
		blockTimes            []int
		expectedRewardIndexes types.RewardIndexes
		expectedRewards       sdk.Coins
	}
	type test struct {
		name string
		args args
	}

	testCases := []test{
		{
			"10 blocks",
			args{
				deposit:               c("bnb", 10000000000),
				rewardsPerSecond:      cs(c("hard", 122354)),
				initialTime:           time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				blockTimes:            []int{10, 10, 10, 10, 10, 10, 10, 10, 10, 10},
				expectedRewardIndexes: types.RewardIndexes{types.NewRewardIndex("hard", d("0.001223540000000000"))},
				expectedRewards:       cs(c("hard", 12235400)),
			},
		},
		{
			"10 blocks - long block time",
			args{
				deposit:               c("bnb", 10000000000),
				rewardsPerSecond:      cs(c("hard", 122354)),
				initialTime:           time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				blockTimes:            []int{86400, 86400, 86400, 86400, 86400, 86400, 86400, 86400, 86400, 86400},
				expectedRewardIndexes: types.RewardIndexes{types.NewRewardIndex("hard", d("10.571385600000000000"))},
				expectedRewards:       cs(c("hard", 105713856000)),
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupWithGenState()
			suite.ctx = suite.ctx.WithBlockTime(tc.args.initialTime)

			// Set up incentive state
			params := types.NewParams(
				types.RewardPeriods{types.NewRewardPeriod(true, tc.args.deposit.Denom, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), tc.args.rewardsPerSecond[0])},
				types.MultiRewardPeriods{types.NewMultiRewardPeriod(true, tc.args.deposit.Denom, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), tc.args.rewardsPerSecond)},
				types.MultiRewardPeriods{types.NewMultiRewardPeriod(true, tc.args.deposit.Denom, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), tc.args.rewardsPerSecond)},
				types.RewardPeriods{types.NewRewardPeriod(true, tc.args.deposit.Denom, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), tc.args.rewardsPerSecond[0])},
				types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
				tc.args.initialTime.Add(time.Hour*24*365*5),
			)
			suite.keeper.SetParams(suite.ctx, params)
			suite.keeper.SetPreviousHardSupplyRewardAccrualTime(suite.ctx, tc.args.deposit.Denom, tc.args.initialTime)
			var rewardIndexes types.RewardIndexes
			for _, rewardCoin := range tc.args.rewardsPerSecond {
				rewardIndex := types.NewRewardIndex(rewardCoin.Denom, sdk.ZeroDec())
				rewardIndexes = append(rewardIndexes, rewardIndex)
			}
			suite.keeper.SetHardSupplyRewardIndexes(suite.ctx, tc.args.deposit.Denom, rewardIndexes)

			// Set up hard state (interest factor for the relevant denom)
			suite.hardKeeper.SetSupplyInterestFactor(suite.ctx, tc.args.deposit.Denom, sdk.MustNewDecFromStr("1.0"))
			suite.hardKeeper.SetPreviousAccrualTime(suite.ctx, tc.args.deposit.Denom, tc.args.initialTime)

			// User deposits and borrows to increase total borrowed amount
			userAddr := suite.addrs[3]
			err := suite.hardKeeper.Deposit(suite.ctx, userAddr, sdk.NewCoins(tc.args.deposit))
			suite.Require().NoError(err)

			// Check that Hard hooks initialized a HardLiquidityProviderClaim
			claim, found := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, suite.addrs[3])
			suite.Require().True(found)
			multiRewardIndex, _ := claim.SupplyRewardIndexes.GetRewardIndex(tc.args.deposit.Denom)
			for _, expectedRewardIndex := range tc.args.expectedRewardIndexes {
				currRewardIndex, found := multiRewardIndex.RewardIndexes.GetRewardIndex(expectedRewardIndex.CollateralType)
				suite.Require().True(found)
				suite.Require().Equal(sdk.ZeroDec(), currRewardIndex.RewardFactor)
			}

			// Run accumulator at several intervals
			var timeElapsed int
			previousBlockTime := suite.ctx.BlockTime()
			for _, t := range tc.args.blockTimes {
				timeElapsed += t
				updatedBlockTime := previousBlockTime.Add(time.Duration(int(time.Second) * t))
				previousBlockTime = updatedBlockTime
				blockCtx := suite.ctx.WithBlockTime(updatedBlockTime)

				// Run Hard begin blocker for each block ctx to update denom's interest factor
				hard.BeginBlocker(blockCtx, suite.hardKeeper)

				// Accumulate hard supply-side rewards
				multiRewardPeriod, found := suite.keeper.GetHardSupplyRewardPeriods(blockCtx, tc.args.deposit.Denom)
				suite.Require().True(found)
				err := suite.keeper.AccumulateHardSupplyRewards(blockCtx, multiRewardPeriod)
				suite.Require().NoError(err)
			}
			updatedBlockTime := suite.ctx.BlockTime().Add(time.Duration(int(time.Second) * timeElapsed))
			suite.ctx = suite.ctx.WithBlockTime(updatedBlockTime)

			// Confirm that the user's claim hasn't been synced
			claimPre, foundPre := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, suite.addrs[3])
			suite.Require().True(foundPre)
			multiRewardIndexPre, _ := claimPre.SupplyRewardIndexes.GetRewardIndex(tc.args.deposit.Denom)
			for _, expectedRewardIndex := range tc.args.expectedRewardIndexes {
				currRewardIndex, found := multiRewardIndexPre.RewardIndexes.GetRewardIndex(expectedRewardIndex.CollateralType)
				suite.Require().True(found)
				suite.Require().Equal(sdk.ZeroDec(), currRewardIndex.RewardFactor)
			}

			// Check that the synced claim held in memory has properly simulated syncing
			syncedClaim := suite.keeper.SimulateHardSynchronization(suite.ctx, claimPre)
			for _, expectedRewardIndex := range tc.args.expectedRewardIndexes {
				// Check that the user's claim's reward index matches the expected reward index
				multiRewardIndex, found := syncedClaim.SupplyRewardIndexes.GetRewardIndex(tc.args.deposit.Denom)
				suite.Require().True(found)
				rewardIndex, found := multiRewardIndex.RewardIndexes.GetRewardIndex(expectedRewardIndex.CollateralType)
				suite.Require().True(found)
				suite.Require().Equal(expectedRewardIndex, rewardIndex)

				// Check that the user's claim holds the expected amount of reward coins
				suite.Require().Equal(
					tc.args.expectedRewards.AmountOf(expectedRewardIndex.CollateralType),
					syncedClaim.Reward.AmountOf(expectedRewardIndex.CollateralType),
				)
			}
		})
	}
}

func TestSupplyRewardsTestSuite(t *testing.T) {
	suite.Run(t, new(SupplyRewardsTestSuite))
}
