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

func (suite *KeeperTestSuite) TestWithdraw() {
	type args struct {
		depositor                 sdk.AccAddress
		initialModAccountBalance  sdk.Coins
		depositAmount             sdk.Coins
		withdrawAmount            sdk.Coins
		createDeposit             bool
		expectedAccountBalance    sdk.Coins
		expectedModAccountBalance sdk.Coins
		finalDepositAmount        sdk.Coins
	}
	type errArgs struct {
		expectPass   bool
		expectDelete bool
		contains     string
	}
	type withdrawTest struct {
		name    string
		args    args
		errArgs errArgs
	}
	testCases := []withdrawTest{
		{
			"valid: partial withdraw",
			args{
				depositor:                 sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				initialModAccountBalance:  sdk.Coins(nil),
				depositAmount:             sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(200))),
				withdrawAmount:            sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(100))),
				createDeposit:             true,
				expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(900)), sdk.NewCoin("btcb", sdk.NewInt(1000))),
				expectedModAccountBalance: sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(100))),
				finalDepositAmount:        sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(100))),
			},
			errArgs{
				expectPass:   true,
				expectDelete: false,
				contains:     "",
			},
		},
		{
			"valid: full withdraw",
			args{
				depositor:                 sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				initialModAccountBalance:  sdk.Coins(nil),
				depositAmount:             sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(200))),
				withdrawAmount:            sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(200))),
				createDeposit:             true,
				expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(1000)), sdk.NewCoin("btcb", sdk.NewInt(1000))),
				expectedModAccountBalance: sdk.Coins(nil),
				finalDepositAmount:        sdk.Coins{},
			},
			errArgs{
				expectPass:   true,
				expectDelete: true,
				contains:     "",
			},
		},
		{
			"valid: withdraw exceeds deposit but is adjusted to match max deposit",
			args{
				depositor:                 sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				initialModAccountBalance:  sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(1000))),
				depositAmount:             sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(200))),
				withdrawAmount:            sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(300))),
				createDeposit:             true,
				expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(1000)), sdk.NewCoin("btcb", sdk.NewInt(1000))),
				expectedModAccountBalance: sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(1000))),
				finalDepositAmount:        sdk.Coins{},
			},
			errArgs{
				expectPass:   true,
				expectDelete: true,
				contains:     "",
			},
		},
		{
			"invalid: withdraw non-supplied coin type",
			args{
				depositor:                 sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				initialModAccountBalance:  sdk.Coins(nil),
				depositAmount:             sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(200))),
				withdrawAmount:            sdk.NewCoins(sdk.NewCoin("btcb", sdk.NewInt(200))),
				createDeposit:             true,
				expectedAccountBalance:    sdk.Coins{},
				expectedModAccountBalance: sdk.Coins{},
				finalDepositAmount:        sdk.Coins{},
			},
			errArgs{
				expectPass:   false,
				expectDelete: false,
				contains:     "no coins of this type deposited",
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// create new app with one funded account

			// Initialize test app and set context
			tApp := app.NewTestApp()
			ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
			authGS := app.NewAuthGenState(
				[]sdk.AccAddress{tc.args.depositor},
				[]sdk.Coins{sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(1000)), sdk.NewCoin("btcb", sdk.NewInt(1000)))},
			)

			loanToValue := sdk.MustNewDecFromStr("0.6")
			hardGS := types.NewGenesisState(types.NewParams(
				types.MoneyMarkets{
					types.NewMoneyMarket("usdx", types.NewBorrowLimit(false, sdk.NewDec(1000000000000000), loanToValue), "usdx:usd", sdk.NewInt(1000000), types.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec()),
					types.NewMoneyMarket("ukava", types.NewBorrowLimit(false, sdk.NewDec(1000000000000000), loanToValue), "kava:usd", sdk.NewInt(1000000), types.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec()),
					types.NewMoneyMarket("bnb", types.NewBorrowLimit(false, sdk.NewDec(1000000000000000), loanToValue), "bnb:usd", sdk.NewInt(100000000), types.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec()),
				},
				sdk.NewDec(10),
			), types.DefaultAccumulationTimes, types.DefaultDeposits, types.DefaultBorrows,
				types.DefaultTotalSupplied, types.DefaultTotalBorrowed, types.DefaultTotalReserves,
			)

			// Pricefeed module genesis state
			pricefeedGS := pricefeed.GenesisState{
				Params: pricefeed.Params{
					Markets: []pricefeed.Market{
						{MarketID: "usdx:usd", BaseAsset: "usdx", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
						{MarketID: "kava:usd", BaseAsset: "kava", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
						{MarketID: "bnb:usd", BaseAsset: "bnb", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
					},
				},
				PostedPrices: []pricefeed.PostedPrice{
					{
						MarketID:      "usdx:usd",
						OracleAddress: sdk.AccAddress{},
						Price:         sdk.MustNewDecFromStr("1.00"),
						Expiry:        time.Now().Add(100 * time.Hour),
					},
					{
						MarketID:      "kava:usd",
						OracleAddress: sdk.AccAddress{},
						Price:         sdk.MustNewDecFromStr("2.00"),
						Expiry:        time.Now().Add(100 * time.Hour),
					},
					{
						MarketID:      "bnb:usd",
						OracleAddress: sdk.AccAddress{},
						Price:         sdk.MustNewDecFromStr("10.00"),
						Expiry:        time.Now().Add(100 * time.Hour),
					},
				},
			}

			tApp.InitializeFromGenesisStates(authGS,
				app.GenesisState{pricefeed.ModuleName: pricefeed.ModuleCdc.MustMarshalJSON(pricefeedGS)},
				app.GenesisState{types.ModuleName: types.ModuleCdc.MustMarshalJSON(hardGS)})
			keeper := tApp.GetHardKeeper()
			suite.app = tApp
			suite.ctx = ctx
			suite.keeper = keeper

			// Mint coins to Hard module account
			supplyKeeper := tApp.GetSupplyKeeper()
			supplyKeeper.MintCoins(ctx, types.ModuleAccountName, tc.args.initialModAccountBalance)

			if tc.args.createDeposit {
				err := suite.keeper.Deposit(suite.ctx, tc.args.depositor, tc.args.depositAmount)
				suite.Require().NoError(err)
			}

			err := suite.keeper.Withdraw(suite.ctx, tc.args.depositor, tc.args.withdrawAmount)

			if tc.errArgs.expectPass {
				suite.Require().NoError(err)
				acc := suite.getAccount(tc.args.depositor)
				suite.Require().Equal(tc.args.expectedAccountBalance, acc.GetCoins())
				mAcc := suite.getModuleAccount(types.ModuleAccountName)
				suite.Require().Equal(tc.args.expectedModAccountBalance, mAcc.GetCoins())
				testDeposit, f := suite.keeper.GetDeposit(suite.ctx, tc.args.depositor)
				if tc.errArgs.expectDelete {
					suite.Require().False(f)
				} else {
					suite.Require().True(f)
					suite.Require().Equal(tc.args.finalDepositAmount, testDeposit.Amount)
				}
			} else {
				suite.Require().Error(err)
				suite.Require().True(strings.Contains(err.Error(), tc.errArgs.contains))
			}
		})

	}
}

