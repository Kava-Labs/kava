package keeper_test

import (
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/hard"
	"github.com/kava-labs/kava/x/hard/types"
	"github.com/kava-labs/kava/x/pricefeed"
)

func (suite *KeeperTestSuite) TestDeposit() {
	type args struct {
		depositor                 sdk.AccAddress
		amount                    sdk.Coins
		numberDeposits            int
		expectedAccountBalance    sdk.Coins
		expectedModAccountBalance sdk.Coins
		expectedDepositCoins      sdk.Coins
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
				amount:                    sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(100))),
				numberDeposits:            1,
				expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(900)), sdk.NewCoin("btcb", sdk.NewInt(1000))),
				expectedModAccountBalance: sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(100))),
				expectedDepositCoins:      sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(100))),
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
				amount:                    sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(100))),
				numberDeposits:            2,
				expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(800)), sdk.NewCoin("btcb", sdk.NewInt(1000))),
				expectedModAccountBalance: sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(200))),
				expectedDepositCoins:      sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(200))),
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"invalid deposit denom",
			args{
				depositor:                 sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				amount:                    sdk.NewCoins(sdk.NewCoin("fake", sdk.NewInt(100))),
				numberDeposits:            1,
				expectedAccountBalance:    sdk.Coins{},
				expectedModAccountBalance: sdk.Coins{},
				expectedDepositCoins:      sdk.Coins{},
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
				amount:                    sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(10000))),
				numberDeposits:            1,
				expectedAccountBalance:    sdk.Coins{},
				expectedModAccountBalance: sdk.Coins{},
				expectedDepositCoins:      sdk.Coins{},
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

			// Initialize test app and set context
			tApp := app.NewTestApp()
			ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
			authGS := app.NewAuthGenState([]sdk.AccAddress{tc.args.depositor}, []sdk.Coins{sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(1000)), sdk.NewCoin("btcb", sdk.NewInt(1000)))})
			loanToValue, _ := sdk.NewDecFromStr("0.6")
			hardGS := types.NewGenesisState(types.NewParams(
				types.MoneyMarkets{
					types.NewMoneyMarket("usdx", types.NewBorrowLimit(false, sdk.NewDec(1000000000000000), loanToValue), "usdx:usd", sdk.NewInt(1000000), sdk.NewInt(USDX_CF*1000), types.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec()),
					types.NewMoneyMarket("ukava", types.NewBorrowLimit(false, sdk.NewDec(1000000000000000), loanToValue), "kava:usd", sdk.NewInt(1000000), sdk.NewInt(KAVA_CF*1000), types.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec()),
					types.NewMoneyMarket("bnb", types.NewBorrowLimit(false, sdk.NewDec(1000000000000000), loanToValue), "bnb:usd", sdk.NewInt(1000000), sdk.NewInt(BNB_CF*1000), types.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec()),
					types.NewMoneyMarket("btcb", types.NewBorrowLimit(false, sdk.NewDec(1000000000000000), loanToValue), "btcb:usd", sdk.NewInt(1000000), sdk.NewInt(BTCB_CF*1000), types.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec()),
				},
			), types.DefaultAccumulationTimes, types.DefaultDeposits, types.DefaultBorrows,
				types.DefaultTotalSupplied, types.DefaultTotalBorrowed, types.DefaultTotalReserves,
			)

			// Pricefeed module genesis state
			pricefeedGS := pricefeed.GenesisState{
				Params: pricefeed.Params{
					Markets: []pricefeed.Market{
						{MarketID: "usdx:usd", BaseAsset: "usdx", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
						{MarketID: "kava:usd", BaseAsset: "kava", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
						{MarketID: "btcb:usd", BaseAsset: "btcb", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
						{MarketID: "bnb:usd", BaseAsset: "bnb", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
					},
				},
				PostedPrices: []pricefeed.PostedPrice{
					{
						MarketID:      "usdx:usd",
						OracleAddress: sdk.AccAddress{},
						Price:         sdk.MustNewDecFromStr("1.00"),
						Expiry:        time.Now().Add(1 * time.Hour),
					},
					{
						MarketID:      "kava:usd",
						OracleAddress: sdk.AccAddress{},
						Price:         sdk.MustNewDecFromStr("2.00"),
						Expiry:        time.Now().Add(1 * time.Hour),
					},
					{
						MarketID:      "btcb:usd",
						OracleAddress: sdk.AccAddress{},
						Price:         sdk.MustNewDecFromStr("100.00"),
						Expiry:        time.Now().Add(1 * time.Hour),
					},
					{
						MarketID:      "bnb:usd",
						OracleAddress: sdk.AccAddress{},
						Price:         sdk.MustNewDecFromStr("10.00"),
						Expiry:        time.Now().Add(1 * time.Hour),
					},
				},
			}

			tApp.InitializeFromGenesisStates(authGS,
				app.GenesisState{pricefeed.ModuleName: pricefeed.ModuleCdc.MustMarshalJSON(pricefeedGS)},
				app.GenesisState{types.ModuleName: types.ModuleCdc.MustMarshalJSON(hardGS)},
			)
			keeper := tApp.GetHardKeeper()
			suite.app = tApp
			suite.ctx = ctx
			suite.keeper = keeper

			// Run BeginBlocker once to transition MoneyMarkets
			hard.BeginBlocker(suite.ctx, suite.keeper)

			// run the test
			var err error
			for i := 0; i < tc.args.numberDeposits; i++ {
				err = suite.keeper.Deposit(suite.ctx, tc.args.depositor, tc.args.amount)
			}

			// verify results
			if tc.errArgs.expectPass {
				suite.Require().NoError(err)
				acc := suite.getAccount(tc.args.depositor)
				suite.Require().Equal(tc.args.expectedAccountBalance, acc.GetCoins())
				mAcc := suite.getModuleAccount(types.ModuleAccountName)
				suite.Require().Equal(tc.args.expectedModAccountBalance, mAcc.GetCoins())
				dep, f := suite.keeper.GetDeposit(suite.ctx, tc.args.depositor)
				suite.Require().True(f)
				suite.Require().Equal(tc.args.expectedDepositCoins, dep.Amount)
			} else {
				suite.Require().Error(err)
				suite.Require().True(strings.Contains(err.Error(), tc.errArgs.contains))
			}
		})
	}
}
