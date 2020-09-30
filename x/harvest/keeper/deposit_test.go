package keeper_test

import (
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/harvest/types"
)

func (suite *KeeperTestSuite) TestDeposit() {
	type args struct {
		depositor                 sdk.AccAddress
		amount                    sdk.Coin
		depositType               types.DepositType
		numberDeposits            int
		expectedAccountBalance    sdk.Coins
		expectedModAccountBalance sdk.Coins
	}
	type errArgs struct {
		expectPass bool
		contains   string
	}
	type depositTest struct {
		name    string
		args    args
		errArgs errArgs
	}
	testCases := []depositTest{
		{
			"valid",
			args{
				depositor:                 sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				amount:                    sdk.NewCoin("bnb", sdk.NewInt(100)),
				depositType:               types.LP,
				numberDeposits:            1,
				expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(900)), sdk.NewCoin("btcb", sdk.NewInt(1000))),
				expectedModAccountBalance: sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(100))),
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"valid multi deposit",
			args{
				depositor:                 sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				amount:                    sdk.NewCoin("bnb", sdk.NewInt(100)),
				depositType:               types.LP,
				numberDeposits:            2,
				expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(800)), sdk.NewCoin("btcb", sdk.NewInt(1000))),
				expectedModAccountBalance: sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(200))),
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"invalid deposit type",
			args{
				depositor:                 sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				amount:                    sdk.NewCoin("bnb", sdk.NewInt(100)),
				depositType:               types.Stake,
				numberDeposits:            1,
				expectedAccountBalance:    sdk.Coins{},
				expectedModAccountBalance: sdk.Coins{},
			},
			errArgs{
				expectPass: false,
				contains:   "invalid deposit type",
			},
		},
		{
			"invalid deposit denom",
			args{
				depositor:                 sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				amount:                    sdk.NewCoin("btcb", sdk.NewInt(100)),
				depositType:               types.LP,
				numberDeposits:            1,
				expectedAccountBalance:    sdk.Coins{},
				expectedModAccountBalance: sdk.Coins{},
			},
			errArgs{
				expectPass: false,
				contains:   "invalid deposit denom",
			},
		},
		{
			"insufficient funds",
			args{
				depositor:                 sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				amount:                    sdk.NewCoin("bnb", sdk.NewInt(10000)),
				depositType:               types.LP,
				numberDeposits:            1,
				expectedAccountBalance:    sdk.Coins{},
				expectedModAccountBalance: sdk.Coins{},
			},
			errArgs{
				expectPass: false,
				contains:   "insufficient funds",
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// create new app with one funded account
			config := sdk.GetConfig()
			app.SetBech32AddressPrefixes(config)
			// Initialize test app and set context
			tApp := app.NewTestApp()
			ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
			authGS := app.NewAuthGenState([]sdk.AccAddress{tc.args.depositor}, []sdk.Coins{sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(1000)), sdk.NewCoin("btcb", sdk.NewInt(1000)))})
			harvestGS := types.NewGenesisState(types.NewParams(
				true,
				types.DistributionSchedules{
					types.NewDistributionSchedule(true, "bnb", time.Date(2020, 10, 8, 14, 0, 0, 0, time.UTC), time.Date(2020, 11, 22, 14, 0, 0, 0, time.UTC), sdk.NewCoin("hard", sdk.NewInt(5000)), time.Date(2021, 11, 22, 14, 0, 0, 0, time.UTC), types.Multipliers{types.NewMultiplier(types.Small, 0, sdk.MustNewDecFromStr("0.33")), types.NewMultiplier(types.Medium, 6, sdk.MustNewDecFromStr("0.5")), types.NewMultiplier(types.Medium, 24, sdk.OneDec())}),
				},
				types.DelegatorDistributionSchedules{types.NewDelegatorDistributionSchedule(
					types.NewDistributionSchedule(true, "bnb", time.Date(2020, 10, 8, 14, 0, 0, 0, time.UTC), time.Date(2025, 10, 8, 14, 0, 0, 0, time.UTC), sdk.NewCoin("hard", sdk.NewInt(500)), time.Date(2026, 10, 8, 14, 0, 0, 0, time.UTC), types.Multipliers{types.NewMultiplier(types.Small, 0, sdk.MustNewDecFromStr("0.33")), types.NewMultiplier(types.Medium, 6, sdk.MustNewDecFromStr("0.5")), types.NewMultiplier(types.Medium, 24, sdk.OneDec())}),
					time.Hour*24,
				),
				},
			), types.DefaultPreviousBlockTime, types.DefaultDistributionTimes)
			tApp.InitializeFromGenesisStates(authGS, app.GenesisState{types.ModuleName: types.ModuleCdc.MustMarshalJSON(harvestGS)})
			keeper := tApp.GetHarvestKeeper()
			suite.app = tApp
			suite.ctx = ctx
			suite.keeper = keeper

			// run the test
			var err error
			for i := 0; i < tc.args.numberDeposits; i++ {
				err = suite.keeper.Deposit(suite.ctx, tc.args.depositor, tc.args.amount, tc.args.depositType)
			}

			// verify results
			if tc.errArgs.expectPass {
				suite.Require().NoError(err)
				acc := suite.getAccount(tc.args.depositor)
				suite.Require().Equal(tc.args.expectedAccountBalance, acc.GetCoins())
				mAcc := suite.getModuleAccount(types.ModuleAccountName)
				suite.Require().Equal(tc.args.expectedModAccountBalance, mAcc.GetCoins())
				_, f := suite.keeper.GetDeposit(suite.ctx, tc.args.depositor, tc.args.amount.Denom, tc.args.depositType)
				suite.Require().True(f)
			} else {
				suite.Require().Error(err)
				suite.Require().True(strings.Contains(err.Error(), tc.errArgs.contains))
			}
		})
	}
}

