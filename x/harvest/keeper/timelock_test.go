package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/harvest/types"
)

func (suite *KeeperTestSuite) TestSendTimeLockedCoinsToAccount() {
	type accountArgs struct {
		addr                 sdk.AccAddress
		vestingAccountBefore bool
		vestingAccountAfter  bool
		coins                sdk.Coins
		periods              vesting.Periods
		origVestingCoins     sdk.Coins
		startTime            int64
		endTime              int64
	}
	type args struct {
		accArgs                   accountArgs
		period                    vesting.Period
		blockTime                 time.Time
		expectedAccountBalance    sdk.Coins
		expectedModAccountBalance sdk.Coins
		expectedPeriods           vesting.Periods
		expectedStartTime         int64
		expectedEndTime           int64
	}
	type errArgs struct {
		expectPass bool
		contains   string
	}
	type testCase struct {
		name    string
		args    args
		errArgs errArgs
	}
	testCases := []testCase{
		{
			name: "send liquid coins to base account",
			args: args{
				accArgs: accountArgs{
					addr:                 sdk.AccAddress(crypto.AddressHash([]byte("test"))),
					vestingAccountBefore: false,
					vestingAccountAfter:  false,
					coins:                sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(1000)), sdk.NewCoin("btcb", sdk.NewInt(1000))),
				},
				period:                    vesting.Period{Length: 0, Amount: sdk.NewCoins(sdk.NewCoin("hard", sdk.NewInt(100)))},
				blockTime:                 time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC),
				expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(1000)), sdk.NewCoin("btcb", sdk.NewInt(1000)), sdk.NewCoin("hard", sdk.NewInt(100))),
				expectedModAccountBalance: sdk.NewCoins(sdk.NewCoin("hard", sdk.NewInt(900))),
				expectedPeriods:           vesting.Periods{},
				expectedStartTime:         0,
				expectedEndTime:           0,
			},
			errArgs: errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			name: "send liquid coins to vesting account",
			args: args{
				accArgs: accountArgs{
					addr:                 sdk.AccAddress(crypto.AddressHash([]byte("test"))),
					vestingAccountBefore: true,
					vestingAccountAfter:  true,
					coins:                sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(1000)), sdk.NewCoin("btcb", sdk.NewInt(1000))),
					periods: vesting.Periods{
						vesting.Period{Amount: sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(100))), Length: 100},
					},
					origVestingCoins: sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(100))),
					startTime:        time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC).Unix(),
					endTime:          time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC).Unix() + 100,
				},
				period:                    vesting.Period{Length: 0, Amount: sdk.NewCoins(sdk.NewCoin("hard", sdk.NewInt(100)))},
				blockTime:                 time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC),
				expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(1000)), sdk.NewCoin("btcb", sdk.NewInt(1000)), sdk.NewCoin("hard", sdk.NewInt(100))),
				expectedModAccountBalance: sdk.NewCoins(sdk.NewCoin("hard", sdk.NewInt(900))),
				expectedPeriods: vesting.Periods{
					vesting.Period{Amount: sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(100))), Length: 100},
				},
				expectedStartTime: time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC).Unix(),
				expectedEndTime:   time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC).Unix() + 100,
			},
			errArgs: errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			name: "insert period at beginning of schedule",
			args: args{
				accArgs: accountArgs{
					addr:                 sdk.AccAddress(crypto.AddressHash([]byte("test"))),
					vestingAccountBefore: true,
					vestingAccountAfter:  true,
					coins:                cs(c("bnb", 20)),
					periods: vesting.Periods{
						vesting.Period{Length: 5, Amount: cs(c("bnb", 5))},
						vesting.Period{Length: 5, Amount: cs(c("bnb", 5))},
						vesting.Period{Length: 5, Amount: cs(c("bnb", 5))},
						vesting.Period{Length: 5, Amount: cs(c("bnb", 5))}},
					origVestingCoins: cs(c("bnb", 20)),
					startTime:        100,
					endTime:          120,
				},
				period:                    vesting.Period{Length: 2, Amount: cs(c("hard", 6))},
				blockTime:                 time.Unix(101, 0),
				expectedAccountBalance:    cs(c("bnb", 20), c("hard", 6)),
				expectedModAccountBalance: cs(c("hard", 994)),
				expectedPeriods: vesting.Periods{
					vesting.Period{Length: 3, Amount: cs(c("hard", 6))},
					vesting.Period{Length: 2, Amount: cs(c("bnb", 5))},
					vesting.Period{Length: 5, Amount: cs(c("bnb", 5))},
					vesting.Period{Length: 5, Amount: cs(c("bnb", 5))},
					vesting.Period{Length: 5, Amount: cs(c("bnb", 5))}},
				expectedStartTime: 100,
				expectedEndTime:   120,
			},
			errArgs: errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			name: "insert period at beginning with new start time",
			args: args{
				accArgs: accountArgs{
					addr:                 sdk.AccAddress(crypto.AddressHash([]byte("test"))),
					vestingAccountBefore: true,
					vestingAccountAfter:  true,
					coins:                cs(c("bnb", 20)),
					periods: vesting.Periods{
						vesting.Period{Length: 5, Amount: cs(c("bnb", 5))},
						vesting.Period{Length: 5, Amount: cs(c("bnb", 5))},
						vesting.Period{Length: 5, Amount: cs(c("bnb", 5))},
						vesting.Period{Length: 5, Amount: cs(c("bnb", 5))}},
					origVestingCoins: cs(c("bnb", 20)),
					startTime:        100,
					endTime:          120,
				},
				period:                    vesting.Period{Length: 7, Amount: cs(c("hard", 6))},
				blockTime:                 time.Unix(80, 0),
				expectedAccountBalance:    cs(c("bnb", 20), c("hard", 6)),
				expectedModAccountBalance: cs(c("hard", 994)),
				expectedPeriods: vesting.Periods{
					vesting.Period{Length: 7, Amount: cs(c("hard", 6))},
					vesting.Period{Length: 18, Amount: cs(c("bnb", 5))},
					vesting.Period{Length: 5, Amount: cs(c("bnb", 5))},
					vesting.Period{Length: 5, Amount: cs(c("bnb", 5))},
					vesting.Period{Length: 5, Amount: cs(c("bnb", 5))}},
				expectedStartTime: 80,
				expectedEndTime:   120,
			},
			errArgs: errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			name: "insert period in middle of schedule",
			args: args{
				accArgs: accountArgs{
					addr:                 sdk.AccAddress(crypto.AddressHash([]byte("test"))),
					vestingAccountBefore: true,
					vestingAccountAfter:  true,
					coins:                cs(c("bnb", 20)),
					periods: vesting.Periods{
						vesting.Period{Length: 5, Amount: cs(c("bnb", 5))},
						vesting.Period{Length: 5, Amount: cs(c("bnb", 5))},
						vesting.Period{Length: 5, Amount: cs(c("bnb", 5))},
						vesting.Period{Length: 5, Amount: cs(c("bnb", 5))}},
					origVestingCoins: cs(c("bnb", 20)),
					startTime:        100,
					endTime:          120,
				},
				period:                    vesting.Period{Length: 7, Amount: cs(c("hard", 6))},
				blockTime:                 time.Unix(101, 0),
				expectedAccountBalance:    cs(c("bnb", 20), c("hard", 6)),
				expectedModAccountBalance: cs(c("hard", 994)),
				expectedPeriods: vesting.Periods{
					vesting.Period{Length: 5, Amount: cs(c("bnb", 5))},
					vesting.Period{Length: 3, Amount: cs(c("hard", 6))},
					vesting.Period{Length: 2, Amount: cs(c("bnb", 5))},
					vesting.Period{Length: 5, Amount: cs(c("bnb", 5))},
					vesting.Period{Length: 5, Amount: cs(c("bnb", 5))}},
				expectedStartTime: 100,
				expectedEndTime:   120,
			},
			errArgs: errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			name: "append to end of schedule",
			args: args{
				accArgs: accountArgs{
					addr:                 sdk.AccAddress(crypto.AddressHash([]byte("test"))),
					vestingAccountBefore: true,
					vestingAccountAfter:  true,
					coins:                cs(c("bnb", 20)),
					periods: vesting.Periods{
						vesting.Period{Length: 5, Amount: cs(c("bnb", 5))},
						vesting.Period{Length: 5, Amount: cs(c("bnb", 5))},
						vesting.Period{Length: 5, Amount: cs(c("bnb", 5))},
						vesting.Period{Length: 5, Amount: cs(c("bnb", 5))}},
					origVestingCoins: cs(c("bnb", 20)),
					startTime:        100,
					endTime:          120,
				},
				period:                    vesting.Period{Length: 7, Amount: cs(c("hard", 6))},
				blockTime:                 time.Unix(125, 0),
				expectedAccountBalance:    cs(c("bnb", 20), c("hard", 6)),
				expectedModAccountBalance: cs(c("hard", 994)),
				expectedPeriods: vesting.Periods{
					vesting.Period{Length: 5, Amount: cs(c("bnb", 5))},
					vesting.Period{Length: 5, Amount: cs(c("bnb", 5))},
					vesting.Period{Length: 5, Amount: cs(c("bnb", 5))},
					vesting.Period{Length: 5, Amount: cs(c("bnb", 5))},
					vesting.Period{Length: 12, Amount: cs(c("hard", 6))}},
				expectedStartTime: 100,
				expectedEndTime:   132,
			},
			errArgs: errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			name: "add coins to existing period",
			args: args{
				accArgs: accountArgs{
					addr:                 sdk.AccAddress(crypto.AddressHash([]byte("test"))),
					vestingAccountBefore: true,
					vestingAccountAfter:  true,
					coins:                cs(c("bnb", 20)),
					periods: vesting.Periods{
						vesting.Period{Length: 5, Amount: cs(c("bnb", 5))},
						vesting.Period{Length: 5, Amount: cs(c("bnb", 5))},
						vesting.Period{Length: 5, Amount: cs(c("bnb", 5))},
						vesting.Period{Length: 5, Amount: cs(c("bnb", 5))}},
					origVestingCoins: cs(c("bnb", 20)),
					startTime:        100,
					endTime:          120,
				},
				period:                    vesting.Period{Length: 5, Amount: cs(c("hard", 6))},
				blockTime:                 time.Unix(110, 0),
				expectedAccountBalance:    cs(c("bnb", 20), c("hard", 6)),
				expectedModAccountBalance: cs(c("hard", 994)),
				expectedPeriods: vesting.Periods{
					vesting.Period{Length: 5, Amount: cs(c("bnb", 5))},
					vesting.Period{Length: 5, Amount: cs(c("bnb", 5))},
					vesting.Period{Length: 5, Amount: cs(c("bnb", 5), c("hard", 6))},
					vesting.Period{Length: 5, Amount: cs(c("bnb", 5))}},
				expectedStartTime: 100,
				expectedEndTime:   120,
			},
			errArgs: errArgs{
				expectPass: true,
				contains:   "",
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// create new app with one funded account

			// Initialize test app and set context
			tApp := app.NewTestApp()
			ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tc.args.blockTime})
			authGS := app.NewAuthGenState([]sdk.AccAddress{tc.args.accArgs.addr}, []sdk.Coins{tc.args.accArgs.coins})
			harvestGS := types.NewGenesisState(types.NewParams(
				true,
				types.DistributionSchedules{
					types.NewDistributionSchedule(true, "bnb", time.Date(2020, 10, 8, 14, 0, 0, 0, time.UTC), time.Date(2020, 11, 22, 14, 0, 0, 0, time.UTC), sdk.NewCoin("hard", sdk.NewInt(5000)), time.Date(2021, 11, 22, 14, 0, 0, 0, time.UTC), types.Multipliers{types.NewMultiplier(types.Small, 0, sdk.MustNewDecFromStr("0.33")), types.NewMultiplier(types.Medium, 6, sdk.MustNewDecFromStr("0.5")), types.NewMultiplier(types.Large, 24, sdk.OneDec())}),
				},
				types.DelegatorDistributionSchedules{types.NewDelegatorDistributionSchedule(
					types.NewDistributionSchedule(true, "bnb", time.Date(2020, 10, 8, 14, 0, 0, 0, time.UTC), time.Date(2025, 10, 8, 14, 0, 0, 0, time.UTC), sdk.NewCoin("hard", sdk.NewInt(500)), time.Date(2026, 10, 8, 14, 0, 0, 0, time.UTC), types.Multipliers{types.NewMultiplier(types.Small, 0, sdk.MustNewDecFromStr("0.33")), types.NewMultiplier(types.Medium, 6, sdk.MustNewDecFromStr("0.5")), types.NewMultiplier(types.Large, 24, sdk.OneDec())}),
					time.Hour*24,
				),
				},
			), types.DefaultPreviousBlockTime, types.DefaultDistributionTimes, types.DefaultDeposits, types.DefaultClaims)
			tApp.InitializeFromGenesisStates(authGS, app.GenesisState{types.ModuleName: types.ModuleCdc.MustMarshalJSON(harvestGS)})
			if tc.args.accArgs.vestingAccountBefore {
				ak := tApp.GetAccountKeeper()
				acc := ak.GetAccount(ctx, tc.args.accArgs.addr)
				bacc := auth.NewBaseAccount(acc.GetAddress(), acc.GetCoins(), acc.GetPubKey(), acc.GetAccountNumber(), acc.GetSequence())
				bva, err := vesting.NewBaseVestingAccount(bacc, tc.args.accArgs.origVestingCoins, tc.args.accArgs.endTime)
				suite.Require().NoError(err)
				pva := vesting.NewPeriodicVestingAccountRaw(bva, tc.args.accArgs.startTime, tc.args.accArgs.periods)
				ak.SetAccount(ctx, pva)
			}
			supplyKeeper := tApp.GetSupplyKeeper()
			supplyKeeper.MintCoins(ctx, types.LPAccount, sdk.NewCoins(sdk.NewCoin("hard", sdk.NewInt(1000))))
			keeper := tApp.GetHarvestKeeper()
			suite.app = tApp
			suite.ctx = ctx
			suite.keeper = keeper

			err := suite.keeper.SendTimeLockedCoinsToAccount(suite.ctx, types.LPAccount, tc.args.accArgs.addr, tc.args.period.Amount, tc.args.period.Length)

			if tc.errArgs.expectPass {
				suite.Require().NoError(err)
				acc := suite.getAccount(tc.args.accArgs.addr)
				suite.Require().Equal(tc.args.expectedAccountBalance, acc.GetCoins())
				mAcc := suite.getModuleAccount(types.LPAccount)
				suite.Require().Equal(tc.args.expectedModAccountBalance, mAcc.GetCoins())
				vacc, ok := acc.(*vesting.PeriodicVestingAccount)
				if tc.args.accArgs.vestingAccountAfter {
					suite.Require().True(ok)
					suite.Require().Equal(tc.args.expectedPeriods, vacc.VestingPeriods)
					suite.Require().Equal(tc.args.expectedStartTime, vacc.StartTime)
					suite.Require().Equal(tc.args.expectedEndTime, vacc.EndTime)
				} else {
					suite.Require().False(ok)
				}
			}

		})
	}
}

func c(denom string, amount int64) sdk.Coin { return sdk.NewInt64Coin(denom, amount) }
func cs(coins ...sdk.Coin) sdk.Coins        { return sdk.NewCoins(coins...) }
