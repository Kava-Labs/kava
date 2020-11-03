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

func (suite *KeeperTestSuite) TestBorrow() {
	type args struct {
		borrower                  sdk.AccAddress
		depositCoin               sdk.Coin
		coins                     sdk.Coins
		maxLoanToValue            string
		expectedAccountBalance    sdk.Coins
		expectedModAccountBalance sdk.Coins
	}
	type errArgs struct {
		expectPass bool
		contains   string
	}
	type borrowTest struct {
		name    string
		args    args
		errArgs errArgs
	}
	testCases := []borrowTest{
		{
			"valid",
			args{
				borrower:                  sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				depositCoin:               sdk.NewCoin("ukava", sdk.NewInt(100)),
				coins:                     sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(50))),
				maxLoanToValue:            "0.6",
				expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(150))),
				expectedModAccountBalance: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(950))),
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"loan-to-value limited",
			args{
				borrower:                  sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				depositCoin:               sdk.NewCoin("ukava", sdk.NewInt(20)),               // 20 KAVA x $5.00 price = $100
				coins:                     sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(61))), // 61 USDX x $1 price = $61
				maxLoanToValue:            "0.6",
				expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(150))),
				expectedModAccountBalance: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(950))),
			},
			errArgs{
				expectPass: false,
				contains:   "total deposited value is insufficient for borrow request",
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Initialize test app and set context
			tApp := app.NewTestApp()
			ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
			authGS := app.NewAuthGenState(
				[]sdk.AccAddress{tc.args.borrower},
				[]sdk.Coins{sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100)))})
			loanToValue := sdk.MustNewDecFromStr(tc.args.maxLoanToValue)
			harvestGS := types.NewGenesisState(types.NewParams(
				true,
				types.DistributionSchedules{
					types.NewDistributionSchedule(true, "usdx", time.Date(2020, 10, 8, 14, 0, 0, 0, time.UTC), time.Date(2020, 11, 22, 14, 0, 0, 0, time.UTC), sdk.NewCoin("hard", sdk.NewInt(5000)), time.Date(2021, 11, 22, 14, 0, 0, 0, time.UTC), types.Multipliers{types.NewMultiplier(types.Small, 0, sdk.MustNewDecFromStr("0.33")), types.NewMultiplier(types.Medium, 6, sdk.MustNewDecFromStr("0.5")), types.NewMultiplier(types.Medium, 24, sdk.OneDec())}),
					types.NewDistributionSchedule(true, "ukava", time.Date(2020, 10, 8, 14, 0, 0, 0, time.UTC), time.Date(2020, 11, 22, 14, 0, 0, 0, time.UTC), sdk.NewCoin("hard", sdk.NewInt(5000)), time.Date(2021, 11, 22, 14, 0, 0, 0, time.UTC), types.Multipliers{types.NewMultiplier(types.Small, 0, sdk.MustNewDecFromStr("0.33")), types.NewMultiplier(types.Medium, 6, sdk.MustNewDecFromStr("0.5")), types.NewMultiplier(types.Medium, 24, sdk.OneDec())}),
				},
				types.DelegatorDistributionSchedules{types.NewDelegatorDistributionSchedule(
					types.NewDistributionSchedule(true, "usdx", time.Date(2020, 10, 8, 14, 0, 0, 0, time.UTC), time.Date(2025, 10, 8, 14, 0, 0, 0, time.UTC), sdk.NewCoin("hard", sdk.NewInt(500)), time.Date(2026, 10, 8, 14, 0, 0, 0, time.UTC), types.Multipliers{types.NewMultiplier(types.Small, 0, sdk.MustNewDecFromStr("0.33")), types.NewMultiplier(types.Medium, 6, sdk.MustNewDecFromStr("0.5")), types.NewMultiplier(types.Medium, 24, sdk.OneDec())}),
					time.Hour*24,
				),
				},
				types.MoneyMarkets{
					types.NewMoneyMarket("usdx", sdk.NewInt(1000000000000000), loanToValue, "usdx:usd", sdk.NewInt(1000000)),
					types.NewMoneyMarket("ukava", sdk.NewInt(1000000000000000), loanToValue, "kava:usd", sdk.NewInt(1000000)),
				},
			), types.DefaultPreviousBlockTime, types.DefaultDistributionTimes)
			tApp.InitializeFromGenesisStates(authGS, NewPricefeedGenStateMulti(),
				app.GenesisState{types.ModuleName: types.ModuleCdc.MustMarshalJSON(harvestGS)})
			keeper := tApp.GetHarvestKeeper()
			supplyKeeper := tApp.GetSupplyKeeper()
			supplyKeeper.MintCoins(ctx, types.ModuleAccountName, sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1000))))
			suite.app = tApp
			suite.ctx = ctx
			suite.keeper = keeper

			var err error

			// deposit some coins
			err = suite.keeper.Deposit(suite.ctx, tc.args.borrower, tc.args.depositCoin, types.LP)
			suite.Require().NoError(err)

			// run the test
			err = suite.keeper.Borrow(suite.ctx, tc.args.borrower, tc.args.coins)

			// verify results
			if tc.errArgs.expectPass {
				suite.Require().NoError(err)
				acc := suite.getAccount(tc.args.borrower)
				suite.Require().Equal(tc.args.expectedAccountBalance.Sub(sdk.NewCoins(tc.args.depositCoin)), acc.GetCoins())
				mAcc := suite.getModuleAccount(types.ModuleAccountName)
				suite.Require().Equal(tc.args.expectedModAccountBalance.Add(tc.args.depositCoin), mAcc.GetCoins())
				_, f := suite.keeper.GetBorrow(suite.ctx, tc.args.borrower)
				suite.Require().True(f)
			} else {
				suite.Require().Error(err)
				suite.Require().True(strings.Contains(err.Error(), tc.errArgs.contains))
			}
		})
	}
}
