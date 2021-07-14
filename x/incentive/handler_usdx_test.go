package incentive_test

import (
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/incentive"
	"github.com/kava-labs/kava/x/incentive/testutil"
	"github.com/kava-labs/kava/x/incentive/types"
	"github.com/kava-labs/kava/x/kavadist"
	validatorvesting "github.com/kava-labs/kava/x/validator-vesting"
)

func (suite *HandlerTestSuite) TestPayoutUSDXMintingClaim() {
	type args struct {
		ctype                    string
		rewardsPerSecond         sdk.Coin
		initialCollateral        sdk.Coin
		initialPrincipal         sdk.Coin
		multipliers              types.Multipliers
		multiplier               string
		timeElapsed              time.Duration
		expectedBalance          sdk.Coins
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
			"valid 1 day",
			args{
				ctype:                    "bnb-a",
				rewardsPerSecond:         c("ukava", 122354),
				initialCollateral:        c("bnb", 1000000000000),
				initialPrincipal:         c("usdx", 10000000000),
				multipliers:              types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
				multiplier:               "large",
				timeElapsed:              86400 * time.Second,
				expectedBalance:          cs(c("usdx", 10000000000), c("ukava", 10571385600)),
				expectedPeriods:          vesting.Periods{vesting.Period{Length: 32918400, Amount: cs(c("ukava", 10571385600))}},
				isPeriodicVestingAccount: true,
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"invalid zero rewards",
			args{
				ctype:                    "bnb-a",
				rewardsPerSecond:         c("ukava", 0),
				initialCollateral:        c("bnb", 1000000000000),
				initialPrincipal:         c("usdx", 10000000000),
				multipliers:              types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
				multiplier:               "large",
				timeElapsed:              86400 * time.Second,
				expectedBalance:          cs(c("usdx", 10000000000)),
				expectedPeriods:          vesting.Periods{},
				isPeriodicVestingAccount: false,
			},
			errArgs{
				expectPass: false,
				contains:   "claim amount rounds to zero",
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			userAddr := suite.addrs[0]
			authBulder := app.NewAuthGenesisBuilder().
				WithSimpleAccount(userAddr, cs(tc.args.initialCollateral)).
				WithSimpleModuleAccount(kavadist.ModuleName, cs(c("ukava", 1000000000000)))

			incentBuilder := testutil.NewIncentiveGenesisBuilder().
				WithGenesisTime(suite.genesisTime).
				WithSimpleUSDXRewardPeriod(tc.args.ctype, tc.args.rewardsPerSecond).
				WithMultipliers(tc.args.multipliers)

			suite.SetupWithGenState(authBulder, incentBuilder, NewHardGenStateMulti(suite.genesisTime))

			// setup cdp state
			err := suite.cdpKeeper.AddCdp(suite.ctx, userAddr, tc.args.initialCollateral, tc.args.initialPrincipal, tc.args.ctype)
			suite.Require().NoError(err)

			claim, found := suite.keeper.GetUSDXMintingClaim(suite.ctx, userAddr)
			suite.Require().True(found)
			suite.Require().Equal(sdk.ZeroDec(), claim.RewardIndexes[0].RewardFactor)

			suite.NextBlockAfter(tc.args.timeElapsed)

			msg := incentive.NewMsgClaimUSDXMintingReward(userAddr, tc.args.multiplier)
			_, err = suite.handler(suite.ctx, msg)

			if tc.errArgs.expectPass {
				suite.Require().NoError(err)

				acc := suite.GetAccount(userAddr)
				suite.Require().Equal(tc.args.expectedBalance, acc.GetCoins())

				if tc.args.isPeriodicVestingAccount {
					vacc, ok := acc.(*vesting.PeriodicVestingAccount)
					suite.Require().True(ok)
					suite.Require().Equal(tc.args.expectedPeriods, vacc.VestingPeriods)
				}

				claim, found := suite.keeper.GetUSDXMintingClaim(suite.ctx, userAddr)
				suite.Require().True(found)
				suite.Require().Equal(c("ukava", 0), claim.Reward)
			} else {
				suite.Require().Error(err)
				suite.Require().True(strings.Contains(err.Error(), tc.errArgs.contains))
			}
		})
	}
}

