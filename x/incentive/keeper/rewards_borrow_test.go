package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/committee"
	committeekeeper "github.com/kava-labs/kava/x/committee/keeper"
	"github.com/kava-labs/kava/x/hard"
	hardkeeper "github.com/kava-labs/kava/x/hard/keeper"
	"github.com/kava-labs/kava/x/incentive/keeper"
	"github.com/kava-labs/kava/x/incentive/types"
)

// Test suite used for all keeper tests
type BorrowRewardsTestSuite struct {
	suite.Suite

	keeper          keeper.Keeper
	hardKeeper      hardkeeper.Keeper
	committeeKeeper committeekeeper.Keeper

	app app.TestApp
	ctx sdk.Context

	genesisTime time.Time
	addrs       []sdk.AccAddress
}

// SetupTest is run automatically before each suite test
func (suite *BorrowRewardsTestSuite) SetupTest() {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)

	_, suite.addrs = app.GeneratePrivKeyAddressPairs(5)

	suite.genesisTime = time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC)
}

func (suite *BorrowRewardsTestSuite) SetupApp() {
	suite.app = app.NewTestApp()

	suite.keeper = suite.app.GetIncentiveKeeper()
	suite.hardKeeper = suite.app.GetHardKeeper()
	suite.committeeKeeper = suite.app.GetCommitteeKeeper()

	suite.ctx = suite.app.NewContext(true, abci.Header{Height: 1, Time: suite.genesisTime})
}

func (suite *BorrowRewardsTestSuite) SetupWithGenState(authBuilder app.AuthGenesisBuilder, incentBuilder IncentiveGenesisBuilder, hardBuilder HardGenesisBuilder) {
	suite.SetupApp()

	suite.app.InitializeFromGenesisStatesWithTime(
		suite.genesisTime,
		authBuilder.BuildMarshalled(),
		NewPricefeedGenStateMultiFromTime(suite.genesisTime),
		hardBuilder.BuildMarshalled(),
		NewCommitteeGenesisState(suite.addrs[:2]),
		incentBuilder.BuildMarshalled(),
	)
}

