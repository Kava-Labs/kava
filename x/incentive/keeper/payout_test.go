package keeper_test

import (
	"errors"
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/kava-labs/kava/app"
	cdptypes "github.com/kava-labs/kava/x/cdp/types"
	"github.com/kava-labs/kava/x/incentive/types"
	"github.com/kava-labs/kava/x/kavadist"
	validatorvesting "github.com/kava-labs/kava/x/validator-vesting"
)

func (suite *KeeperTestSuite) TestPayoutClaim() {
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
				expectedPeriods:          vesting.Periods{vesting.Period{Length: 31536000, Amount: cs(c("ukava", 10571385600))}},
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
			suite.SetupWithGenState()
			suite.ctx = suite.ctx.WithBlockTime(tc.args.initialTime)

			// setup incentive state
			params := types.NewParams(
				true,
				types.RewardPeriods{types.NewRewardPeriod(true, tc.args.ctype, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), tc.args.rewardsPerSecond)},
				types.RewardPeriods{types.NewRewardPeriod(true, tc.args.ctype, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), tc.args.rewardsPerSecond)},
				types.RewardPeriods{types.NewRewardPeriod(true, tc.args.ctype, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), tc.args.rewardsPerSecond)},
				types.RewardPeriods{types.NewRewardPeriod(true, tc.args.ctype, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), tc.args.rewardsPerSecond)},
				tc.args.multipliers,
				tc.args.initialTime.Add(time.Hour*24*365*5),
			)
			suite.keeper.SetParams(suite.ctx, params)
			suite.keeper.SetPreviousUSDXMintingAccrualTime(suite.ctx, tc.args.ctype, tc.args.initialTime)
			suite.keeper.SetUSDXMintingRewardFactor(suite.ctx, tc.args.ctype, sdk.ZeroDec())

			// setup account state
			sk := suite.app.GetSupplyKeeper()
			err := sk.MintCoins(suite.ctx, cdptypes.ModuleName, sdk.NewCoins(tc.args.initialCollateral))
			suite.Require().NoError(err)
			err = sk.SendCoinsFromModuleToAccount(suite.ctx, cdptypes.ModuleName, suite.addrs[0], sdk.NewCoins(tc.args.initialCollateral))
			suite.Require().NoError(err)

			// setup kavadist state
			err = sk.MintCoins(suite.ctx, kavadist.ModuleName, cs(c("ukava", 1000000000000)))
			suite.Require().NoError(err)

			// setup cdp state
			cdpKeeper := suite.app.GetCDPKeeper()
			err = cdpKeeper.AddCdp(suite.ctx, suite.addrs[0], tc.args.initialCollateral, tc.args.initialPrincipal, tc.args.ctype)
			suite.Require().NoError(err)

			claim, found := suite.keeper.GetUSDXMintingClaim(suite.ctx, suite.addrs[0])
			suite.Require().True(found)
			suite.Require().Equal(sdk.ZeroDec(), claim.RewardIndexes[0].RewardFactor)

			updatedBlockTime := suite.ctx.BlockTime().Add(time.Duration(int(time.Second) * tc.args.timeElapsed))
			suite.ctx = suite.ctx.WithBlockTime(updatedBlockTime)
			rewardPeriod, found := suite.keeper.GetUSDXMintingRewardPeriod(suite.ctx, tc.args.ctype)
			suite.Require().True(found)
			err = suite.keeper.AccumulateUSDXMintingRewards(suite.ctx, rewardPeriod)
			suite.Require().NoError(err)

			err = suite.keeper.ClaimReward(suite.ctx, suite.addrs[0], tc.args.multiplier)

			if tc.errArgs.expectPass {
				suite.Require().NoError(err)
				ak := suite.app.GetAccountKeeper()
				acc := ak.GetAccount(suite.ctx, suite.addrs[0])
				suite.Require().Equal(tc.args.expectedBalance, acc.GetCoins())

				if tc.args.isPeriodicVestingAccount {
					vacc, ok := acc.(*vesting.PeriodicVestingAccount)
					suite.Require().True(ok)
					suite.Require().Equal(tc.args.expectedPeriods, vacc.VestingPeriods)
				}

				claim, found := suite.keeper.GetUSDXMintingClaim(suite.ctx, suite.addrs[0])
				fmt.Println(claim)
				suite.Require().True(found)
				suite.Require().Equal(c("ukava", 0), claim.Reward)
			} else {
				suite.Require().Error(err)
				suite.Require().True(strings.Contains(err.Error(), tc.errArgs.contains))
			}
		})
	}
}

