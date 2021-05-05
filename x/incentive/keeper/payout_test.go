package keeper_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	supplyexported "github.com/cosmos/cosmos-sdk/x/supply/exported"
	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	cdpkeeper "github.com/kava-labs/kava/x/cdp/keeper"
	cdptypes "github.com/kava-labs/kava/x/cdp/types"
	"github.com/kava-labs/kava/x/hard"
	hardkeeper "github.com/kava-labs/kava/x/hard/keeper"
	"github.com/kava-labs/kava/x/incentive/keeper"
	"github.com/kava-labs/kava/x/incentive/types"
	"github.com/kava-labs/kava/x/kavadist"
)

// Test suite used for all keeper tests
type PayoutTestSuite struct {
	suite.Suite

	keeper     keeper.Keeper
	hardKeeper hardkeeper.Keeper
	cdpKeeper  cdpkeeper.Keeper

	app   app.TestApp
	ctx   sdk.Context
	addrs []sdk.AccAddress
}

// SetupTest is run automatically before each suite test
func (suite *PayoutTestSuite) SetupTest() {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)

	_, suite.addrs = app.GeneratePrivKeyAddressPairs(5)
}

func (suite *PayoutTestSuite) SetupApp() {
	suite.app = app.NewTestApp()

	suite.keeper = suite.app.GetIncentiveKeeper()
	suite.hardKeeper = suite.app.GetHardKeeper()
	suite.cdpKeeper = suite.app.GetCDPKeeper()

	suite.ctx = suite.app.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
}

func (suite *PayoutTestSuite) SetupWithGenState(authBuilder AuthGenesisBuilder, incentBuilder incentiveGenesisBuilder) {
	suite.SetupApp()

	suite.app.InitializeFromGenesisStates(
		authBuilder.BuildMarshalled(),
		NewPricefeedGenStateMulti(),
		NewCDPGenStateMulti(),
		NewHardGenStateMulti(),
		incentBuilder.buildMarshalled(),
	)
}

func (suite *PayoutTestSuite) getAccount(addr sdk.AccAddress) authexported.Account {
	ak := suite.app.GetAccountKeeper()
	return ak.GetAccount(suite.ctx, addr)
}

func (suite *PayoutTestSuite) getModuleAccount(name string) supplyexported.ModuleAccountI {
	sk := suite.app.GetSupplyKeeper()
	return sk.GetModuleAccount(suite.ctx, name)
}