func (suite *BorrowRewardsTestSuite) TestAccumulateHardBorrowRewards() {
	type args struct {
		borrow                sdk.Coin
		rewardsPerSecond      sdk.Coins
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
				timeElapsed:           7,
				expectedRewardIndexes: types.RewardIndexes{types.NewRewardIndex("hard", d("0.000000856478000001"))},
			},
		},
		{
			"single reward denom: 1 day",
			args{
				borrow:                c("bnb", 1000000000000),
				rewardsPerSecond:      cs(c("hard", 122354)),
				timeElapsed:           86400,
				expectedRewardIndexes: types.RewardIndexes{types.NewRewardIndex("hard", d("0.010571385600010177"))},
			},
		},
		{
			"single reward denom: 0 seconds",
			args{
				borrow:                c("bnb", 1000000000000),
				rewardsPerSecond:      cs(c("hard", 122354)),
				timeElapsed:           0,
				expectedRewardIndexes: types.RewardIndexes{types.NewRewardIndex("hard", d("0.0"))},
			},
		},
		{
			"multiple reward denoms: 7 seconds",
			args{
				borrow:           c("bnb", 1000000000000),
				rewardsPerSecond: cs(c("hard", 122354), c("ukava", 122354)),
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
			userAddr := suite.addrs[3]
			authBuilder := app.NewAuthGenesisBuilder().WithSimpleAccount(
				userAddr,
				cs(c("bnb", 1e15), c("ukava", 1e15), c("btcb", 1e15), c("xrp", 1e15), c("zzz", 1e15)),
			)

			incentBuilder := NewIncentiveGenesisBuilder().
				WithGenesisTime(suite.genesisTime).
				WithSimpleBorrowRewardPeriod(tc.args.borrow.Denom, tc.args.rewardsPerSecond)

			suite.SetupWithGenState(authBuilder, incentBuilder, NewHardGenStateMulti(suite.genesisTime))

			// User deposits and borrows to increase total borrowed amount
			err := suite.hardKeeper.Deposit(suite.ctx, userAddr, sdk.NewCoins(sdk.NewCoin(tc.args.borrow.Denom, tc.args.borrow.Amount.Mul(sdk.NewInt(2)))))
			suite.Require().NoError(err)
			err = suite.hardKeeper.Borrow(suite.ctx, userAddr, sdk.NewCoins(tc.args.borrow))
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

func (suite *BorrowRewardsTestSuite) TestInitializeHardBorrowRewards() {

	type args struct {
		moneyMarketRewardDenoms          map[string]sdk.Coins
		deposit                          sdk.Coins
		borrow                           sdk.Coins
		expectedClaimBorrowRewardIndexes types.MultiRewardIndexes
	}
	type test struct {
		name string
		args args
	}

	standardMoneyMarketRewardDenoms := map[string]sdk.Coins{
		"bnb":  cs(c("hard", 1)),
		"btcb": cs(c("hard", 1), c("ukava", 1)),
	}

	testCases := []test{
		{
			"single deposit denom, single reward denom",
			args{
				moneyMarketRewardDenoms: standardMoneyMarketRewardDenoms,
				deposit:                 cs(c("bnb", 1000000000000)),
				borrow:                  cs(c("bnb", 100000000000)),
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
			userAddr := suite.addrs[3]
			authBuilder := app.NewAuthGenesisBuilder().WithSimpleAccount(
				userAddr,
				cs(c("bnb", 1e15), c("ukava", 1e15), c("btcb", 1e15), c("xrp", 1e15), c("zzz", 1e15)),
			)

			incentBuilder := NewIncentiveGenesisBuilder().WithGenesisTime(suite.genesisTime)
			for moneyMarketDenom, rewardsPerSecond := range tc.args.moneyMarketRewardDenoms {
				incentBuilder = incentBuilder.WithSimpleBorrowRewardPeriod(moneyMarketDenom, rewardsPerSecond)
			}

			suite.SetupWithGenState(authBuilder, incentBuilder, NewHardGenStateMulti(suite.genesisTime))

			// User deposits
			err := suite.hardKeeper.Deposit(suite.ctx, userAddr, tc.args.deposit)
			suite.Require().NoError(err)
			// User borrows
			err = suite.hardKeeper.Borrow(suite.ctx, userAddr, tc.args.borrow)
			suite.Require().NoError(err)

			claim, foundClaim := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, userAddr)
			suite.Require().True(foundClaim)
			suite.Require().Equal(tc.args.expectedClaimBorrowRewardIndexes, claim.BorrowRewardIndexes)
		})
	}
}

func (suite *BorrowRewardsTestSuite) TestSynchronizeHardBorrowReward() {
	type args struct {
		incentiveBorrowRewardDenom   string
		borrow                       sdk.Coin
		rewardsPerSecond             sdk.Coins
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
				blockTimes:                 []int{10, 10, 10, 10, 10, 10, 10, 10, 10, 10},
				expectedRewardIndexes: types.RewardIndexes{
					types.NewRewardIndex("hard", d("0.001223540000173228")),
					types.NewRewardIndex("ukava", d("0.005555550000786558")),
				},
				expectedRewards: cs(c("hard", 12235400), c("ukava", 55555500)),
			},
		},
		{
			"denom is in incentive's hard borrow reward params and has rewards; add new reward type",
			args{
				incentiveBorrowRewardDenom: "bnb",
				borrow:                     c("bnb", 10000000000),
				rewardsPerSecond:           cs(c("hard", 122354)),
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
				rewardsPerSecond:           nil,
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
			"denom is in hard's money market params but not in incentive's hard supply reward params; add multiple reward types",
			args{
				incentiveBorrowRewardDenom: "bnb",
				borrow:                     c("zzz", 10000000000),
				rewardsPerSecond:           nil,
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
			userAddr := suite.addrs[3]
			authBuilder := app.NewAuthGenesisBuilder().
				WithSimpleAccount(suite.addrs[2], cs(c("ukava", 1e9))).
				WithSimpleAccount(userAddr, cs(c("bnb", 1e15), c("ukava", 1e15), c("btcb", 1e15), c("xrp", 1e15), c("zzz", 1e15)))

			incentBuilder := NewIncentiveGenesisBuilder().WithGenesisTime(suite.genesisTime)
			if tc.args.rewardsPerSecond != nil {
				incentBuilder = incentBuilder.WithSimpleBorrowRewardPeriod(tc.args.incentiveBorrowRewardDenom, tc.args.rewardsPerSecond)
			}
			// Set the minimum borrow to 0 to allow testing small borrows
			hardBuilder := NewHardGenStateMulti(suite.genesisTime).WithMinBorrow(sdk.ZeroDec())

			suite.SetupWithGenState(authBuilder, incentBuilder, hardBuilder)

			// Borrow a fixed amount from another user to dilute primary user's rewards per second.
			suite.Require().NoError(
				suite.hardKeeper.Deposit(suite.ctx, suite.addrs[2], cs(c("ukava", 200_000_000))),
			)
			suite.Require().NoError(
				suite.hardKeeper.Borrow(suite.ctx, suite.addrs[2], cs(c("ukava", 100_000_000))),
			)

			// User deposits and borrows to increase total borrowed amount
			err := suite.hardKeeper.Deposit(suite.ctx, userAddr, sdk.NewCoins(sdk.NewCoin(tc.args.borrow.Denom, tc.args.borrow.Amount.Mul(sdk.NewInt(2)))))
			suite.Require().NoError(err)
			err = suite.hardKeeper.Borrow(suite.ctx, userAddr, sdk.NewCoins(tc.args.borrow))
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
			borrow, found := suite.hardKeeper.GetBorrow(suite.ctx, userAddr)
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
				newMultiRewardPeriod := types.NewMultiRewardPeriod(true, tc.args.borrow.Denom, suite.genesisTime, suite.genesisTime.Add(time.Hour*24*365*4), tc.args.updatedRewardsPerSecond)
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
			borrow, found = suite.hardKeeper.GetBorrow(suite.ctx, userAddr)
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

func (suite *BorrowRewardsTestSuite) TestUpdateHardBorrowIndexDenoms() {
	type withdrawModification struct {
		coins sdk.Coins
		repay bool
	}

	type args struct {
		initialDeposit            sdk.Coins
		firstBorrow               sdk.Coins
		modification              withdrawModification
		rewardsPerSecond          sdk.Coins
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
				expectedBorrowIndexDenoms: []string{"ukava"},
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			userAddr := suite.addrs[3]
			authBuilder := app.NewAuthGenesisBuilder().
				WithSimpleAccount(
					userAddr,
					cs(c("bnb", 1e15), c("ukava", 1e15), c("btcb", 1e15), c("xrp", 1e15), c("zzz", 1e15)),
				).
				WithSimpleAccount(
					suite.addrs[0],
					cs(c("bnb", 1e15), c("ukava", 1e15), c("btcb", 1e15), c("xrp", 1e15), c("zzz", 1e15)),
				)

			incentBuilder := NewIncentiveGenesisBuilder().
				WithGenesisTime(suite.genesisTime).
				WithSimpleBorrowRewardPeriod("bnb", tc.args.rewardsPerSecond).
				WithSimpleBorrowRewardPeriod("ukava", tc.args.rewardsPerSecond).
				WithSimpleBorrowRewardPeriod("btcb", tc.args.rewardsPerSecond).
				WithSimpleBorrowRewardPeriod("xrp", tc.args.rewardsPerSecond)

			suite.SetupWithGenState(authBuilder, incentBuilder, NewHardGenStateMulti(suite.genesisTime))

			// Fill the hard supply to allow user to borrow
			err := suite.hardKeeper.Deposit(suite.ctx, suite.addrs[0], tc.args.firstBorrow.Add(tc.args.modification.coins...))
			suite.Require().NoError(err)

			// User deposits initial funds (so that user can borrow)
			err = suite.hardKeeper.Deposit(suite.ctx, userAddr, tc.args.initialDeposit)
			suite.Require().NoError(err)

			// Confirm that claim exists but no borrow reward indexes have been added
			claimAfterDeposit, found := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, userAddr)
			suite.Require().True(found)
			suite.Require().Equal(0, len(claimAfterDeposit.BorrowRewardIndexes))

			// User borrows (first time)
			err = suite.hardKeeper.Borrow(suite.ctx, userAddr, tc.args.firstBorrow)
			suite.Require().NoError(err)

			// Confirm that claim's borrow reward indexes have been updated
			claimAfterFirstBorrow, found := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, userAddr)
			suite.Require().True(found)
			for _, coin := range tc.args.firstBorrow {
				_, hasIndex := claimAfterFirstBorrow.HasBorrowRewardIndex(coin.Denom)
				suite.Require().True(hasIndex)
			}
			suite.Require().True(len(claimAfterFirstBorrow.BorrowRewardIndexes) == len(tc.args.firstBorrow))

			// User modifies their Borrow by either repaying or borrowing more
			if tc.args.modification.repay {
				err = suite.hardKeeper.Repay(suite.ctx, userAddr, userAddr, tc.args.modification.coins)
			} else {
				err = suite.hardKeeper.Borrow(suite.ctx, userAddr, tc.args.modification.coins)
			}
			suite.Require().NoError(err)

			// Confirm that claim's borrow reward indexes contain expected values
			claimAfterModification, found := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, userAddr)
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

func (suite *BorrowRewardsTestSuite) TestSimulateHardBorrowRewardSynchronization() {
	type args struct {
		borrow                sdk.Coin
		rewardsPerSecond      sdk.Coins
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
				blockTimes:            []int{86400, 86400, 86400, 86400, 86400, 86400, 86400, 86400, 86400, 86400},
				expectedRewardIndexes: types.RewardIndexes{types.NewRewardIndex("hard", d("10.571385603126235340"))},
				expectedRewards:       cs(c("hard", 105713856031)),
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			userAddr := suite.addrs[3]
			authBuilder := app.NewAuthGenesisBuilder().WithSimpleAccount(userAddr, cs(c("bnb", 1e15), c("ukava", 1e15), c("btcb", 1e15), c("xrp", 1e15), c("zzz", 1e15)))

			incentBuilder := NewIncentiveGenesisBuilder().
				WithGenesisTime(suite.genesisTime).
				WithSimpleBorrowRewardPeriod(tc.args.borrow.Denom, tc.args.rewardsPerSecond)

			suite.SetupWithGenState(authBuilder, incentBuilder, NewHardGenStateMulti(suite.genesisTime))

			// User deposits and borrows to increase total borrowed amount
			err := suite.hardKeeper.Deposit(suite.ctx, userAddr, sdk.NewCoins(sdk.NewCoin(tc.args.borrow.Denom, tc.args.borrow.Amount.Mul(sdk.NewInt(2)))))
			suite.Require().NoError(err)
			err = suite.hardKeeper.Borrow(suite.ctx, userAddr, sdk.NewCoins(tc.args.borrow))
			suite.Require().NoError(err)

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
			claimPre, foundPre := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, userAddr)
			suite.Require().True(foundPre)
			multiRewardIndexPre, _ := claimPre.BorrowRewardIndexes.GetRewardIndex(tc.args.borrow.Denom)
			for _, expectedRewardIndex := range tc.args.expectedRewardIndexes {
				currRewardIndex, found := multiRewardIndexPre.RewardIndexes.GetRewardIndex(expectedRewardIndex.CollateralType)
				suite.Require().True(found)
				suite.Require().Equal(sdk.ZeroDec(), currRewardIndex.RewardFactor)
			}

			// Check that the synced claim held in memory has properly simulated syncing
			syncedClaim := suite.keeper.SimulateHardSynchronization(suite.ctx, claimPre)
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

func TestBorrowRewardsTestSuite(t *testing.T) {
	suite.Run(t, new(BorrowRewardsTestSuite))
}