func (suite *KeeperTestSuite) TestWithdraw() {
	type args struct {
		depositor                 sdk.AccAddress
		depositAmount             sdk.Coin
		withdrawAmount            sdk.Coin
		depositType               types.DepositType
		withdrawType              types.DepositType
		createDeposit             bool
		expectedAccountBalance    sdk.Coins
		expectedModAccountBalance sdk.Coins
		depositExists             bool
		finalDepositAmount        sdk.Coin
	}
	type errArgs struct {
		expectPass bool
		contains   string
	}
	type withdrawTest struct {
		name    string
		args    args
		errArgs errArgs
	}
	testCases := []withdrawTest{
		{
			"valid partial withdraw",
			args{
				depositor:                 sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				depositAmount:             sdk.NewCoin("bnb", sdk.NewInt(200)),
				withdrawAmount:            sdk.NewCoin("bnb", sdk.NewInt(100)),
				depositType:               types.LP,
				withdrawType:              types.LP,
				createDeposit:             true,
				expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(900)), sdk.NewCoin("btcb", sdk.NewInt(1000))),
				expectedModAccountBalance: sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(100))),
				depositExists:             true,
				finalDepositAmount:        sdk.NewCoin("bnb", sdk.NewInt(100)),
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"valid full withdraw",
			args{
				depositor:                 sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				depositAmount:             sdk.NewCoin("bnb", sdk.NewInt(200)),
				withdrawAmount:            sdk.NewCoin("bnb", sdk.NewInt(200)),
				depositType:               types.LP,
				withdrawType:              types.LP,
				createDeposit:             true,
				expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(1000)), sdk.NewCoin("btcb", sdk.NewInt(1000))),
				expectedModAccountBalance: sdk.Coins(nil),
				depositExists:             false,
				finalDepositAmount:        sdk.Coin{},
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"deposit not found invalid denom",
			args{
				depositor:                 sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				depositAmount:             sdk.NewCoin("bnb", sdk.NewInt(200)),
				withdrawAmount:            sdk.NewCoin("btcb", sdk.NewInt(200)),
				depositType:               types.LP,
				withdrawType:              types.LP,
				createDeposit:             true,
				expectedAccountBalance:    sdk.Coins{},
				expectedModAccountBalance: sdk.Coins{},
				depositExists:             false,
				finalDepositAmount:        sdk.Coin{},
			},
			errArgs{
				expectPass: false,
				contains:   "deposit not found",
			},
		},
		{
			"deposit not found invalid deposit type",
			args{
				depositor:                 sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				depositAmount:             sdk.NewCoin("bnb", sdk.NewInt(200)),
				withdrawAmount:            sdk.NewCoin("bnb", sdk.NewInt(200)),
				depositType:               types.LP,
				withdrawType:              types.Stake,
				createDeposit:             true,
				expectedAccountBalance:    sdk.Coins{},
				expectedModAccountBalance: sdk.Coins{},
				depositExists:             false,
				finalDepositAmount:        sdk.Coin{},
			},
			errArgs{
				expectPass: false,
				contains:   "deposit not found",
			},
		},
		{
			"withdraw exceeds deposit",
			args{
				depositor:                 sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				depositAmount:             sdk.NewCoin("bnb", sdk.NewInt(200)),
				withdrawAmount:            sdk.NewCoin("bnb", sdk.NewInt(300)),
				depositType:               types.LP,
				withdrawType:              types.LP,
				createDeposit:             true,
				expectedAccountBalance:    sdk.Coins{},
				expectedModAccountBalance: sdk.Coins{},
				depositExists:             false,
				finalDepositAmount:        sdk.Coin{},
			},
			errArgs{
				expectPass: false,
				contains:   "withdrawal amount exceeds deposit amount",
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// create new app with one funded account
			config := sdk.GetConfig()
			app.SetBech32AddressPrefixes(config)
			// Initialize test app and set context
			tApp := app.NewTestApp()
			ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
			authGS := app.NewAuthGenState([]sdk.AccAddress{tc.args.depositor}, []sdk.Coins{sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(1000)), sdk.NewCoin("btcb", sdk.NewInt(1000)))})
			harvestGS := types.NewGenesisState(types.NewParams(
				true,
				types.DistributionSchedules{
					types.NewDistributionSchedule(true, "bnb", time.Date(2020, 10, 8, 14, 0, 0, 0, time.UTC), time.Date(2020, 11, 22, 14, 0, 0, 0, time.UTC), sdk.NewCoin("hard", sdk.NewInt(5000)), time.Date(2021, 11, 22, 14, 0, 0, 0, time.UTC), types.Multipliers{types.NewMultiplier(types.Small, 0, sdk.MustNewDecFromStr("0.33")), types.NewMultiplier(types.Medium, 6, sdk.MustNewDecFromStr("0.5")), types.NewMultiplier(types.Medium, 24, sdk.OneDec())}),
				},
				types.DelegatorDistributionSchedules{types.NewDelegatorDistributionSchedule(
					types.NewDistributionSchedule(true, "bnb", time.Date(2020, 10, 8, 14, 0, 0, 0, time.UTC), time.Date(2025, 10, 8, 14, 0, 0, 0, time.UTC), sdk.NewCoin("hard", sdk.NewInt(500)), time.Date(2026, 10, 8, 14, 0, 0, 0, time.UTC), types.Multipliers{types.NewMultiplier(types.Small, 0, sdk.MustNewDecFromStr("0.33")), types.NewMultiplier(types.Medium, 6, sdk.MustNewDecFromStr("0.5")), types.NewMultiplier(types.Medium, 24, sdk.OneDec())}),
					time.Hour*24,
				),
				},
			), types.DefaultPreviousBlockTime, types.DefaultDistributionTimes)
			tApp.InitializeFromGenesisStates(authGS, app.GenesisState{types.ModuleName: types.ModuleCdc.MustMarshalJSON(harvestGS)})
			keeper := tApp.GetHarvestKeeper()
			suite.app = tApp
			suite.ctx = ctx
			suite.keeper = keeper

			if tc.args.createDeposit {
				err := suite.keeper.Deposit(suite.ctx, tc.args.depositor, tc.args.depositAmount, tc.args.depositType)
				suite.Require().NoError(err)
			}

			err := suite.keeper.Withdraw(suite.ctx, tc.args.depositor, tc.args.withdrawAmount, tc.args.withdrawType)

			if tc.errArgs.expectPass {
				suite.Require().NoError(err)
				acc := suite.getAccount(tc.args.depositor)
				suite.Require().Equal(tc.args.expectedAccountBalance, acc.GetCoins())
				mAcc := suite.getModuleAccount(types.ModuleAccountName)
				suite.Require().Equal(tc.args.expectedModAccountBalance, mAcc.GetCoins())
				testDeposit, f := suite.keeper.GetDeposit(suite.ctx, tc.args.depositor, tc.args.depositAmount.Denom, tc.args.depositType)
				if tc.args.depositExists {
					suite.Require().True(f)
					suite.Require().Equal(tc.args.finalDepositAmount, testDeposit.Amount)
				} else {
					suite.Require().False(f)
				}
			} else {
				suite.Require().Error(err)
				suite.Require().True(strings.Contains(err.Error(), tc.errArgs.contains))
			}
		})

	}
}
