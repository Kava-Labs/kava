package keeper_test

import (
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/harvest/types"
	validatorvesting "github.com/kava-labs/kava/x/validator-vesting"
)

func (suite *KeeperTestSuite) TestClaim() {
	type args struct {
		claimOwner                sdk.AccAddress
		receiver                  sdk.AccAddress
		denom                     string
		claimType                 types.ClaimType
		multiplier                types.MultiplierName
		blockTime                 time.Time
		createClaim               bool
		claimAmount               sdk.Coin
		validatorVesting          bool
		expectedAccountBalance    sdk.Coins
		expectedModAccountBalance sdk.Coins
		expectedVestingAccount    bool
		expectedVestingLength     int64
	}
	type errArgs struct {
		expectPass bool
		contains   string
	}
	type claimTest struct {
		name    string
		args    args
		errArgs errArgs
	}
	testCases := []claimTest{
		{
			"valid liquid claim",
			args{
				claimOwner:                sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				receiver:                  sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				denom:                     "bnb",
				claimType:                 types.LP,
				blockTime:                 time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC),
				createClaim:               true,
				claimAmount:               sdk.NewCoin("hard", sdk.NewInt(100)),
				validatorVesting:          false,
				expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("hard", sdk.NewInt(33)), sdk.NewCoin("bnb", sdk.NewInt(1000)), sdk.NewCoin("btcb", sdk.NewInt(1000))),
				expectedModAccountBalance: sdk.NewCoins(sdk.NewCoin("hard", sdk.NewInt(967))),
				expectedVestingAccount:    false,
				expectedVestingLength:     0,
				multiplier:                types.Small,
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"valid liquid delegator claim",
			args{
				claimOwner:                sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				receiver:                  sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				denom:                     "bnb",
				claimType:                 types.Stake,
				blockTime:                 time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC),
				createClaim:               true,
				claimAmount:               sdk.NewCoin("hard", sdk.NewInt(100)),
				validatorVesting:          false,
				expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("hard", sdk.NewInt(33)), sdk.NewCoin("bnb", sdk.NewInt(1000)), sdk.NewCoin("btcb", sdk.NewInt(1000))),
				expectedModAccountBalance: sdk.NewCoins(sdk.NewCoin("hard", sdk.NewInt(967))),
				expectedVestingAccount:    false,
				expectedVestingLength:     0,
				multiplier:                types.Small,
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"valid medium vesting claim",
			args{
				claimOwner:                sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				receiver:                  sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				denom:                     "bnb",
				claimType:                 types.LP,
				blockTime:                 time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC),
				createClaim:               true,
				claimAmount:               sdk.NewCoin("hard", sdk.NewInt(100)),
				validatorVesting:          false,
				expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("hard", sdk.NewInt(50)), sdk.NewCoin("bnb", sdk.NewInt(1000)), sdk.NewCoin("btcb", sdk.NewInt(1000))),
				expectedModAccountBalance: sdk.NewCoins(sdk.NewCoin("hard", sdk.NewInt(950))),
				expectedVestingAccount:    true,
				expectedVestingLength:     16848000,
				multiplier:                types.Medium,
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"valid large vesting claim",
			args{
				claimOwner:                sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				receiver:                  sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				denom:                     "bnb",
				claimType:                 types.LP,
				blockTime:                 time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC),
				createClaim:               true,
				claimAmount:               sdk.NewCoin("hard", sdk.NewInt(100)),
				validatorVesting:          false,
				expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("hard", sdk.NewInt(100)), sdk.NewCoin("bnb", sdk.NewInt(1000)), sdk.NewCoin("btcb", sdk.NewInt(1000))),
				expectedModAccountBalance: sdk.NewCoins(sdk.NewCoin("hard", sdk.NewInt(900))),
				expectedVestingAccount:    true,
				expectedVestingLength:     64281600,
				multiplier:                types.Large,
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"valid validator vesting",
			args{
				claimOwner:                sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				receiver:                  sdk.AccAddress(crypto.AddressHash([]byte("test2"))),
				denom:                     "bnb",
				claimType:                 types.LP,
				blockTime:                 time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC),
				createClaim:               true,
				claimAmount:               sdk.NewCoin("hard", sdk.NewInt(100)),
				validatorVesting:          true,
				expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("hard", sdk.NewInt(100)), sdk.NewCoin("bnb", sdk.NewInt(1000)), sdk.NewCoin("btcb", sdk.NewInt(1000))),
				expectedModAccountBalance: sdk.NewCoins(sdk.NewCoin("hard", sdk.NewInt(900))),
				expectedVestingAccount:    true,
				expectedVestingLength:     64281600,
				multiplier:                types.Large,
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"invalid validator vesting",
			args{
				claimOwner:                sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				receiver:                  sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				denom:                     "bnb",
				claimType:                 types.LP,
				blockTime:                 time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC),
				createClaim:               true,
				claimAmount:               sdk.NewCoin("hard", sdk.NewInt(100)),
				validatorVesting:          true,
				expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("hard", sdk.NewInt(100)), sdk.NewCoin("bnb", sdk.NewInt(1000)), sdk.NewCoin("btcb", sdk.NewInt(1000))),
				expectedModAccountBalance: sdk.NewCoins(sdk.NewCoin("hard", sdk.NewInt(900))),
				expectedVestingAccount:    true,
				expectedVestingLength:     64281600,
				multiplier:                types.Large,
			},
			errArgs{
				expectPass: false,
				contains:   "receiver account type not supported",
			},
		},
		{
			"claim not found",
			args{
				claimOwner:                sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				receiver:                  sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				denom:                     "bnb",
				claimType:                 types.LP,
				blockTime:                 time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC),
				createClaim:               false,
				claimAmount:               sdk.NewCoin("hard", sdk.NewInt(100)),
				validatorVesting:          false,
				expectedAccountBalance:    sdk.Coins{},
				expectedModAccountBalance: sdk.Coins{},
				expectedVestingAccount:    false,
				expectedVestingLength:     0,
				multiplier:                types.Small,
			},
			errArgs{
				expectPass: false,
				contains:   "claim not found",
			},
		},
		{
			"claim expired",
			args{
				claimOwner:                sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				receiver:                  sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				denom:                     "bnb",
				claimType:                 types.LP,
				blockTime:                 time.Date(2022, 11, 1, 14, 0, 0, 0, time.UTC),
				createClaim:               true,
				claimAmount:               sdk.NewCoin("hard", sdk.NewInt(100)),
				validatorVesting:          false,
				expectedAccountBalance:    sdk.Coins{},
				expectedModAccountBalance: sdk.Coins{},
				expectedVestingAccount:    false,
				expectedVestingLength:     0,
				multiplier:                types.Small,
			},
			errArgs{
				expectPass: false,
				contains:   "claim period expired",
			},
		},
		{
			"different receiver address",
			args{
				claimOwner:                sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				receiver:                  sdk.AccAddress(crypto.AddressHash([]byte("test2"))),
				denom:                     "bnb",
				claimType:                 types.LP,
				blockTime:                 time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC),
				createClaim:               true,
				claimAmount:               sdk.NewCoin("hard", sdk.NewInt(100)),
				validatorVesting:          false,
				expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("hard", sdk.NewInt(100)), sdk.NewCoin("bnb", sdk.NewInt(1000)), sdk.NewCoin("btcb", sdk.NewInt(1000))),
				expectedModAccountBalance: sdk.NewCoins(sdk.NewCoin("hard", sdk.NewInt(900))),
				expectedVestingAccount:    true,
				expectedVestingLength:     64281600,
				multiplier:                types.Large,
			},
			errArgs{
				expectPass: false,
				contains:   "receiver account must match sender account",
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// create new app with one funded account

			// Initialize test app and set context
			tApp := app.NewTestApp()
			ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tc.args.blockTime})
			authGS := app.NewAuthGenState(
				[]sdk.AccAddress{tc.args.claimOwner, tc.args.receiver},
				[]sdk.Coins{
					sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(1000)), sdk.NewCoin("btcb", sdk.NewInt(1000))),
					sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(1000)), sdk.NewCoin("btcb", sdk.NewInt(1000))),
				})
			loanToValue := sdk.MustNewDecFromStr("0.6")
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
				types.MoneyMarkets{
					types.NewMoneyMarket("usdx", false, sdk.NewDec(1000000000000000), loanToValue, "usdx:usd", sdk.NewInt(1000000), types.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10"), sdk.MustNewDecFromStr("0.05"))),
					types.NewMoneyMarket("ukava", false, sdk.NewDec(1000000000000000), loanToValue, "kava:usd", sdk.NewInt(1000000), types.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10"), sdk.MustNewDecFromStr("0.05"))),
				},
			), types.DefaultPreviousBlockTime, types.DefaultDistributionTimes)
			tApp.InitializeFromGenesisStates(authGS, app.GenesisState{types.ModuleName: types.ModuleCdc.MustMarshalJSON(harvestGS)})
			if tc.args.validatorVesting {
				ak := tApp.GetAccountKeeper()
				acc := ak.GetAccount(ctx, tc.args.claimOwner)
				bacc := auth.NewBaseAccount(acc.GetAddress(), acc.GetCoins(), acc.GetPubKey(), acc.GetAccountNumber(), acc.GetSequence())
				bva, err := vesting.NewBaseVestingAccount(
					bacc,
					sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(20))), time.Date(2020, 10, 8, 14, 0, 0, 0, time.UTC).Unix()+100)
				suite.Require().NoError(err)
				vva := validatorvesting.NewValidatorVestingAccountRaw(
					bva,
					time.Date(2020, 10, 8, 14, 0, 0, 0, time.UTC).Unix(),
					vesting.Periods{
						vesting.Period{Length: 25, Amount: cs(c("bnb", 5))},
						vesting.Period{Length: 25, Amount: cs(c("bnb", 5))},
						vesting.Period{Length: 25, Amount: cs(c("bnb", 5))},
						vesting.Period{Length: 25, Amount: cs(c("bnb", 5))}},
					sdk.ConsAddress(crypto.AddressHash([]byte("test"))),
					sdk.AccAddress{},
					95,
				)
				err = vva.Validate()
				suite.Require().NoError(err)
				ak.SetAccount(ctx, vva)
			}
			supplyKeeper := tApp.GetSupplyKeeper()
			supplyKeeper.MintCoins(ctx, types.LPAccount, sdk.NewCoins(sdk.NewCoin("hard", sdk.NewInt(1000))))
			supplyKeeper.MintCoins(ctx, types.DelegatorAccount, sdk.NewCoins(sdk.NewCoin("hard", sdk.NewInt(1000))))
			keeper := tApp.GetHarvestKeeper()
			suite.app = tApp
			suite.ctx = ctx
			suite.keeper = keeper

			if tc.args.createClaim {
				claim := types.NewClaim(tc.args.claimOwner, tc.args.denom, tc.args.claimAmount, tc.args.claimType)
				suite.Require().NotPanics(func() { suite.keeper.SetClaim(suite.ctx, claim) })
			}

			err := suite.keeper.ClaimReward(suite.ctx, tc.args.claimOwner, tc.args.receiver, tc.args.denom, tc.args.claimType, tc.args.multiplier)
			if tc.errArgs.expectPass {
				suite.Require().NoError(err)
				acc := suite.getAccount(tc.args.receiver)
				suite.Require().Equal(tc.args.expectedAccountBalance, acc.GetCoins())
				mAcc := suite.getModuleAccount(types.LPAccount)
				if tc.args.claimType == types.Stake {
					mAcc = suite.getModuleAccount(types.DelegatorAccount)
				}
				suite.Require().Equal(tc.args.expectedModAccountBalance, mAcc.GetCoins())
				vacc, ok := acc.(*vesting.PeriodicVestingAccount)
				if tc.args.expectedVestingAccount {
					suite.Require().True(ok)
					suite.Require().Equal(tc.args.expectedVestingLength, vacc.VestingPeriods[0].Length)
				} else {
					suite.Require().False(ok)
				}
				_, f := suite.keeper.GetClaim(ctx, tc.args.claimOwner, tc.args.denom, tc.args.claimType)
				suite.Require().False(f)
			} else {
				suite.Require().Error(err)
				suite.Require().True(strings.Contains(err.Error(), tc.errArgs.contains))
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGetPeriodLength() {
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