func (suite *PayoutTestSuite) TestPayoutUSDXMintingClaim() {
	type args struct {
		ctype                    string
		rewardsPerSecond         sdk.Coin
		initialTime              time.Time
		initialCollateral        sdk.Coin
		initialPrincipal         sdk.Coin
		multipliers              types.Multipliers
		multiplier               types.MultiplierName
		timeElapsed              int
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
				initialTime:              time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				initialCollateral:        c("bnb", 1000000000000),
				initialPrincipal:         c("usdx", 10000000000),
				multipliers:              types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
				multiplier:               types.MultiplierName("large"),
				timeElapsed:              86400,
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
				initialTime:              time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				initialCollateral:        c("bnb", 1000000000000),
				initialPrincipal:         c("usdx", 10000000000),
				multipliers:              types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
				multiplier:               types.MultiplierName("large"),
				timeElapsed:              86400,
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
			authBulder := NewAuthGenesisBuilder().
				WithSimpleAccount(userAddr, cs(tc.args.initialCollateral)).
				WithSimpleModuleAccount(kavadist.ModuleName, cs(c("ukava", 1000000000000)))

			incentBuilder := newIncentiveGenesisBuilder().
				withGenesisTime(tc.args.initialTime).
				withSimpleUSDXRewardPeriod(tc.args.ctype, tc.args.rewardsPerSecond).
				withMultipliers(tc.args.multipliers)

			suite.SetupWithGenState(authBulder, incentBuilder)
			suite.ctx = suite.ctx.WithBlockTime(tc.args.initialTime)

			// setup cdp state
			err := suite.cdpKeeper.AddCdp(suite.ctx, userAddr, tc.args.initialCollateral, tc.args.initialPrincipal, tc.args.ctype)
			suite.Require().NoError(err)

			claim, found := suite.keeper.GetUSDXMintingClaim(suite.ctx, userAddr)
			suite.Require().True(found)
			suite.Require().Equal(sdk.ZeroDec(), claim.RewardIndexes[0].RewardFactor)

			updatedBlockTime := suite.ctx.BlockTime().Add(time.Duration(int(time.Second) * tc.args.timeElapsed))
			suite.ctx = suite.ctx.WithBlockTime(updatedBlockTime)
			rewardPeriod, found := suite.keeper.GetUSDXMintingRewardPeriod(suite.ctx, tc.args.ctype)
			suite.Require().True(found)
			err = suite.keeper.AccumulateUSDXMintingRewards(suite.ctx, rewardPeriod)
			suite.Require().NoError(err)

			err = suite.keeper.ClaimUSDXMintingReward(suite.ctx, userAddr, tc.args.multiplier)

			if tc.errArgs.expectPass {
				suite.Require().NoError(err)
				ak := suite.app.GetAccountKeeper()
				acc := ak.GetAccount(suite.ctx, userAddr)
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

func (suite *PayoutTestSuite) TestPayoutHardLiquidityProviderClaim() {
	type args struct {
		deposit                  sdk.Coins
		borrow                   sdk.Coins
		rewardsPerSecond         sdk.Coins
		initialTime              time.Time
		multipliers              types.Multipliers
		multiplier               types.MultiplierName
		timeElapsed              int64
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
				initialTime:              time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				multipliers:              types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
				multiplier:               types.MultiplierName("large"),
				timeElapsed:              86400,
				expectedRewards:          cs(c("hard", 21142771200)), // 10571385600 (deposit reward) + 10571385600 (borrow reward)
				expectedPeriods:          vesting.Periods{vesting.Period{Length: 32918400, Amount: cs(c("hard", 21142771200))}},
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
				initialTime:              time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				multipliers:              types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
				multiplier:               types.MultiplierName("large"),
				timeElapsed:              864000,
				expectedRewards:          cs(c("hard", 211427712000)), // 105713856000 (deposit reward) + 105713856000 (borrow reward)
				expectedPeriods:          vesting.Periods{vesting.Period{Length: 32140800, Amount: cs(c("hard", 211427712000))}},
				isPeriodicVestingAccount: true,
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		// {
		// 	"invalid zero rewards",
		// 	args{
		// 		deposit:                  cs(c("bnb", 10000000000)),
		// 		borrow:                   cs(c("bnb", 5000000000)),
		// 		rewardsPerSecond:         cs(c("hard", 0)),
		// 		initialTime:              time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
		// 		multipliers:              types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
		// 		multiplier:               types.MultiplierName("large"),
		// 		timeElapsed:              86400,
		// 		expectedRewards:          cs(c("hard", 0)),
		// 		expectedPeriods:          vesting.Periods{},
		// 		isPeriodicVestingAccount: false,
		// 	},
		// 	errArgs{
		// 		expectPass: false,
		// 		contains:   "claim amount rounds to zero",
		// 	},
		// },
		{
			"multiple reward denoms: valid 1 day",
			args{
				deposit:                  cs(c("bnb", 10000000000)),
				borrow:                   cs(c("bnb", 5000000000)),
				rewardsPerSecond:         cs(c("hard", 122354), c("ukava", 122354)),
				initialTime:              time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				multipliers:              types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
				multiplier:               types.MultiplierName("large"),
				timeElapsed:              86400,
				expectedRewards:          cs(c("hard", 21142771200), c("ukava", 21142771200)), // 10571385600 (deposit reward) + 10571385600 (borrow reward)
				expectedPeriods:          vesting.Periods{vesting.Period{Length: 32918400, Amount: cs(c("hard", 21142771200), c("ukava", 21142771200))}},
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
				initialTime:              time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				multipliers:              types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
				multiplier:               types.MultiplierName("large"),
				timeElapsed:              864000,
				expectedRewards:          cs(c("hard", 211427712000), c("ukava", 211427712000)), // 105713856000 (deposit reward) + 105713856000 (borrow reward)
				expectedPeriods:          vesting.Periods{vesting.Period{Length: 32140800, Amount: cs(c("hard", 211427712000), c("ukava", 211427712000))}},
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
				initialTime:              time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				multipliers:              types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
				multiplier:               types.MultiplierName("large"),
				timeElapsed:              86400,
				expectedRewards:          cs(c("hard", 21142771200), c("ukava", 38399961600)),
				expectedPeriods:          vesting.Periods{vesting.Period{Length: 32918400, Amount: cs(c("hard", 21142771200), c("ukava", 38399961600))}},
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
			authBulder := NewAuthGenesisBuilder().
				WithSimpleAccount(userAddr, cs(c("bnb", 1e15), c("ukava", 1e15), c("btcb", 1e15), c("xrp", 1e15), c("zzz", 1e15))).
				WithSimpleModuleAccount(kavadist.ModuleName, cs(c("hard", 1000000000000000000), c("ukava", 1000000000000000000)))

			incentBuilder := newIncentiveGenesisBuilder().
				withGenesisTime(tc.args.initialTime).
				withMultipliers(tc.args.multipliers)
			for _, c := range tc.args.deposit {
				incentBuilder = incentBuilder.withSimpleSupplyRewardPeriod(c.Denom, tc.args.rewardsPerSecond)
			}
			for _, c := range tc.args.borrow {
				incentBuilder = incentBuilder.withSimpleBorrowRewardPeriod(c.Denom, tc.args.rewardsPerSecond)
			}

			suite.SetupWithGenState(authBulder, incentBuilder)
			suite.ctx = suite.ctx.WithBlockTime(tc.args.initialTime)

			// User deposits and borrows
			err := suite.hardKeeper.Deposit(suite.ctx, userAddr, tc.args.deposit)
			suite.Require().NoError(err)
			err = suite.hardKeeper.Borrow(suite.ctx, userAddr, tc.args.borrow)
			suite.Require().NoError(err)

			// Check that Hard hooks initialized a HardLiquidityProviderClaim that has 0 rewards
			claim, found := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, userAddr)
			suite.Require().True(found)
			for _, coin := range tc.args.deposit {
				suite.Require().Equal(sdk.ZeroInt(), claim.Reward.AmountOf(coin.Denom))
			}

			// Set up future runtime context
			runAtTime := time.Unix(suite.ctx.BlockTime().Unix()+(tc.args.timeElapsed), 0)
			runCtx := suite.ctx.WithBlockTime(runAtTime)

			// Run Hard begin blocker
			hard.BeginBlocker(runCtx, suite.hardKeeper)

			// Accumulate supply rewards for each deposit denom
			for _, coin := range tc.args.deposit {
				rewardPeriod, found := suite.keeper.GetHardSupplyRewardPeriods(runCtx, coin.Denom)
				suite.Require().True(found)
				err = suite.keeper.AccumulateHardSupplyRewards(runCtx, rewardPeriod)
				suite.Require().NoError(err)
			}

			// Accumulate borrow rewards for each deposit denom
			for _, coin := range tc.args.borrow {
				rewardPeriod, found := suite.keeper.GetHardBorrowRewardPeriods(runCtx, coin.Denom)
				suite.Require().True(found)
				err = suite.keeper.AccumulateHardBorrowRewards(runCtx, rewardPeriod)
				suite.Require().NoError(err)
			}

			// Sync hard supply rewards
			deposit, found := suite.hardKeeper.GetDeposit(suite.ctx, userAddr)
			suite.Require().True(found)
			suite.keeper.SynchronizeHardSupplyReward(suite.ctx, deposit)

			// Sync hard borrow rewards
			borrow, found := suite.hardKeeper.GetBorrow(suite.ctx, userAddr)
			suite.Require().True(found)
			suite.keeper.SynchronizeHardBorrowReward(suite.ctx, borrow)

			// Fetch pre-claim balances
			ak := suite.app.GetAccountKeeper()
			preClaimAcc := ak.GetAccount(runCtx, userAddr)

			err = suite.keeper.ClaimHardReward(runCtx, userAddr, tc.args.multiplier)
			if tc.errArgs.expectPass {
				suite.Require().NoError(err)

				// Check that user's balance has increased by expected reward amount
				postClaimAcc := ak.GetAccount(suite.ctx, userAddr)
				suite.Require().Equal(preClaimAcc.GetCoins().Add(tc.args.expectedRewards...), postClaimAcc.GetCoins())

				if tc.args.isPeriodicVestingAccount {
					vacc, ok := postClaimAcc.(*vesting.PeriodicVestingAccount)
					suite.Require().True(ok)
					suite.Require().Equal(tc.args.expectedPeriods, vacc.VestingPeriods)
				}

				// Check that each claim reward coin's amount has been reset to 0
				claim, found := suite.keeper.GetHardLiquidityProviderClaim(runCtx, userAddr)
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

func (suite *PayoutTestSuite) TestSendCoinsToPeriodicVestingAccount() {
	type accountArgs struct {
		periods          vesting.Periods
		origVestingCoins sdk.Coins
		startTime        int64
		endTime          int64
	}
	type args struct {
		accArgs             accountArgs
		period              vesting.Period
		ctxTime             time.Time
		mintModAccountCoins bool
		expectedPeriods     vesting.Periods
		expectedStartTime   int64
		expectedEndTime     int64
	}
	type errArgs struct {
		expectErr bool
		contains  string
	}
	type testCase struct {
		name    string
		args    args
		errArgs errArgs
	}
	type testCases []testCase

	tests := testCases{
		{
			name: "insert period at beginning schedule",
			args: args{
				accArgs: accountArgs{
					periods: vesting.Periods{
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))}},
					origVestingCoins: cs(c("ukava", 20)),
					startTime:        100,
					endTime:          120,
				},
				period:              vesting.Period{Length: 2, Amount: cs(c("ukava", 6))},
				ctxTime:             time.Unix(101, 0),
				mintModAccountCoins: true,
				expectedPeriods: vesting.Periods{
					vesting.Period{Length: 3, Amount: cs(c("ukava", 6))},
					vesting.Period{Length: 2, Amount: cs(c("ukava", 5))},
					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))}},
				expectedStartTime: 100,
				expectedEndTime:   120,
			},
			errArgs: errArgs{
				expectErr: false,
				contains:  "",
			},
		},
		{
			name: "insert period at beginning with new start time",
			args: args{
				accArgs: accountArgs{
					periods: vesting.Periods{
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))}},
					origVestingCoins: cs(c("ukava", 20)),
					startTime:        100,
					endTime:          120,
				},
				period:              vesting.Period{Length: 7, Amount: cs(c("ukava", 6))},
				ctxTime:             time.Unix(80, 0),
				mintModAccountCoins: true,
				expectedPeriods: vesting.Periods{
					vesting.Period{Length: 7, Amount: cs(c("ukava", 6))},
					vesting.Period{Length: 18, Amount: cs(c("ukava", 5))},
					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))}},
				expectedStartTime: 80,
				expectedEndTime:   120,
			},
			errArgs: errArgs{
				expectErr: false,
				contains:  "",
			},
		},
		{
			name: "insert period in middle of schedule",
			args: args{
				accArgs: accountArgs{
					periods: vesting.Periods{
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))}},
					origVestingCoins: cs(c("ukava", 20)),
					startTime:        100,
					endTime:          120,
				},
				period:              vesting.Period{Length: 7, Amount: cs(c("ukava", 6))},
				ctxTime:             time.Unix(101, 0),
				mintModAccountCoins: true,
				expectedPeriods: vesting.Periods{
					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vesting.Period{Length: 3, Amount: cs(c("ukava", 6))},
					vesting.Period{Length: 2, Amount: cs(c("ukava", 5))},
					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))}},
				expectedStartTime: 100,
				expectedEndTime:   120,
			},
			errArgs: errArgs{
				expectErr: false,
				contains:  "",
			},
		},
		{
			name: "append to end of schedule",
			args: args{
				accArgs: accountArgs{
					periods: vesting.Periods{
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))}},
					origVestingCoins: cs(c("ukava", 20)),
					startTime:        100,
					endTime:          120,
				},
				period:              vesting.Period{Length: 7, Amount: cs(c("ukava", 6))},
				ctxTime:             time.Unix(125, 0),
				mintModAccountCoins: true,
				expectedPeriods: vesting.Periods{
					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vesting.Period{Length: 12, Amount: cs(c("ukava", 6))}},
				expectedStartTime: 100,
				expectedEndTime:   132,
			},
			errArgs: errArgs{
				expectErr: false,
				contains:  "",
			},
		},
		{
			name: "add coins to existing period",
			args: args{
				accArgs: accountArgs{
					periods: vesting.Periods{
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))}},
					origVestingCoins: cs(c("ukava", 20)),
					startTime:        100,
					endTime:          120,
				},
				period:              vesting.Period{Length: 5, Amount: cs(c("ukava", 6))},
				ctxTime:             time.Unix(110, 0),
				mintModAccountCoins: true,
				expectedPeriods: vesting.Periods{
					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vesting.Period{Length: 5, Amount: cs(c("ukava", 11))},
					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))}},
				expectedStartTime: 100,
				expectedEndTime:   120,
			},
			errArgs: errArgs{
				expectErr: false,
				contains:  "",
			},
		},
		{
			name: "insufficient mod account balance",
			args: args{
				accArgs: accountArgs{
					periods: vesting.Periods{
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))}},
					origVestingCoins: cs(c("ukava", 20)),
					startTime:        100,
					endTime:          120,
				},
				period:              vesting.Period{Length: 7, Amount: cs(c("ukava", 6))},
				ctxTime:             time.Unix(125, 0),
				mintModAccountCoins: false,
				expectedPeriods: vesting.Periods{
					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vesting.Period{Length: 12, Amount: cs(c("ukava", 6))}},
				expectedStartTime: 100,
				expectedEndTime:   132,
			},
			errArgs: errArgs{
				expectErr: true,
				contains:  "insufficient funds",
			},
		},
		{
			name: "add large period mid schedule",
			args: args{
				accArgs: accountArgs{
					periods: vesting.Periods{
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))}},
					origVestingCoins: cs(c("ukava", 20)),
					startTime:        100,
					endTime:          120,
				},
				period:              vesting.Period{Length: 50, Amount: cs(c("ukava", 6))},
				ctxTime:             time.Unix(110, 0),
				mintModAccountCoins: true,
				expectedPeriods: vesting.Periods{
					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vesting.Period{Length: 40, Amount: cs(c("ukava", 6))}},
				expectedStartTime: 100,
				expectedEndTime:   160,
			},
			errArgs: errArgs{
				expectErr: false,
				contains:  "",
			},
		},
	}
	for _, tc := range tests {
		suite.Run(tc.name, func() {
			authBuilder := NewAuthGenesisBuilder().WithSimplePeriodicVestingAccount(
				suite.addrs[0],
				tc.args.accArgs.origVestingCoins,
				tc.args.accArgs.periods,
				tc.args.accArgs.startTime,
			)
			if tc.args.mintModAccountCoins {
				authBuilder = authBuilder.WithSimpleModuleAccount(kavadist.ModuleName, tc.args.period.Amount)
			}

			suite.SetupApp()
			suite.app.InitializeFromGenesisStates(
				authBuilder.BuildMarshalled(),
			)

			suite.ctx = suite.ctx.WithBlockTime(tc.args.ctxTime)

			err := suite.keeper.SendTimeLockedCoinsToPeriodicVestingAccount(suite.ctx, kavadist.ModuleName, suite.addrs[0], tc.args.period.Amount, tc.args.period.Length)

			if tc.errArgs.expectErr {
				suite.Require().Error(err)
				suite.Require().True(strings.Contains(err.Error(), tc.errArgs.contains))
			} else {
				suite.Require().NoError(err)

				acc := suite.getAccount(suite.addrs[0])
				vacc, ok := acc.(*vesting.PeriodicVestingAccount)
				suite.Require().True(ok)
				suite.Require().Equal(tc.args.expectedPeriods, vacc.VestingPeriods)
				suite.Require().Equal(tc.args.expectedStartTime, vacc.StartTime)
				suite.Require().Equal(tc.args.expectedEndTime, vacc.EndTime)
			}
		})
	}
}