func (suite *HandlerTestSuite) TestPayoutUSDXMintingClaimVVesting() {
	type args struct {
		ctype             string
		rewardsPerSecond  sdk.Coin
		initialCollateral sdk.Coin
		initialPrincipal  sdk.Coin
		multipliers       types.Multipliers
		multiplier        string
		timeElapsed       time.Duration
		expectedBalance   sdk.Coins
		expectedPeriods   vesting.Periods
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
			"valid 1 day",
			args{
				ctype:             "bnb-a",
				rewardsPerSecond:  c("ukava", 122354),
				initialCollateral: c("bnb", 1e12),
				initialPrincipal:  c("usdx", 1e10),
				multipliers:       types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
				multiplier:        "large",
				timeElapsed:       86400 * time.Second,
				expectedBalance:   cs(c("ukava", 10571385600)),
				expectedPeriods:   vesting.Periods{vesting.Period{Length: 32918400, Amount: cs(c("ukava", 10571385600))}},
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"invalid zero rewards",
			args{
				ctype:             "bnb-a",
				rewardsPerSecond:  c("ukava", 0),
				initialCollateral: c("bnb", 1e12),
				initialPrincipal:  c("usdx", 1e10),
				multipliers:       types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
				multiplier:        "large",
				timeElapsed:       86400 * time.Second,
				expectedBalance:   cs(),
				expectedPeriods:   vesting.Periods{},
			},
			errArgs{
				expectPass: false,
				contains:   "claim amount rounds to zero",
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {

			bacc := auth.NewBaseAccount(suite.addrs[2], cs(tc.args.initialCollateral, c("ukava", 400)), nil, 0, 0)
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
				WithSimpleModuleAccount(kavadist.ModuleName, cs(c("ukava", 1e18))).
				WithSimpleAccount(suite.addrs[0], cs()) // the recipient address needs to be a instantiated account // TODO change?

			incentBuilder := testutil.NewIncentiveGenesisBuilder().
				WithGenesisTime(suite.genesisTime).
				WithSimpleUSDXRewardPeriod(tc.args.ctype, tc.args.rewardsPerSecond).
				WithMultipliers(tc.args.multipliers)

			suite.SetupWithGenState(authBulder, incentBuilder, NewHardGenStateMulti(suite.genesisTime))

			// setup cdp state
			cdpKeeper := suite.app.GetCDPKeeper()
			err = cdpKeeper.AddCdp(suite.ctx, suite.addrs[2], tc.args.initialCollateral, tc.args.initialPrincipal, tc.args.ctype)
			suite.Require().NoError(err)

			claim, found := suite.keeper.GetUSDXMintingClaim(suite.ctx, suite.addrs[2])
			suite.Require().True(found)
			suite.Require().Equal(sdk.ZeroDec(), claim.RewardIndexes[0].RewardFactor)

			// accumulate some usdx rewards
			suite.NextBlockAfter(tc.args.timeElapsed)

			msg := incentive.NewMsgClaimUSDXMintingRewardVVesting(suite.addrs[2], suite.addrs[0], tc.args.multiplier)
			_, err = suite.handler(suite.ctx, msg)

			if tc.errArgs.expectPass {
				suite.Require().NoError(err)
				ak := suite.app.GetAccountKeeper()
				acc := ak.GetAccount(suite.ctx, suite.addrs[0])
				suite.Require().Equal(tc.args.expectedBalance, acc.GetCoins())

				vacc, ok := acc.(*vesting.PeriodicVestingAccount)
				suite.Require().True(ok)
				suite.Require().Equal(tc.args.expectedPeriods, vacc.VestingPeriods)

				claim, found := suite.keeper.GetUSDXMintingClaim(suite.ctx, suite.addrs[2])
				suite.Require().True(found)
				suite.Require().Equal(c("ukava", 0), claim.Reward)
			} else {
				suite.Require().Error(err)
				suite.Require().True(strings.Contains(err.Error(), tc.errArgs.contains))
			}
		})
	}
}
