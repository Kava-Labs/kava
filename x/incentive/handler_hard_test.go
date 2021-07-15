package incentive_test

import (
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/incentive/testutil"
	"github.com/kava-labs/kava/x/incentive/types"
	"github.com/kava-labs/kava/x/kavadist"
	validatorvesting "github.com/kava-labs/kava/x/validator-vesting"
)

func (suite *HandlerTestSuite) TestPayoutHardLiquidityProviderClaim() {
	type args struct {
		deposit                  sdk.Coins
		borrow                   sdk.Coins
		rewardsPerSecond         sdk.Coins
		multipliers              types.Multipliers
		multiplier               string
		timeElapsed              time.Duration
		expectedRewards          sdk.Coins
		expectedPeriods          vesting.Periods
		isPeriodicVestingAccount bool
	}
	type errArgs struct {
		expectPass bool
		contains   string
	}
	type test struct {
		name    string
		args    args
		errArgs errArgs
	}
	testCases := []test{
		{
			"single reward denom: valid 1 day",
			args{
				deposit:                  cs(c("bnb", 10000000000)),
				borrow:                   cs(c("bnb", 5000000000)),
				rewardsPerSecond:         cs(c("hard", 122354)),
				multipliers:              types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
				multiplier:               "large",
				timeElapsed:              86400 * time.Second,
				expectedRewards:          cs(c("hard", 21142771202)), // 10571385600 (deposit reward) + 10571385600 (borrow reward) + 2 for interest on deposit
				expectedPeriods:          vesting.Periods{vesting.Period{Length: 32918400, Amount: cs(c("hard", 21142771202))}},
				isPeriodicVestingAccount: true,
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"single reward denom: valid 10 days",
			args{
				deposit:                  cs(c("bnb", 10000000000)),
				borrow:                   cs(c("bnb", 5000000000)),
				rewardsPerSecond:         cs(c("hard", 122354)),
				multipliers:              types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
				multiplier:               "large",
				timeElapsed:              864000 * time.Second,
				expectedRewards:          cs(c("hard", 211427712008)), // 105713856000 (deposit reward) + 105713856000 (borrow reward) + 8 for interest on deposit
				expectedPeriods:          vesting.Periods{vesting.Period{Length: 32140800, Amount: cs(c("hard", 211427712008))}},
				isPeriodicVestingAccount: true,
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"multiple reward denoms: valid 1 day",
			args{
				deposit:                  cs(c("bnb", 10000000000)),
				borrow:                   cs(c("bnb", 5000000000)),
				rewardsPerSecond:         cs(c("hard", 122354), c("ukava", 122354)),
				multipliers:              types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
				multiplier:               "large",
				timeElapsed:              86400 * time.Second,
				expectedRewards:          cs(c("hard", 21142771202), c("ukava", 21142771202)), // 10571385600 (deposit reward) + 10571385600 (borrow reward) + 2 for interest on deposit
				expectedPeriods:          vesting.Periods{vesting.Period{Length: 32918400, Amount: cs(c("hard", 21142771202), c("ukava", 21142771202))}},
				isPeriodicVestingAccount: true,
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"multiple reward denoms: valid 10 days",
			args{
				deposit:                  cs(c("bnb", 10000000000)),
				borrow:                   cs(c("bnb", 5000000000)),
				rewardsPerSecond:         cs(c("hard", 122354), c("ukava", 122354)),
				multipliers:              types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
				multiplier:               "large",
				timeElapsed:              864000 * time.Second,
				expectedRewards:          cs(c("hard", 211427712008), c("ukava", 211427712008)), // 105713856000 (deposit reward) + 105713856000 (borrow reward) + 8 for interest on deposit
				expectedPeriods:          vesting.Periods{vesting.Period{Length: 32140800, Amount: cs(c("hard", 211427712008), c("ukava", 211427712008))}},
				isPeriodicVestingAccount: true,
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"multiple reward denoms with different rewards per second: valid 1 day",
			args{
				deposit:                  cs(c("bnb", 10000000000)),
				borrow:                   cs(c("bnb", 5000000000)),
				rewardsPerSecond:         cs(c("hard", 122354), c("ukava", 222222)),
				multipliers:              types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
				multiplier:               "large",
				timeElapsed:              86400 * time.Second,
				expectedRewards:          cs(c("hard", 21142771202), c("ukava", 38399961603)),
				expectedPeriods:          vesting.Periods{vesting.Period{Length: 32918400, Amount: cs(c("hard", 21142771202), c("ukava", 38399961603))}},
				isPeriodicVestingAccount: true,
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			userAddr := suite.addrs[3]
			authBulder := app.NewAuthGenesisBuilder().
				WithSimpleAccount(userAddr, cs(c("bnb", 1e15), c("ukava", 1e15), c("btcb", 1e15), c("xrp", 1e15), c("zzz", 1e15))).
				WithSimpleModuleAccount(kavadist.ModuleName, cs(c("hard", 1000000000000000000), c("ukava", 1000000000000000000)))

			incentBuilder := testutil.NewIncentiveGenesisBuilder().
				WithGenesisTime(suite.genesisTime).
				WithMultipliers(tc.args.multipliers)
			for _, c := range tc.args.deposit {
				incentBuilder = incentBuilder.WithSimpleSupplyRewardPeriod(c.Denom, tc.args.rewardsPerSecond)
			}
			for _, c := range tc.args.borrow {
				incentBuilder = incentBuilder.WithSimpleBorrowRewardPeriod(c.Denom, tc.args.rewardsPerSecond)
			}

			suite.SetupWithGenState(authBulder, incentBuilder, NewHardGenStateMulti(suite.genesisTime))

			// User deposits and borrows
			err := suite.hardKeeper.Deposit(suite.Ctx, userAddr, tc.args.deposit)
			suite.Require().NoError(err)
			err = suite.hardKeeper.Borrow(suite.Ctx, userAddr, tc.args.borrow)
			suite.Require().NoError(err)

			// Check that Hard hooks initialized a HardLiquidityProviderClaim that has 0 rewards
			claim, found := suite.keeper.GetHardLiquidityProviderClaim(suite.Ctx, userAddr)
			suite.Require().True(found)
			for _, coin := range tc.args.deposit {
				suite.Require().Equal(sdk.ZeroInt(), claim.Reward.AmountOf(coin.Denom))
			}

			// Accumulate supply and borrow rewards
			suite.NextBlockAfter(tc.args.timeElapsed)

			// Fetch pre-claim balances
			ak := suite.App.GetAccountKeeper()
			preClaimAcc := ak.GetAccount(suite.Ctx, userAddr)

			msg := types.NewMsgClaimHardReward(userAddr, tc.args.multiplier, nil)
			_, err = suite.handler(suite.Ctx, msg)

			if tc.errArgs.expectPass {
				suite.Require().NoError(err)

				// Check that user's balance has increased by expected reward amount
				postClaimAcc := ak.GetAccount(suite.Ctx, userAddr)
				suite.Require().Equal(preClaimAcc.GetCoins().Add(tc.args.expectedRewards...), postClaimAcc.GetCoins())

				if tc.args.isPeriodicVestingAccount {
					vacc, ok := postClaimAcc.(*vesting.PeriodicVestingAccount)
					suite.Require().True(ok)
					suite.Require().Equal(tc.args.expectedPeriods, vacc.VestingPeriods)
				}

				// Check that each claim reward coin's amount has been reset to 0
				claim, found := suite.keeper.GetHardLiquidityProviderClaim(suite.Ctx, userAddr)
				suite.Require().True(found)
				for _, claimRewardCoin := range claim.Reward {
					suite.Require().Equal(c(claimRewardCoin.Denom, 0), claimRewardCoin)
				}
			} else {
				suite.Require().Error(err)
				suite.Require().True(strings.Contains(err.Error(), tc.errArgs.contains))
			}
		})
	}
}

func (suite *HandlerTestSuite) TestPayoutHardLiquidityProviderClaimVVesting() {
	type args struct {
		deposit          sdk.Coins
		borrow           sdk.Coins
		rewardsPerSecond sdk.Coins
		multipliers      types.Multipliers
		multiplier       string
		timeElapsed      time.Duration
		expectedRewards  sdk.Coins
		expectedPeriods  vesting.Periods
	}
	type errArgs struct {
		expectPass bool
		contains   string
	}
	type test struct {
		name    string
		args    args
		errArgs errArgs
	}
	testCases := []test{
		{
			"single reward denom: valid 1 day",
			args{
				deposit:          cs(c("bnb", 10000000000)),
				borrow:           cs(c("bnb", 5000000000)),
				rewardsPerSecond: cs(c("hard", 122354)),
				multipliers:      types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
				multiplier:       "large",
				timeElapsed:      86400 * time.Second,
				expectedRewards:  cs(c("hard", 21142771202)),
				expectedPeriods:  vesting.Periods{vesting.Period{Length: 32918400, Amount: cs(c("hard", 21142771202))}},
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"single reward denom: valid 10 days",
			args{
				deposit:          cs(c("bnb", 10000000000)),
				borrow:           cs(c("bnb", 5000000000)),
				rewardsPerSecond: cs(c("hard", 122354)),
				multipliers:      types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
				multiplier:       "large",
				timeElapsed:      864000 * time.Second,
				expectedRewards:  cs(c("hard", 211427712008)),
				expectedPeriods:  vesting.Periods{vesting.Period{Length: 32140800, Amount: cs(c("hard", 211427712008))}},
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"multiple reward denoms: valid 1 day",
			args{
				deposit:          cs(c("bnb", 10000000000)),
				borrow:           cs(c("bnb", 5000000000)),
				rewardsPerSecond: cs(c("hard", 122354), c("ukava", 122354)),
				multipliers:      types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
				multiplier:       "large",
				timeElapsed:      86400 * time.Second,
				expectedRewards:  cs(c("hard", 21142771202), c("ukava", 21142771202)),
				expectedPeriods:  vesting.Periods{vesting.Period{Length: 32918400, Amount: cs(c("hard", 21142771202), c("ukava", 21142771202))}},
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"multiple reward denoms: valid 10 days",
			args{
				deposit:          cs(c("bnb", 10000000000)),
				borrow:           cs(c("bnb", 5000000000)),
				rewardsPerSecond: cs(c("hard", 122354), c("ukava", 122354)),
				multipliers:      types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
				multiplier:       "large",
				timeElapsed:      864000 * time.Second,
				expectedRewards:  cs(c("hard", 211427712008), c("ukava", 211427712008)),
				expectedPeriods:  vesting.Periods{vesting.Period{Length: 32140800, Amount: cs(c("hard", 211427712008), c("ukava", 211427712008))}},
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"multiple reward denoms with different rewards per second: valid 1 day",
			args{
				deposit:          cs(c("bnb", 10000000000)),
				borrow:           cs(c("bnb", 5000000000)),
				rewardsPerSecond: cs(c("hard", 122354), c("ukava", 222222)),
				multipliers:      types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
				multiplier:       "large",
				timeElapsed:      86400 * time.Second,
				expectedRewards:  cs(c("hard", 21142771202), c("ukava", 38399961603)),
				expectedPeriods:  vesting.Periods{vesting.Period{Length: 32918400, Amount: cs(c("hard", 21142771202), c("ukava", 38399961603))}},
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {

			userAddr := suite.addrs[3]

			bacc := auth.NewBaseAccount(userAddr, cs(c("bnb", 1e15), c("ukava", 1e15), c("btcb", 1e15), c("xrp", 1e15), c("zzz", 1e15)), nil, 0, 0)
			bva, err := vesting.NewBaseVestingAccount(bacc, cs(c("ukava", 400)), suite.genesisTime.Unix()+16)
			suite.Require().NoError(err)
			periods := vesting.Periods{
				vesting.Period{Length: int64(1), Amount: cs(c("ukava", 100))},
				vesting.Period{Length: int64(2), Amount: cs(c("ukava", 100))},
				vesting.Period{Length: int64(8), Amount: cs(c("ukava", 100))},
				vesting.Period{Length: int64(5), Amount: cs(c("ukava", 100))},
			}
			vva := validatorvesting.NewValidatorVestingAccountRaw(bva, suite.genesisTime.Unix(), periods, sdk.ConsAddress{}, nil, 90)

			authBulder := app.NewAuthGenesisBuilder().
				WithAccounts(vva).
				WithSimpleAccount(suite.addrs[2], cs()).
				WithSimpleModuleAccount(kavadist.ModuleName, cs(c("hard", 1000000000000000000), c("ukava", 1000000000000000000)))

			incentBuilder := testutil.NewIncentiveGenesisBuilder().
				WithGenesisTime(suite.genesisTime).
				WithMultipliers(tc.args.multipliers)
			for _, c := range tc.args.deposit {
				incentBuilder = incentBuilder.WithSimpleSupplyRewardPeriod(c.Denom, tc.args.rewardsPerSecond)
			}
			for _, c := range tc.args.borrow {
				incentBuilder = incentBuilder.WithSimpleBorrowRewardPeriod(c.Denom, tc.args.rewardsPerSecond)
			}

			suite.SetupWithGenState(authBulder, incentBuilder, NewHardGenStateMulti(suite.genesisTime))

			ak := suite.App.GetAccountKeeper()
			hardKeeper := suite.App.GetHardKeeper()

			// User deposits and borrows
			err = hardKeeper.Deposit(suite.Ctx, userAddr, tc.args.deposit)
			suite.Require().NoError(err)
			err = hardKeeper.Borrow(suite.Ctx, userAddr, tc.args.borrow)
			suite.Require().NoError(err)

			// Check that Hard hooks initialized a HardLiquidityProviderClaim that has 0 rewards
			claim, found := suite.keeper.GetHardLiquidityProviderClaim(suite.Ctx, userAddr)
			suite.Require().True(found)
			for _, coin := range tc.args.deposit {
				suite.Require().Equal(sdk.ZeroInt(), claim.Reward.AmountOf(coin.Denom))
			}

			// Accumulate supply and borrow rewards
			suite.NextBlockAfter(tc.args.timeElapsed)

			// Fetch pre-claim balances
			preClaimAcc := ak.GetAccount(suite.Ctx, suite.addrs[2])

			msg := types.NewMsgClaimHardRewardVVesting(userAddr, suite.addrs[2], tc.args.multiplier, nil)
			_, err = suite.handler(suite.Ctx, msg)

			if tc.errArgs.expectPass {
				suite.Require().NoError(err)

				// Check that user's balance has increased by expected reward amount
				postClaimAcc := ak.GetAccount(suite.Ctx, suite.addrs[2])
				suite.Require().Equal(preClaimAcc.GetCoins().Add(tc.args.expectedRewards...), postClaimAcc.GetCoins())

				vacc, ok := postClaimAcc.(*vesting.PeriodicVestingAccount)
				suite.Require().True(ok)
				suite.Require().Equal(tc.args.expectedPeriods, vacc.VestingPeriods)

				// Check that each claim reward coin's amount has been reset to 0
				claim, found := suite.keeper.GetHardLiquidityProviderClaim(suite.Ctx, suite.addrs[3])
				suite.Require().True(found)
				for _, claimRewardCoin := range claim.Reward {
					suite.Require().Equal(c(claimRewardCoin.Denom, 0), claimRewardCoin)
				}
			} else {
				suite.Require().Error(err)
				suite.Require().True(strings.Contains(err.Error(), tc.errArgs.contains))
			}
		})
	}
}