func (suite *PayoutTestSuite) TestSendCoinsToBaseAccount() {
	authBuilder := NewAuthGenesisBuilder().
		WithSimpleAccount(suite.addrs[1], cs(c("ukava", 400))).
		WithSimpleModuleAccount(kavadist.ModuleName, cs(c("ukava", 600)))

	suite.SetupApp()

	suite.ctx = suite.app.NewContext(true, abci.Header{Height: 1, Time: time.Unix(100, 0)})

	suite.app.InitializeFromGenesisStates(
		authBuilder.BuildMarshalled(),
	)

	// send coins to base account
	err := suite.keeper.SendTimeLockedCoinsToAccount(suite.ctx, kavadist.ModuleName, suite.addrs[1], cs(c("ukava", 100)), 5)
	suite.Require().NoError(err)
	acc := suite.getAccount(suite.addrs[1])
	vacc, ok := acc.(*vesting.PeriodicVestingAccount)
	suite.True(ok)
	expectedPeriods := vesting.Periods{
		vesting.Period{Length: int64(5), Amount: cs(c("ukava", 100))},
	}
	suite.Equal(expectedPeriods, vacc.VestingPeriods)
	suite.Equal(cs(c("ukava", 100)), vacc.OriginalVesting)
	suite.Equal(cs(c("ukava", 500)), vacc.Coins)
	suite.Equal(int64(105), vacc.EndTime)
	suite.Equal(int64(100), vacc.StartTime)

}