func (suite *KeeperTestSuite) TestLtvWithdraw() {
	type args struct {
		borrower             sdk.AccAddress
		initialModuleCoins   sdk.Coins
		initialBorrowerCoins sdk.Coins
		depositCoins         sdk.Coins
		borrowCoins          sdk.Coins
		futureTime           int64
	}

	type errArgs struct {
		expectPass bool
		contains   string
	}

	type liqTest struct {
		name    string
		args    args
		errArgs errArgs
	}

	// Set up test constants
	model := types.NewInterestRateModel(sdk.MustNewDecFromStr("0"), sdk.MustNewDecFromStr("0.1"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("0.5"))
	reserveFactor := sdk.MustNewDecFromStr("0.05")
	oneMonthInSeconds := int64(2592000)
	borrower := sdk.AccAddress(crypto.AddressHash([]byte("testborrower")))

	testCases := []liqTest{
		{
			"invalid: withdraw is outside loan-to-value range",
			args{
				borrower:             borrower,
				initialModuleCoins:   sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))),
				initialBorrowerCoins: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF)), sdk.NewCoin("usdx", sdk.NewInt(100*KAVA_CF))),
				depositCoins:         sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(10*KAVA_CF))), // 10 * 2 = $20
				borrowCoins:          sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(8*KAVA_CF))),  // 8 * 2 = $16
				futureTime:           oneMonthInSeconds,
			},
			errArgs{
				expectPass: false,
				contains:   "proposed withdraw outside loan-to-value range",
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Initialize test app and set context
			tApp := app.NewTestApp()
			ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})

			// Auth module genesis state
			authGS := app.NewAuthGenState(
				[]sdk.AccAddress{tc.args.borrower},
				[]sdk.Coins{tc.args.initialBorrowerCoins},
			)

			// Harvest module genesis state
			harvestGS := types.NewGenesisState(types.NewParams(
				types.MoneyMarkets{
					types.NewMoneyMarket("ukava",
						types.NewBorrowLimit(false, sdk.NewDec(100000000*KAVA_CF), sdk.MustNewDecFromStr("0.8")), // Borrow Limit
						"kava:usd",                     // Market ID
						sdk.NewInt(KAVA_CF),            // Conversion Factor
						model,                          // Interest Rate Model
						reserveFactor,                  // Reserve Factor
						sdk.MustNewDecFromStr("0.05")), // Keeper Reward Percent
					types.NewMoneyMarket("usdx",
						types.NewBorrowLimit(false, sdk.NewDec(100000000*KAVA_CF), sdk.MustNewDecFromStr("0.8")), // Borrow Limit
						"usdx:usd",                     // Market ID
						sdk.NewInt(KAVA_CF),            // Conversion Factor
						model,                          // Interest Rate Model
						reserveFactor,                  // Reserve Factor
						sdk.MustNewDecFromStr("0.05")), // Keeper Reward Percent
				},
				sdk.NewDec(10),
			), types.DefaultAccumulationTimes, types.DefaultDeposits, types.DefaultBorrows,
				types.DefaultTotalSupplied, types.DefaultTotalBorrowed, types.DefaultTotalReserves,
			)

			// Pricefeed module genesis state
			pricefeedGS := pricefeed.GenesisState{
				Params: pricefeed.Params{
					Markets: []pricefeed.Market{
						{MarketID: "usdx:usd", BaseAsset: "usdx", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
						{MarketID: "kava:usd", BaseAsset: "kava", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
					},
				},
				PostedPrices: []pricefeed.PostedPrice{
					{
						MarketID:      "usdx:usd",
						OracleAddress: sdk.AccAddress{},
						Price:         sdk.MustNewDecFromStr("1.00"),
						Expiry:        time.Now().Add(100 * time.Hour),
					},
					{
						MarketID:      "kava:usd",
						OracleAddress: sdk.AccAddress{},
						Price:         sdk.MustNewDecFromStr("2.00"),
						Expiry:        time.Now().Add(100 * time.Hour),
					},
				},
			}

			// Initialize test application
			tApp.InitializeFromGenesisStates(authGS,
				app.GenesisState{pricefeed.ModuleName: pricefeed.ModuleCdc.MustMarshalJSON(pricefeedGS)},
				app.GenesisState{types.ModuleName: types.ModuleCdc.MustMarshalJSON(harvestGS)})

			// Mint coins to Harvest module account
			supplyKeeper := tApp.GetSupplyKeeper()
			supplyKeeper.MintCoins(ctx, types.ModuleAccountName, tc.args.initialModuleCoins)

			auctionKeeper := tApp.GetAuctionKeeper()

			keeper := tApp.GetHardKeeper()
			suite.app = tApp
			suite.ctx = ctx
			suite.keeper = keeper
			suite.auctionKeeper = auctionKeeper

			var err error

			// Run begin blocker to set up state
			hard.BeginBlocker(suite.ctx, suite.keeper)

			// Borrower deposits coins
			err = suite.keeper.Deposit(suite.ctx, tc.args.borrower, tc.args.depositCoins)
			suite.Require().NoError(err)

			// Borrower borrows coins
			err = suite.keeper.Borrow(suite.ctx, tc.args.borrower, tc.args.borrowCoins)
			suite.Require().NoError(err)

			// Attempting to withdraw fails
			err = suite.keeper.Withdraw(suite.ctx, tc.args.borrower, sdk.NewCoins(sdk.NewCoin("ukava", sdk.OneInt())))
			suite.Require().Error(err)
			suite.Require().True(strings.Contains(err.Error(), tc.errArgs.contains))

			// Set up future chain context and run begin blocker, increasing user's owed borrow balance
			runAtTime := time.Unix(suite.ctx.BlockTime().Unix()+(tc.args.futureTime), 0)
			liqCtx := suite.ctx.WithBlockTime(runAtTime)
			hard.BeginBlocker(liqCtx, suite.keeper)

			// Attempted withdraw of 1 coin still fails
			err = suite.keeper.Withdraw(suite.ctx, tc.args.borrower, sdk.NewCoins(sdk.NewCoin("ukava", sdk.OneInt())))
			suite.Require().Error(err)
			suite.Require().True(strings.Contains(err.Error(), tc.errArgs.contains))

			// Repay the initial principal
			err = suite.keeper.Repay(suite.ctx, tc.args.borrower, tc.args.borrower, tc.args.borrowCoins)
			suite.Require().NoError(err)

			// Attempted withdraw of all deposited coins fails as user hasn't repaid interest debt
			err = suite.keeper.Withdraw(suite.ctx, tc.args.borrower, tc.args.depositCoins)
			suite.Require().Error(err)
			suite.Require().True(strings.Contains(err.Error(), tc.errArgs.contains))

			// Withdrawing half the coins should succeed
			withdrawCoins := sdk.NewCoins(sdk.NewCoin("ukava", tc.args.depositCoins[0].Amount.Quo(sdk.NewInt(2))))
			err = suite.keeper.Withdraw(suite.ctx, tc.args.borrower, withdrawCoins)
			suite.Require().NoError(err)
		})
	}
}
