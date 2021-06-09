package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/kava/x/committee"
	"github.com/kava-labs/kava/x/hard"
	hardtypes "github.com/kava-labs/kava/x/hard/types"
	"github.com/kava-labs/kava/x/incentive/types"
)

func (suite *KeeperTestSuite) TestAccumulateHardBorrowRewards() {
	type args struct {
		borrow                sdk.Coin
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
				borrow:                c("bnb", 1000000000000),
				rewardsPerSecond:      cs(c("hard", 122354)),
				initialTime:           time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				timeElapsed:           7,
				expectedRewardIndexes: types.RewardIndexes{types.NewRewardIndex("hard", d("0.000000856478000001"))},
			},
		},
		{
			"single reward denom: 1 day",
			args{
				borrow:                c("bnb", 1000000000000),
				rewardsPerSecond:      cs(c("hard", 122354)),
				initialTime:           time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				timeElapsed:           86400,
				expectedRewardIndexes: types.RewardIndexes{types.NewRewardIndex("hard", d("0.010571385600010177"))},
			},
		},
		{
			"single reward denom: 0 seconds",
			args{
				borrow:                c("bnb", 1000000000000),
				rewardsPerSecond:      cs(c("hard", 122354)),
				initialTime:           time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				timeElapsed:           0,
				expectedRewardIndexes: types.RewardIndexes{types.NewRewardIndex("hard", d("0.0"))},
			},
		},
		{
			"multiple reward denoms: 7 seconds",
			args{
				borrow:           c("bnb", 1000000000000),
				rewardsPerSecond: cs(c("hard", 122354), c("ukava", 122354)),
				initialTime:      time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				timeElapsed:      7,
				expectedRewardIndexes: types.RewardIndexes{
					types.NewRewardIndex("hard", d("0.000000856478000001")),
					types.NewRewardIndex("ukava", d("0.000000856478000001")),
				},
			},
		},
		{
			"multiple reward denoms: 1 day",
			args{
				borrow:           c("bnb", 1000000000000),
				rewardsPerSecond: cs(c("hard", 122354), c("ukava", 122354)),
				initialTime:      time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				timeElapsed:      86400,
				expectedRewardIndexes: types.RewardIndexes{
					types.NewRewardIndex("hard", d("0.010571385600010177")),
					types.NewRewardIndex("ukava", d("0.010571385600010177")),
				},
			},
		},
		{
			"multiple reward denoms: 0 seconds",
			args{
				borrow:           c("bnb", 1000000000000),
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
				borrow:           c("bnb", 1000000000000),
				rewardsPerSecond: cs(c("hard", 122354), c("ukava", 555555)),
				initialTime:      time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				timeElapsed:      86400,
				expectedRewardIndexes: types.RewardIndexes{
					types.NewRewardIndex("hard", d("0.010571385600010177")),
					types.NewRewardIndex("ukava", d("0.047999952000046210")),
				},
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupWithGenState()
			suite.ctx = suite.ctx.WithBlockTime(tc.args.initialTime)

			// Mint coins to hard module account
			supplyKeeper := suite.app.GetSupplyKeeper()
			hardMaccCoins := sdk.NewCoins(sdk.NewCoin("usdx", sdk.NewInt(200000000)))
			supplyKeeper.MintCoins(suite.ctx, hardtypes.ModuleAccountName, hardMaccCoins)

			// setup incentive state
			params := types.NewParams(
				types.RewardPeriods{types.NewRewardPeriod(true, tc.args.borrow.Denom, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), tc.args.rewardsPerSecond[0])},
				types.MultiRewardPeriods{types.NewMultiRewardPeriod(true, tc.args.borrow.Denom, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), tc.args.rewardsPerSecond)},
				types.MultiRewardPeriods{types.NewMultiRewardPeriod(true, tc.args.borrow.Denom, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), tc.args.rewardsPerSecond)},
				types.RewardPeriods{types.NewRewardPeriod(true, tc.args.borrow.Denom, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), tc.args.rewardsPerSecond[0])},
				types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
				tc.args.initialTime.Add(time.Hour*24*365*5),
			)
			suite.keeper.SetParams(suite.ctx, params)
			suite.keeper.SetPreviousHardBorrowRewardAccrualTime(suite.ctx, tc.args.borrow.Denom, tc.args.initialTime)
			var rewardIndexes types.RewardIndexes
			for _, rewardCoin := range tc.args.rewardsPerSecond {
				rewardIndex := types.NewRewardIndex(rewardCoin.Denom, sdk.ZeroDec())
				rewardIndexes = append(rewardIndexes, rewardIndex)
			}
			suite.keeper.SetHardBorrowRewardIndexes(suite.ctx, tc.args.borrow.Denom, rewardIndexes)

			// Set up hard state (interest factor for the relevant denom)
			suite.hardKeeper.SetSupplyInterestFactor(suite.ctx, tc.args.borrow.Denom, sdk.MustNewDecFromStr("1.0"))
			suite.hardKeeper.SetBorrowInterestFactor(suite.ctx, tc.args.borrow.Denom, sdk.MustNewDecFromStr("1.0"))
			suite.hardKeeper.SetPreviousAccrualTime(suite.ctx, tc.args.borrow.Denom, tc.args.initialTime)

			// User deposits and borrows to increase total borrowed amount
			hardKeeper := suite.app.GetHardKeeper()
			userAddr := suite.addrs[3]
			err := hardKeeper.Deposit(suite.ctx, userAddr, sdk.NewCoins(sdk.NewCoin(tc.args.borrow.Denom, tc.args.borrow.Amount.Mul(sdk.NewInt(2)))))
			suite.Require().NoError(err)
			err = hardKeeper.Borrow(suite.ctx, userAddr, sdk.NewCoins(tc.args.borrow))
			suite.Require().NoError(err)

			// Set up chain context at future time
			runAtTime := suite.ctx.BlockTime().Add(time.Duration(int(time.Second) * tc.args.timeElapsed))
			runCtx := suite.ctx.WithBlockTime(runAtTime)

			// Run Hard begin blocker in order to update the denom's index factor
			hard.BeginBlocker(runCtx, suite.hardKeeper)

			// Accumulate hard borrow rewards for the deposit denom
			multiRewardPeriod, found := suite.keeper.GetHardBorrowRewardPeriods(runCtx, tc.args.borrow.Denom)
			suite.Require().True(found)
			err = suite.keeper.AccumulateHardBorrowRewards(runCtx, multiRewardPeriod)
			suite.Require().NoError(err)

			// Check that each expected reward index matches the current stored reward index for the denom
			globalRewardIndexes, found := suite.keeper.GetHardBorrowRewardIndexes(runCtx, tc.args.borrow.Denom)
			suite.Require().True(found)
			for _, expectedRewardIndex := range tc.args.expectedRewardIndexes {
				globalRewardIndex, found := globalRewardIndexes.GetRewardIndex(expectedRewardIndex.CollateralType)
				suite.Require().True(found)
				suite.Require().Equal(expectedRewardIndex, globalRewardIndex)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestInitializeHardBorrowRewards() {

	type args struct {
		moneyMarketRewardDenoms          map[string][]string
		deposit                          sdk.Coins
		borrow                           sdk.Coins
		initialTime                      time.Time
		expectedClaimBorrowRewardIndexes types.MultiRewardIndexes
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
				borrow:                  cs(c("bnb", 100000000000)),
				initialTime:             time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				expectedClaimBorrowRewardIndexes: types.MultiRewardIndexes{
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
				borrow:                  cs(c("btcb", 100000000000)),
				initialTime:             time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				expectedClaimBorrowRewardIndexes: types.MultiRewardIndexes{
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
				borrow:                  cs(c("xrp", 100000000000)),
				initialTime:             time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				expectedClaimBorrowRewardIndexes: types.MultiRewardIndexes{
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
				borrow:                  cs(c("bnb", 100000000000), c("btcb", 100000000000)),
				initialTime:             time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				expectedClaimBorrowRewardIndexes: types.MultiRewardIndexes{
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
				borrow:                  cs(c("bnb", 100000000000), c("xrp", 100000000000)),
				initialTime:             time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				expectedClaimBorrowRewardIndexes: types.MultiRewardIndexes{
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

			// Mint coins to hard module account
			supplyKeeper := suite.app.GetSupplyKeeper()
			hardMaccCoins := sdk.NewCoins(sdk.NewCoin("usdx", sdk.NewInt(200000000)))
			supplyKeeper.MintCoins(suite.ctx, hardtypes.ModuleAccountName, hardMaccCoins)

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
				suite.keeper.SetPreviousHardBorrowRewardAccrualTime(suite.ctx, moneyMarketDenom, tc.args.initialTime)
				if len(rewardIndexes) > 0 {
					suite.keeper.SetHardBorrowRewardIndexes(suite.ctx, moneyMarketDenom, rewardIndexes)
				}
			}

			hardKeeper := suite.app.GetHardKeeper()
			// User deposits
			err := hardKeeper.Deposit(suite.ctx, userAddr, tc.args.deposit)
			suite.Require().NoError(err)
			// User borrows
			err = hardKeeper.Borrow(suite.ctx, userAddr, tc.args.borrow)
			suite.Require().NoError(err)

			claim, foundClaim := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, userAddr)
			suite.Require().True(foundClaim)
			suite.Require().Equal(tc.args.expectedClaimBorrowRewardIndexes, claim.BorrowRewardIndexes)
		})
	}
}

func (suite *KeeperTestSuite) TestSynchronizeHardBorrowReward() {
	type args struct {
		incentiveBorrowRewardDenom   string
		borrow                       sdk.Coin
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
				incentiveBorrowRewardDenom: "bnb",
				borrow:                     c("bnb", 10000000000),
				rewardsPerSecond:           cs(c("hard", 122354)),
				initialTime:                time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				blockTimes:                 []int{10, 10, 10, 10, 10, 10, 10, 10, 10, 10},
				expectedRewardIndexes:      types.RewardIndexes{types.NewRewardIndex("hard", d("0.001223540000173228"))},
				expectedRewards:            cs(c("hard", 12235400)),
				updateRewardsViaCommmittee: false,
			},
		},
		{
			"single reward denom: 10 blocks - long block time",
			args{
				incentiveBorrowRewardDenom: "bnb",
				borrow:                     c("bnb", 10000000000),
				rewardsPerSecond:           cs(c("hard", 122354)),
				initialTime:                time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				blockTimes:                 []int{86400, 86400, 86400, 86400, 86400, 86400, 86400, 86400, 86400, 86400},
				expectedRewardIndexes:      types.RewardIndexes{types.NewRewardIndex("hard", d("10.571385603126235340"))},
				expectedRewards:            cs(c("hard", 105713856031)),
			},
		},
		{
			"single reward denom: user reward index updated when reward is zero",
			args{
				incentiveBorrowRewardDenom: "ukava",
				borrow:                     c("ukava", 1), // borrow a tiny amount so that rewards round to zero
				rewardsPerSecond:           cs(c("hard", 122354)),
				initialTime:                time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				blockTimes:                 []int{10, 10, 10, 10, 10, 10, 10, 10, 10, 10},
				expectedRewardIndexes:      types.RewardIndexes{types.NewRewardIndex("hard", d("0.122354003908172328"))},
				expectedRewards:            cs(),
				updateRewardsViaCommmittee: false,
			},
		},
		{
			"multiple reward denoms: 10 blocks",
			args{
				incentiveBorrowRewardDenom: "bnb",
				borrow:                     c("bnb", 10000000000),
				rewardsPerSecond:           cs(c("hard", 122354), c("ukava", 122354)),
				initialTime:                time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				blockTimes:                 []int{10, 10, 10, 10, 10, 10, 10, 10, 10, 10},
				expectedRewardIndexes: types.RewardIndexes{
					types.NewRewardIndex("hard", d("0.001223540000173228")),
					types.NewRewardIndex("ukava", d("0.001223540000173228")),
				},
				expectedRewards: cs(c("hard", 12235400), c("ukava", 12235400)),
			},
		},
		{
			"multiple reward denoms: 10 blocks - long block time",
			args{
				incentiveBorrowRewardDenom: "bnb",
				borrow:                     c("bnb", 10000000000),
				rewardsPerSecond:           cs(c("hard", 122354), c("ukava", 122354)),
				initialTime:                time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				blockTimes:                 []int{86400, 86400, 86400, 86400, 86400, 86400, 86400, 86400, 86400, 86400},
				expectedRewardIndexes: types.RewardIndexes{
					types.NewRewardIndex("hard", d("10.571385603126235340")),
					types.NewRewardIndex("ukava", d("10.571385603126235340")),
				},
				expectedRewards: cs(c("hard", 105713856031), c("ukava", 105713856031)),
			},
		},
		{
			"multiple reward denoms with different rewards per second: 10 blocks",
			args{
				incentiveBorrowRewardDenom: "bnb",
				borrow:                     c("bnb", 10000000000),
				rewardsPerSecond:           cs(c("hard", 122354), c("ukava", 555555)),
				initialTime:                time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				blockTimes:                 []int{10, 10, 10, 10, 10, 10, 10, 10, 10, 10},
				expectedRewardIndexes: types.RewardIndexes{
					types.NewRewardIndex("hard", d("0.001223540000173228")),
					types.NewRewardIndex("ukava", d("0.005555550000786558")),
				},
				expectedRewards: cs(c("hard", 12235400), c("ukava", 55555500)),
			},
		},
		{
			"denom is in incentive's hard borrow reward params but it has no rewards; add reward",
			args{
				incentiveBorrowRewardDenom: "bnb",
				borrow:                     c("bnb", 10000000000),
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
					types.NewRewardIndex("hard", d("0.864000000049803065")),
				},
				updatedTimeDuration: 86400,
			},
		},
		{
			"denom is in incentive's hard borrow reward params and has rewards; add new reward type",
			args{
				incentiveBorrowRewardDenom: "bnb",
				borrow:                     c("bnb", 10000000000),
				rewardsPerSecond:           cs(c("hard", 122354)),
				initialTime:                time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				blockTimes:                 []int{86400},
				expectedRewardIndexes: types.RewardIndexes{
					types.NewRewardIndex("hard", d("1.057138560060101160")),
				},
				expectedRewards:            cs(c("hard", 10571385601)),
				updateRewardsViaCommmittee: true,
				updatedBaseDenom:           "bnb",
				updatedRewardsPerSecond:    cs(c("hard", 122354), c("ukava", 100000)),
				updatedExpectedRewards:     cs(c("hard", 21142771202), c("ukava", 8640000000)),
				updatedExpectedRewardIndexes: types.RewardIndexes{
					types.NewRewardIndex("hard", d("2.114277120120202320")),
					types.NewRewardIndex("ukava", d("0.864000000049120715")),
				},
				updatedTimeDuration: 86400,
			},
		},
		{
			"denom is in hard's money market params but not in incentive's hard supply reward params; add reward",
			args{
				incentiveBorrowRewardDenom: "bnb",
				borrow:                     c("zzz", 10000000000),
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
					types.NewRewardIndex("hard", d("0.864000000049803065")),
				},
				updatedTimeDuration: 86400,
			},
		},
		{
			"denom incentive's hard borrow reward params but it has no rewards; add multiple reward types",
			args{
				incentiveBorrowRewardDenom: "bnb",
				borrow:                     c("bnb", 10000000000),
				rewardsPerSecond:           sdk.Coins{},
				initialTime:                time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				blockTimes:                 []int{100},
				expectedRewardIndexes:      types.RewardIndexes{},
				expectedRewards:            sdk.Coins{},
				updateRewardsViaCommmittee: true,
				updatedBaseDenom:           "bnb",
				updatedRewardsPerSecond:    cs(c("hard", 100000), c("ukava", 100500), c("swap", 500)),
				updatedExpectedRewards:     cs(c("hard", 8640000000), c("ukava", 8683200001), c("swap", 43200000)),
				updatedExpectedRewardIndexes: types.RewardIndexes{
					types.NewRewardIndex("hard", d("0.864000000049803065")),
					types.NewRewardIndex("ukava", d("0.868320000050052081")),
					types.NewRewardIndex("swap", d("0.004320000000249015")),
				},
				updatedTimeDuration: 86400,
			},
		},
		{
			"denom is in hard's money market params but not in incentive's hard supply reward params; add multiple reward types",
			args{
				incentiveBorrowRewardDenom: "bnb",
				borrow:                     c("zzz", 10000000000),
				rewardsPerSecond:           sdk.Coins{},
				initialTime:                time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				blockTimes:                 []int{100},
				expectedRewardIndexes:      types.RewardIndexes{},
				expectedRewards:            sdk.Coins{},
				updateRewardsViaCommmittee: true,
				updatedBaseDenom:           "zzz",
				updatedRewardsPerSecond:    cs(c("hard", 100000), c("ukava", 100500), c("swap", 500)),
				updatedExpectedRewards:     cs(c("hard", 8640000000), c("ukava", 8683200001), c("swap", 43200000)),
				updatedExpectedRewardIndexes: types.RewardIndexes{
					types.NewRewardIndex("hard", d("0.864000000049803065")),
					types.NewRewardIndex("ukava", d("0.868320000050052081")),
					types.NewRewardIndex("swap", d("0.004320000000249015")),
				},
				updatedTimeDuration: 86400,
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupWithGenState()
			suite.ctx = suite.ctx.WithBlockTime(tc.args.initialTime)

			// Mint coins to hard module account
			supplyKeeper := suite.app.GetSupplyKeeper()
			hardMaccCoins := sdk.NewCoins(sdk.NewCoin("usdx", sdk.NewInt(200000000)))
			supplyKeeper.MintCoins(suite.ctx, hardtypes.ModuleAccountName, hardMaccCoins)

			// Set up incentive state
			incentiveParams := types.NewParams(
				types.RewardPeriods{types.NewRewardPeriod(true, tc.args.incentiveBorrowRewardDenom, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), c("hard", 1))},
				types.MultiRewardPeriods{types.NewMultiRewardPeriod(true, tc.args.incentiveBorrowRewardDenom, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), sdk.Coins{})}, // Don't set any supply rewards for easier accounting
				types.MultiRewardPeriods{types.NewMultiRewardPeriod(true, tc.args.incentiveBorrowRewardDenom, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), tc.args.rewardsPerSecond)},
				types.RewardPeriods{types.NewRewardPeriod(true, tc.args.incentiveBorrowRewardDenom, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), c("hard", 1))},
				types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
				tc.args.initialTime.Add(time.Hour*24*365*5),
			)
			suite.keeper.SetParams(suite.ctx, incentiveParams)
			suite.keeper.SetPreviousHardBorrowRewardAccrualTime(suite.ctx, tc.args.incentiveBorrowRewardDenom, tc.args.initialTime)
			var rewardIndexes types.RewardIndexes
			for _, rewardCoin := range tc.args.rewardsPerSecond {
				rewardIndex := types.NewRewardIndex(rewardCoin.Denom, sdk.ZeroDec())
				rewardIndexes = append(rewardIndexes, rewardIndex)
			}
			if len(rewardIndexes) > 0 {
				suite.keeper.SetHardBorrowRewardIndexes(suite.ctx, tc.args.incentiveBorrowRewardDenom, rewardIndexes)
			}

			// Set up hard state (interest factor for the relevant denom)
			suite.hardKeeper.SetSupplyInterestFactor(suite.ctx, tc.args.borrow.Denom, sdk.MustNewDecFromStr("1.0"))
			suite.hardKeeper.SetBorrowInterestFactor(suite.ctx, tc.args.borrow.Denom, sdk.MustNewDecFromStr("1.0"))
			suite.hardKeeper.SetPreviousAccrualTime(suite.ctx, tc.args.borrow.Denom, tc.args.initialTime)
			// Set the minimum borrow to 0 to allow testing small borrows
			hardParams := suite.hardKeeper.GetParams(suite.ctx)
			hardParams.MinimumBorrowUSDValue = sdk.ZeroDec()
			suite.hardKeeper.SetParams(suite.ctx, hardParams)

			// Borrow a fixed amount from another user to dilute primary user's rewards per second.
			suite.Require().NoError(
				suite.hardKeeper.Deposit(suite.ctx, suite.addrs[2], cs(c("ukava", 200_000_000))),
			)
			suite.Require().NoError(
				suite.hardKeeper.Borrow(suite.ctx, suite.addrs[2], cs(c("ukava", 100_000_000))),
			)

			// User deposits and borrows to increase total borrowed amount
			hardKeeper := suite.app.GetHardKeeper()
			userAddr := suite.addrs[3]
			err := hardKeeper.Deposit(suite.ctx, userAddr, sdk.NewCoins(sdk.NewCoin(tc.args.borrow.Denom, tc.args.borrow.Amount.Mul(sdk.NewInt(2)))))
			suite.Require().NoError(err)
			err = hardKeeper.Borrow(suite.ctx, userAddr, sdk.NewCoins(tc.args.borrow))
			suite.Require().NoError(err)

			// Check that Hard hooks initialized a HardLiquidityProviderClaim
			claim, found := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, userAddr)
			suite.Require().True(found)
			multiRewardIndex, _ := claim.BorrowRewardIndexes.GetRewardIndex(tc.args.borrow.Denom)
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

				// Accumulate hard borrow-side rewards
				multiRewardPeriod, found := suite.keeper.GetHardBorrowRewardPeriods(blockCtx, tc.args.borrow.Denom)
				if found {
					err := suite.keeper.AccumulateHardBorrowRewards(blockCtx, multiRewardPeriod)
					suite.Require().NoError(err)
				}
			}
			updatedBlockTime := suite.ctx.BlockTime().Add(time.Duration(int(time.Second) * timeElapsed))
			suite.ctx = suite.ctx.WithBlockTime(updatedBlockTime)

			// After we've accumulated, run synchronize
			borrow, found := hardKeeper.GetBorrow(suite.ctx, userAddr)
			suite.Require().True(found)
			suite.Require().NotPanics(func() {
				suite.keeper.SynchronizeHardBorrowReward(suite.ctx, borrow)
			})

			// Check that the global reward index's reward factor and user's claim have been updated as expected
			claim, found = suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, userAddr)
			suite.Require().True(found)
			globalRewardIndexes, foundGlobalRewardIndexes := suite.keeper.GetHardBorrowRewardIndexes(suite.ctx, tc.args.borrow.Denom)
			if len(tc.args.rewardsPerSecond) > 0 {
				suite.Require().True(foundGlobalRewardIndexes)
				for _, expectedRewardIndex := range tc.args.expectedRewardIndexes {
					// Check that global reward index has been updated as expected
					globalRewardIndex, found := globalRewardIndexes.GetRewardIndex(expectedRewardIndex.CollateralType)
					suite.Require().True(found)
					suite.Require().Equal(expectedRewardIndex, globalRewardIndex)

					// Check that the user's claim's reward index matches the corresponding global reward index
					multiRewardIndex, found := claim.BorrowRewardIndexes.GetRewardIndex(tc.args.borrow.Denom)
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
			// 1. Construct incentive's new HardBorrowRewardPeriods param
			currIncentiveHardBorrowRewardPeriods := suite.keeper.GetParams(suite.ctx).HardBorrowRewardPeriods
			multiRewardPeriod, found := currIncentiveHardBorrowRewardPeriods.GetMultiRewardPeriod(tc.args.borrow.Denom)
			if found {
				// Borrow denom's reward period exists, but it doesn't have any rewards per second
				index, found := currIncentiveHardBorrowRewardPeriods.GetMultiRewardPeriodIndex(tc.args.borrow.Denom)
				suite.Require().True(found)
				multiRewardPeriod.RewardsPerSecond = tc.args.updatedRewardsPerSecond
				currIncentiveHardBorrowRewardPeriods[index] = multiRewardPeriod
			} else {
				// Borrow denom's reward period does not exist
				_, found := currIncentiveHardBorrowRewardPeriods.GetMultiRewardPeriodIndex(tc.args.borrow.Denom)
				suite.Require().False(found)
				newMultiRewardPeriod := types.NewMultiRewardPeriod(true, tc.args.borrow.Denom, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), tc.args.updatedRewardsPerSecond)
				currIncentiveHardBorrowRewardPeriods = append(currIncentiveHardBorrowRewardPeriods, newMultiRewardPeriod)
			}

			// 2. Construct the parameter change proposal to update HardBorrowRewardPeriods param
			pubProposal := params.NewParameterChangeProposal(
				"Update hard borrow rewards", "Adds a new reward coin to the incentive module's hard borrow rewards.",
				[]params.ParamChange{
					{
						Subspace: types.ModuleName,                         // target incentive module
						Key:      string(types.KeyHardBorrowRewardPeriods), // target hard borrow rewards key
						Value:    string(suite.app.Codec().MustMarshalJSON(currIncentiveHardBorrowRewardPeriods)),
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
			err = suite.committeeKeeper.AddVote(suite.ctx, proposalID, committeeMemberOne, committee.Yes)
			err = suite.committeeKeeper.AddVote(suite.ctx, proposalID, committeeMemberTwo, committee.Yes)

			// 6. Check proposal passed
			com, found := suite.committeeKeeper.GetCommittee(suite.ctx, 1)
			suite.Require().True(found)
			proposalPasses := suite.committeeKeeper.GetProposalResult(suite.ctx, proposalID, com)
			suite.Require().NoError(err)
			suite.Require().True(proposalPasses)

			// 7. Run committee module's begin blocker to enact proposal
			suite.NotPanics(func() {
				committee.BeginBlocker(suite.ctx, abci.RequestBeginBlock{}, suite.committeeKeeper)
			})

			// We need to accumulate hard supply-side rewards again
			multiRewardPeriod, found = suite.keeper.GetHardBorrowRewardPeriods(suite.ctx, tc.args.borrow.Denom)
			suite.Require().True(found)

			// But new borrow denoms don't have their PreviousHardBorrowRewardAccrualTime set yet,
			// so we need to call the accumulation method once to set the initial reward accrual time
			if tc.args.borrow.Denom != tc.args.incentiveBorrowRewardDenom {
				err = suite.keeper.AccumulateHardBorrowRewards(suite.ctx, multiRewardPeriod)
				suite.Require().NoError(err)
			}

			// Now we can jump forward in time and accumulate rewards
			updatedBlockTime = previousBlockTime.Add(time.Duration(int(time.Second) * tc.args.updatedTimeDuration))
			suite.ctx = suite.ctx.WithBlockTime(updatedBlockTime)
			err = suite.keeper.AccumulateHardBorrowRewards(suite.ctx, multiRewardPeriod)
			suite.Require().NoError(err)

			// After we've accumulated, run synchronize
			borrow, found = hardKeeper.GetBorrow(suite.ctx, userAddr)
			suite.Require().True(found)
			suite.Require().NotPanics(func() {
				suite.keeper.SynchronizeHardBorrowReward(suite.ctx, borrow)
			})

			// Check that the global reward index's reward factor and user's claim have been updated as expected
			globalRewardIndexes, found = suite.keeper.GetHardBorrowRewardIndexes(suite.ctx, tc.args.borrow.Denom)
			suite.Require().True(found)
			claim, found = suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, userAddr)
			suite.Require().True(found)

			for _, expectedRewardIndex := range tc.args.updatedExpectedRewardIndexes {
				// Check that global reward index has been updated as expected
				globalRewardIndex, found := globalRewardIndexes.GetRewardIndex(expectedRewardIndex.CollateralType)
				suite.Require().True(found)
				suite.Require().Equal(expectedRewardIndex, globalRewardIndex)
				// Check that the user's claim's reward index matches the corresponding global reward index
				multiRewardIndex, found := claim.BorrowRewardIndexes.GetRewardIndex(tc.args.borrow.Denom)
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

func (suite *KeeperTestSuite) TestUpdateHardBorrowIndexDenoms() {
	type withdrawModification struct {
		coins sdk.Coins
		repay bool
	}

	type args struct {
		initialDeposit            sdk.Coins
		firstBorrow               sdk.Coins
		modification              withdrawModification
		rewardsPerSecond          sdk.Coins
		initialTime               time.Time
		expectedBorrowIndexDenoms []string
	}
	type test struct {
		name string
		args args
	}

	testCases := []test{
		{
			"single reward denom: update adds one borrow reward index",
			args{
				initialDeposit:            cs(c("bnb", 10000000000)),
				firstBorrow:               cs(c("bnb", 50000000)),
				modification:              withdrawModification{coins: cs(c("ukava", 500000000))},
				rewardsPerSecond:          cs(c("hard", 122354)),
				initialTime:               time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				expectedBorrowIndexDenoms: []string{"bnb", "ukava"},
			},
		},
		{
			"single reward denom: update adds multiple borrow supply reward indexes",
			args{
				initialDeposit:            cs(c("btcb", 10000000000)),
				firstBorrow:               cs(c("btcb", 50000000)),
				modification:              withdrawModification{coins: cs(c("ukava", 500000000), c("bnb", 50000000000), c("xrp", 50000000000))},
				rewardsPerSecond:          cs(c("hard", 122354)),
				initialTime:               time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				expectedBorrowIndexDenoms: []string{"btcb", "ukava", "bnb", "xrp"},
			},
		},
		{
			"single reward denom: update doesn't add duplicate borrow reward index for same denom",
			args{
				initialDeposit:            cs(c("bnb", 100000000000)),
				firstBorrow:               cs(c("bnb", 50000000)),
				modification:              withdrawModification{coins: cs(c("bnb", 50000000000))},
				rewardsPerSecond:          cs(c("hard", 122354)),
				initialTime:               time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				expectedBorrowIndexDenoms: []string{"bnb"},
			},
		},
		{
			"multiple reward denoms: update adds one borrow reward index",
			args{
				initialDeposit:            cs(c("bnb", 10000000000)),
				firstBorrow:               cs(c("bnb", 50000000)),
				modification:              withdrawModification{coins: cs(c("ukava", 500000000))},
				rewardsPerSecond:          cs(c("hard", 122354), c("ukava", 122354)),
				initialTime:               time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				expectedBorrowIndexDenoms: []string{"bnb", "ukava"},
			},
		},
		{
			"multiple reward denoms: update adds multiple borrow supply reward indexes",
			args{
				initialDeposit:            cs(c("btcb", 10000000000)),
				firstBorrow:               cs(c("btcb", 50000000)),
				modification:              withdrawModification{coins: cs(c("ukava", 500000000), c("bnb", 50000000000), c("xrp", 50000000000))},
				rewardsPerSecond:          cs(c("hard", 122354), c("ukava", 122354)),
				initialTime:               time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				expectedBorrowIndexDenoms: []string{"btcb", "ukava", "bnb", "xrp"},
			},
		},
		{
			"multiple reward denoms: update doesn't add duplicate borrow reward index for same denom",
			args{
				initialDeposit:            cs(c("bnb", 100000000000)),
				firstBorrow:               cs(c("bnb", 50000000)),
				modification:              withdrawModification{coins: cs(c("bnb", 50000000000))},
				rewardsPerSecond:          cs(c("hard", 122354), c("ukava", 122354)),
				initialTime:               time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				expectedBorrowIndexDenoms: []string{"bnb"},
			},
		},
		{
			"single reward denom: fully repaying a denom deletes the denom's supply reward index",
			args{
				initialDeposit:            cs(c("bnb", 1000000000)),
				firstBorrow:               cs(c("bnb", 100000000)),
				modification:              withdrawModification{coins: cs(c("bnb", 1100000000)), repay: true},
				rewardsPerSecond:          cs(c("hard", 122354)),
				initialTime:               time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				expectedBorrowIndexDenoms: []string{},
			},
		},
		{
			"single reward denom: fully repaying a denom deletes only the denom's supply reward index",
			args{
				initialDeposit:            cs(c("bnb", 1000000000)),
				firstBorrow:               cs(c("bnb", 100000000), c("ukava", 10000000)),
				modification:              withdrawModification{coins: cs(c("bnb", 1100000000)), repay: true},
				rewardsPerSecond:          cs(c("hard", 122354)),
				initialTime:               time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				expectedBorrowIndexDenoms: []string{"ukava"},
			},
		},
		{
			"multiple reward denoms: fully repaying a denom deletes the denom's supply reward index",
			args{
				initialDeposit:            cs(c("bnb", 1000000000)),
				firstBorrow:               cs(c("bnb", 100000000), c("ukava", 10000000)),
				modification:              withdrawModification{coins: cs(c("bnb", 1100000000)), repay: true},
				rewardsPerSecond:          cs(c("hard", 122354), c("ukava", 122354)),
				initialTime:               time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				expectedBorrowIndexDenoms: []string{"ukava"},
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupWithGenState()
			suite.ctx = suite.ctx.WithBlockTime(tc.args.initialTime)

			// Mint coins to hard module account so it can service borrow requests
			supplyKeeper := suite.app.GetSupplyKeeper()
			hardMaccCoins := tc.args.firstBorrow.Add(tc.args.modification.coins...)
			supplyKeeper.MintCoins(suite.ctx, hardtypes.ModuleAccountName, hardMaccCoins)

			// Set up generic reward periods
			var multiRewardPeriods types.MultiRewardPeriods
			var rewardPeriods types.RewardPeriods
			for i, denom := range tc.args.expectedBorrowIndexDenoms {
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

			// Set each expected borrow denom's previous accrual time and borrow reward factor
			var rewardIndexes types.RewardIndexes
			for _, rewardCoin := range tc.args.rewardsPerSecond {
				rewardIndex := types.NewRewardIndex(rewardCoin.Denom, sdk.ZeroDec())
				rewardIndexes = append(rewardIndexes, rewardIndex)
			}
			for _, denom := range tc.args.expectedBorrowIndexDenoms {
				suite.keeper.SetPreviousHardSupplyRewardAccrualTime(suite.ctx, denom, tc.args.initialTime)
				suite.keeper.SetHardBorrowRewardIndexes(suite.ctx, denom, rewardIndexes)
			}

			// User deposits initial funds (so that user can borrow)
			hardKeeper := suite.app.GetHardKeeper()
			userAddr := suite.addrs[3]
			err := hardKeeper.Deposit(suite.ctx, userAddr, tc.args.initialDeposit)
			suite.Require().NoError(err)

			// Confirm that claim exists but no borrow reward indexes have been added
			claimAfterDeposit, found := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, suite.addrs[3])
			suite.Require().True(found)
			suite.Require().Equal(0, len(claimAfterDeposit.BorrowRewardIndexes))

			// User borrows (first time)
			err = hardKeeper.Borrow(suite.ctx, userAddr, tc.args.firstBorrow)
			suite.Require().NoError(err)

			// Confirm that claim's borrow reward indexes have been updated
			claimAfterFirstBorrow, found := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, suite.addrs[3])
			suite.Require().True(found)
			for _, coin := range tc.args.firstBorrow {
				_, hasIndex := claimAfterFirstBorrow.HasBorrowRewardIndex(coin.Denom)
				suite.Require().True(hasIndex)
			}
			suite.Require().True(len(claimAfterFirstBorrow.BorrowRewardIndexes) == len(tc.args.firstBorrow))

			// User modifies their Borrow by either repaying or borrowing more
			if tc.args.modification.repay {
				err = hardKeeper.Repay(suite.ctx, userAddr, userAddr, tc.args.modification.coins)
			} else {
				err = hardKeeper.Borrow(suite.ctx, userAddr, tc.args.modification.coins)
			}
			suite.Require().NoError(err)

			// Confirm that claim's borrow reward indexes contain expected values
			claimAfterModification, found := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, suite.addrs[3])
			suite.Require().True(found)
			for _, coin := range tc.args.modification.coins {
				_, hasIndex := claimAfterModification.HasBorrowRewardIndex(coin.Denom)
				if tc.args.modification.repay {
					// Only false if denom is repaid in full
					if tc.args.modification.coins.AmountOf(coin.Denom).GTE(tc.args.firstBorrow.AmountOf(coin.Denom)) {
						suite.Require().False(hasIndex)
					}
				} else {
					suite.Require().True(hasIndex)
				}
			}
			suite.Require().True(len(claimAfterModification.BorrowRewardIndexes) == len(tc.args.expectedBorrowIndexDenoms))
		})
	}
}

func (suite *KeeperTestSuite) TestSimulateHardBorrowRewardSynchronization() {
	type args struct {
		borrow                sdk.Coin
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
				borrow:                c("bnb", 10000000000),
				rewardsPerSecond:      cs(c("hard", 122354)),
				initialTime:           time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				blockTimes:            []int{10, 10, 10, 10, 10, 10, 10, 10, 10, 10},
				expectedRewardIndexes: types.RewardIndexes{types.NewRewardIndex("hard", d("0.001223540000173228"))},
				expectedRewards:       cs(c("hard", 12235400)),
			},
		},
		{
			"10 blocks - long block time",
			args{
				borrow:                c("bnb", 10000000000),
				rewardsPerSecond:      cs(c("hard", 122354)),
				initialTime:           time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				blockTimes:            []int{86400, 86400, 86400, 86400, 86400, 86400, 86400, 86400, 86400, 86400},
				expectedRewardIndexes: types.RewardIndexes{types.NewRewardIndex("hard", d("10.571385603126235340"))},
				expectedRewards:       cs(c("hard", 105713856031)),
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupWithGenState()
			suite.ctx = suite.ctx.WithBlockTime(tc.args.initialTime)

			// Mint coins to hard module account
			supplyKeeper := suite.app.GetSupplyKeeper()
			hardMaccCoins := sdk.NewCoins(sdk.NewCoin("usdx", sdk.NewInt(200000000)))
			supplyKeeper.MintCoins(suite.ctx, hardtypes.ModuleAccountName, hardMaccCoins)

			// setup incentive state
			params := types.NewParams(
				types.RewardPeriods{types.NewRewardPeriod(true, tc.args.borrow.Denom, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), tc.args.rewardsPerSecond[0])},
				types.MultiRewardPeriods{types.NewMultiRewardPeriod(true, tc.args.borrow.Denom, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), tc.args.rewardsPerSecond)},
				types.MultiRewardPeriods{types.NewMultiRewardPeriod(true, tc.args.borrow.Denom, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), tc.args.rewardsPerSecond)},
				types.RewardPeriods{types.NewRewardPeriod(true, tc.args.borrow.Denom, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), tc.args.rewardsPerSecond[0])},
				types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
				tc.args.initialTime.Add(time.Hour*24*365*5),
			)
			suite.keeper.SetParams(suite.ctx, params)
			suite.keeper.SetPreviousHardBorrowRewardAccrualTime(suite.ctx, tc.args.borrow.Denom, tc.args.initialTime)
			var rewardIndexes types.RewardIndexes
			for _, rewardCoin := range tc.args.rewardsPerSecond {
				rewardIndex := types.NewRewardIndex(rewardCoin.Denom, sdk.ZeroDec())
				rewardIndexes = append(rewardIndexes, rewardIndex)
			}
			suite.keeper.SetHardBorrowRewardIndexes(suite.ctx, tc.args.borrow.Denom, rewardIndexes)

			// Set up hard state (interest factor for the relevant denom)
			suite.hardKeeper.SetSupplyInterestFactor(suite.ctx, tc.args.borrow.Denom, sdk.MustNewDecFromStr("1.0"))
			suite.hardKeeper.SetBorrowInterestFactor(suite.ctx, tc.args.borrow.Denom, sdk.MustNewDecFromStr("1.0"))
			suite.hardKeeper.SetPreviousAccrualTime(suite.ctx, tc.args.borrow.Denom, tc.args.initialTime)

			// User deposits and borrows to increase total borrowed amount
			hardKeeper := suite.app.GetHardKeeper()
			userAddr := suite.addrs[3]
			err := hardKeeper.Deposit(suite.ctx, userAddr, sdk.NewCoins(sdk.NewCoin(tc.args.borrow.Denom, tc.args.borrow.Amount.Mul(sdk.NewInt(2)))))
			suite.Require().NoError(err)
			err = hardKeeper.Borrow(suite.ctx, userAddr, sdk.NewCoins(tc.args.borrow))
			suite.Require().NoError(err)

			// Check that Hard hooks initialized a HardLiquidityProviderClaim
			claim, found := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, suite.addrs[3])
			suite.Require().True(found)
			multiRewardIndex, _ := claim.BorrowRewardIndexes.GetRewardIndex(tc.args.borrow.Denom)
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

				// Accumulate hard borrow-side rewards
				multiRewardPeriod, found := suite.keeper.GetHardBorrowRewardPeriods(blockCtx, tc.args.borrow.Denom)
				suite.Require().True(found)
				err := suite.keeper.AccumulateHardBorrowRewards(blockCtx, multiRewardPeriod)
				suite.Require().NoError(err)
			}
			updatedBlockTime := suite.ctx.BlockTime().Add(time.Duration(int(time.Second) * timeElapsed))
			suite.ctx = suite.ctx.WithBlockTime(updatedBlockTime)

			// Confirm that the user's claim hasn't been synced
			claimPre, foundPre := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, suite.addrs[3])
			suite.Require().True(foundPre)
			multiRewardIndexPre, _ := claimPre.BorrowRewardIndexes.GetRewardIndex(tc.args.borrow.Denom)
			for _, expectedRewardIndex := range tc.args.expectedRewardIndexes {
				currRewardIndex, found := multiRewardIndexPre.RewardIndexes.GetRewardIndex(expectedRewardIndex.CollateralType)
				suite.Require().True(found)
				suite.Require().Equal(sdk.ZeroDec(), currRewardIndex.RewardFactor)
			}

			// Check that the synced claim held in memory has properly simulated syncing
			syncedClaim := suite.keeper.SimulateHardSynchronization(suite.ctx, claim)
			for _, expectedRewardIndex := range tc.args.expectedRewardIndexes {
				// Check that the user's claim's reward index matches the expected reward index
				multiRewardIndex, found := syncedClaim.BorrowRewardIndexes.GetRewardIndex(tc.args.borrow.Denom)
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