func (suite *PayoutTestSuite) TestSendCoinsToInvalidAccount() {
	authBuilder := NewAuthGenesisBuilder().
		WithSimpleModuleAccount(kavadist.ModuleName, cs(c("ukava", 600))).
		WithEmptyValidatorVestingAccount(suite.addrs[2])

	suite.SetupApp()
	suite.app.InitializeFromGenesisStates(
		authBuilder.BuildMarshalled(),
	)
	err := suite.keeper.SendTimeLockedCoinsToAccount(suite.ctx, kavadist.ModuleName, suite.addrs[2], cs(c("ukava", 100)), 5)
	suite.Require().True(errors.Is(err, types.ErrInvalidAccountType))
	macc := suite.getModuleAccount(cdptypes.ModuleName)
	err = suite.keeper.SendTimeLockedCoinsToAccount(suite.ctx, kavadist.ModuleName, macc.GetAddress(), cs(c("ukava", 100)), 5)
	suite.Require().True(errors.Is(err, types.ErrInvalidAccountType))
}

func (suite *PayoutTestSuite) TestGetPeriodLength() {
	type args struct {
		blockTime      time.Time
		multiplier     types.Multiplier
		expectedLength int64
	}
	type errArgs struct {
		expectPass bool
		contains   string
	}
	type periodTest struct {
		name    string
		args    args
		errArgs errArgs
	}
	testCases := []periodTest{
		{
			name: "first half of month",
			args: args{
				blockTime:      time.Date(2020, 11, 2, 15, 0, 0, 0, time.UTC),
				multiplier:     types.NewMultiplier(types.Medium, 6, sdk.MustNewDecFromStr("0.333333")),
				expectedLength: time.Date(2021, 5, 15, 14, 0, 0, 0, time.UTC).Unix() - time.Date(2020, 11, 2, 15, 0, 0, 0, time.UTC).Unix(),
			},
			errArgs: errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			name: "first half of month long lockup",
			args: args{
				blockTime:      time.Date(2020, 11, 2, 15, 0, 0, 0, time.UTC),
				multiplier:     types.NewMultiplier(types.Medium, 24, sdk.MustNewDecFromStr("0.333333")),
				expectedLength: time.Date(2022, 11, 15, 14, 0, 0, 0, time.UTC).Unix() - time.Date(2020, 11, 2, 15, 0, 0, 0, time.UTC).Unix(),
			},
			errArgs: errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			name: "second half of month",
			args: args{
				blockTime:      time.Date(2020, 12, 31, 15, 0, 0, 0, time.UTC),
				multiplier:     types.NewMultiplier(types.Medium, 6, sdk.MustNewDecFromStr("0.333333")),
				expectedLength: time.Date(2021, 7, 1, 14, 0, 0, 0, time.UTC).Unix() - time.Date(2020, 12, 31, 15, 0, 0, 0, time.UTC).Unix(),
			},
			errArgs: errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			name: "second half of month long lockup",
			args: args{
				blockTime:      time.Date(2020, 12, 31, 15, 0, 0, 0, time.UTC),
				multiplier:     types.NewMultiplier(types.Large, 24, sdk.MustNewDecFromStr("0.333333")),
				expectedLength: time.Date(2023, 1, 1, 14, 0, 0, 0, time.UTC).Unix() - time.Date(2020, 12, 31, 15, 0, 0, 0, time.UTC).Unix(),
			},
			errArgs: errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			name: "end of feb",
			args: args{
				blockTime:      time.Date(2021, 2, 28, 15, 0, 0, 0, time.UTC),
				multiplier:     types.NewMultiplier(types.Medium, 6, sdk.MustNewDecFromStr("0.333333")),
				expectedLength: time.Date(2021, 9, 1, 14, 0, 0, 0, time.UTC).Unix() - time.Date(2021, 2, 28, 15, 0, 0, 0, time.UTC).Unix(),
			},
			errArgs: errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			name: "leap year",
			args: args{
				blockTime:      time.Date(2020, 2, 29, 15, 0, 0, 0, time.UTC),
				multiplier:     types.NewMultiplier(types.Medium, 6, sdk.MustNewDecFromStr("0.333333")),
				expectedLength: time.Date(2020, 9, 1, 14, 0, 0, 0, time.UTC).Unix() - time.Date(2020, 2, 29, 15, 0, 0, 0, time.UTC).Unix(),
			},
			errArgs: errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			name: "leap year long lockup",
			args: args{
				blockTime:      time.Date(2020, 2, 29, 15, 0, 0, 0, time.UTC),
				multiplier:     types.NewMultiplier(types.Large, 24, sdk.MustNewDecFromStr("1")),
				expectedLength: time.Date(2022, 3, 1, 14, 0, 0, 0, time.UTC).Unix() - time.Date(2020, 2, 29, 15, 0, 0, 0, time.UTC).Unix(),
			},
			errArgs: errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			name: "exactly half of month",
			args: args{
				blockTime:      time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				multiplier:     types.NewMultiplier(types.Medium, 6, sdk.MustNewDecFromStr("0.333333")),
				expectedLength: time.Date(2021, 7, 1, 14, 0, 0, 0, time.UTC).Unix() - time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC).Unix(),
			},
			errArgs: errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			name: "just before half of month",
			args: args{
				blockTime:      time.Date(2020, 12, 15, 13, 59, 59, 0, time.UTC),
				multiplier:     types.NewMultiplier(types.Medium, 6, sdk.MustNewDecFromStr("0.333333")),
				expectedLength: time.Date(2021, 6, 15, 14, 0, 0, 0, time.UTC).Unix() - time.Date(2020, 12, 15, 13, 59, 59, 0, time.UTC).Unix(),
			},
			errArgs: errArgs{
				expectPass: true,
				contains:   "",
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupApp()
			ctx := suite.ctx.WithBlockTime(tc.args.blockTime)
			length, err := suite.keeper.GetPeriodLength(ctx, tc.args.multiplier)
			if tc.errArgs.expectPass {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.args.expectedLength, length)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func TestPayoutTestSuite(t *testing.T) {
	suite.Run(t, new(PayoutTestSuite))
}