func (suite *KeeperTestSuite) TestSendCoinsToPeriodicVestingAccount() {
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
			// create the periodic vesting account
			pva, err := createPeriodicVestingAccount(tc.args.accArgs.origVestingCoins, tc.args.accArgs.periods, tc.args.accArgs.startTime, tc.args.accArgs.endTime)
			suite.Require().NoError(err)

			// setup store state with account and kavadist module account
			suite.ctx = suite.ctx.WithBlockTime(tc.args.ctxTime)
			ak := suite.app.GetAccountKeeper()
			ak.SetAccount(suite.ctx, pva)
			// mint module account coins if required
			if tc.args.mintModAccountCoins {
				sk := suite.app.GetSupplyKeeper()
				err = sk.MintCoins(suite.ctx, kavadist.ModuleName, tc.args.period.Amount)
				suite.Require().NoError(err)
			}

			err = suite.keeper.SendTimeLockedCoinsToPeriodicVestingAccount(suite.ctx, kavadist.ModuleName, pva.Address, tc.args.period.Amount, tc.args.period.Length)
			if tc.errArgs.expectErr {
				suite.Require().Error(err)
				suite.Require().True(strings.Contains(err.Error(), tc.errArgs.contains))
			} else {
				suite.Require().NoError(err)

				acc := suite.getAccount(pva.Address)
				vacc, ok := acc.(*vesting.PeriodicVestingAccount)
				suite.Require().True(ok)
				suite.Require().Equal(tc.args.expectedPeriods, vacc.VestingPeriods)
				suite.Require().Equal(tc.args.expectedStartTime, vacc.StartTime)
				suite.Require().Equal(tc.args.expectedEndTime, vacc.EndTime)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestSendCoinsToBaseAccount() {
	suite.SetupWithAccountState()
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

func (suite *KeeperTestSuite) TestSendCoinsToInvalidAccount() {
	suite.SetupWithAccountState()
	err := suite.keeper.SendTimeLockedCoinsToAccount(suite.ctx, kavadist.ModuleName, suite.addrs[2], cs(c("ukava", 100)), 5)
	suite.Require().True(errors.Is(err, types.ErrInvalidAccountType))
	macc := suite.getModuleAccount(cdptypes.ModuleName)
	err = suite.keeper.SendTimeLockedCoinsToAccount(suite.ctx, kavadist.ModuleName, macc.GetAddress(), cs(c("ukava", 100)), 5)
	suite.Require().True(errors.Is(err, types.ErrInvalidAccountType))
}

func (suite *KeeperTestSuite) SetupWithAccountState() {
	// creates a new app state with 4 funded addresses and 1 module account
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: time.Unix(100, 0)})
	_, addrs := app.GeneratePrivKeyAddressPairs(4)
	authGS := app.NewAuthGenState(
		addrs,
		[]sdk.Coins{
			cs(c("ukava", 400)),
			cs(c("ukava", 400)),
			cs(c("ukava", 400)),
			cs(c("ukava", 400)),
		})
	tApp.InitializeFromGenesisStates(
		authGS,
	)
	supplyKeeper := tApp.GetSupplyKeeper()
	macc := supplyKeeper.GetModuleAccount(ctx, kavadist.ModuleName)
	err := supplyKeeper.MintCoins(ctx, macc.GetName(), cs(c("ukava", 600)))
	suite.Require().NoError(err)

	// sets addrs[0] to be a periodic vesting account
	ak := tApp.GetAccountKeeper()
	acc := ak.GetAccount(ctx, addrs[0])
	bacc := auth.NewBaseAccount(acc.GetAddress(), acc.GetCoins(), acc.GetPubKey(), acc.GetAccountNumber(), acc.GetSequence())
	periods := vesting.Periods{
		vesting.Period{Length: int64(1), Amount: cs(c("ukava", 100))},
		vesting.Period{Length: int64(2), Amount: cs(c("ukava", 100))},
		vesting.Period{Length: int64(8), Amount: cs(c("ukava", 100))},
		vesting.Period{Length: int64(5), Amount: cs(c("ukava", 100))},
	}
	bva, err2 := vesting.NewBaseVestingAccount(bacc, cs(c("ukava", 400)), ctx.BlockTime().Unix()+16)
	suite.Require().NoError(err2)
	pva := vesting.NewPeriodicVestingAccountRaw(bva, ctx.BlockTime().Unix(), periods)
	ak.SetAccount(ctx, pva)

	// sets addrs[2] to be a validator vesting account
	acc = ak.GetAccount(ctx, addrs[2])
	bacc = auth.NewBaseAccount(acc.GetAddress(), acc.GetCoins(), acc.GetPubKey(), acc.GetAccountNumber(), acc.GetSequence())
	bva, err2 = vesting.NewBaseVestingAccount(bacc, cs(c("ukava", 400)), ctx.BlockTime().Unix()+16)
	suite.Require().NoError(err2)
	vva := validatorvesting.NewValidatorVestingAccountRaw(bva, ctx.BlockTime().Unix(), periods, sdk.ConsAddress{}, nil, 90)
	ak.SetAccount(ctx, vva)
	suite.app = tApp
	suite.keeper = tApp.GetIncentiveKeeper()
	suite.ctx = ctx
	suite.addrs = addrs
}
