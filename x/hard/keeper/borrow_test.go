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

const (
	USDX_CF = 1000000
	KAVA_CF = 1000000
	BTCB_CF = 100000000
	BNB_CF  = 100000000
	BUSD_CF = 100000000
)

func (suite *KeeperTestSuite) TestBorrow() {

	type args struct {
		usdxBorrowLimit           sdk.Dec
		priceKAVA                 sdk.Dec
		loanToValueKAVA           sdk.Dec
		priceBTCB                 sdk.Dec
		loanToValueBTCB           sdk.Dec
		priceBNB                  sdk.Dec
		loanToValueBNB            sdk.Dec
		borrower                  sdk.AccAddress
		depositCoins              []sdk.Coin
		previousBorrowCoins       sdk.Coins
		borrowCoins               sdk.Coins
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
				usdxBorrowLimit:           sdk.MustNewDecFromStr("100000000000"),
				priceKAVA:                 sdk.MustNewDecFromStr("5.00"),
				loanToValueKAVA:           sdk.MustNewDecFromStr("0.6"),
				priceBTCB:                 sdk.MustNewDecFromStr("0.00"),
				loanToValueBTCB:           sdk.MustNewDecFromStr("0.01"),
				priceBNB:                  sdk.MustNewDecFromStr("0.00"),
				loanToValueBNB:            sdk.MustNewDecFromStr("0.01"),
				borrower:                  sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				depositCoins:              []sdk.Coin{sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))},
				previousBorrowCoins:       sdk.NewCoins(),
				borrowCoins:               sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(20*KAVA_CF))),
				expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(20*KAVA_CF)), sdk.NewCoin("btcb", sdk.NewInt(100*BTCB_CF)), sdk.NewCoin("bnb", sdk.NewInt(100*BNB_CF)), sdk.NewCoin("xyz", sdk.NewInt(1))),
				expectedModAccountBalance: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1080*KAVA_CF)), sdk.NewCoin("usdx", sdk.NewInt(200*USDX_CF)), sdk.NewCoin("busd", sdk.NewInt(100*BUSD_CF))),
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"invalid: loan-to-value limited",
			args{
				usdxBorrowLimit:           sdk.MustNewDecFromStr("100000000000"),
				priceKAVA:                 sdk.MustNewDecFromStr("5.00"),
				loanToValueKAVA:           sdk.MustNewDecFromStr("0.6"),
				priceBTCB:                 sdk.MustNewDecFromStr("0.00"),
				loanToValueBTCB:           sdk.MustNewDecFromStr("0.01"),
				priceBNB:                  sdk.MustNewDecFromStr("0.00"),
				loanToValueBNB:            sdk.MustNewDecFromStr("0.01"),
				borrower:                  sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				depositCoins:              []sdk.Coin{sdk.NewCoin("ukava", sdk.NewInt(20*KAVA_CF))},  // 20 KAVA x $5.00 price = $100
				borrowCoins:               sdk.NewCoins(sdk.NewCoin("usdx", sdk.NewInt(61*USDX_CF))), // 61 USDX x $1 price = $61
				expectedAccountBalance:    sdk.NewCoins(),
				expectedModAccountBalance: sdk.NewCoins(),
			},
			errArgs{
				expectPass: false,
				contains:   "exceeds the allowable amount as determined by the collateralization ratio",
			},
		},
		{
			"valid: multiple deposits",
			args{
				usdxBorrowLimit:           sdk.MustNewDecFromStr("100000000000"),
				priceKAVA:                 sdk.MustNewDecFromStr("2.00"),
				loanToValueKAVA:           sdk.MustNewDecFromStr("0.80"),
				priceBTCB:                 sdk.MustNewDecFromStr("10000.00"),
				loanToValueBTCB:           sdk.MustNewDecFromStr("0.10"),
				priceBNB:                  sdk.MustNewDecFromStr("0.00"),
				loanToValueBNB:            sdk.MustNewDecFromStr("0.01"),
				borrower:                  sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				depositCoins:              sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(50*KAVA_CF)), sdk.NewCoin("btcb", sdk.NewInt(0.1*BTCB_CF))),
				borrowCoins:               sdk.NewCoins(sdk.NewCoin("usdx", sdk.NewInt(180*USDX_CF))),
				expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(50*KAVA_CF)), sdk.NewCoin("btcb", sdk.NewInt(99.9*BTCB_CF)), sdk.NewCoin("usdx", sdk.NewInt(180*USDX_CF)), sdk.NewCoin("bnb", sdk.NewInt(100*BNB_CF)), sdk.NewCoin("xyz", sdk.NewInt(1))),
				expectedModAccountBalance: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1050*KAVA_CF)), sdk.NewCoin("usdx", sdk.NewInt(20*USDX_CF)), sdk.NewCoin("btcb", sdk.NewInt(0.1*BTCB_CF)), sdk.NewCoin("busd", sdk.NewInt(100*BUSD_CF))),
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"invalid: multiple deposits",
			args{
				usdxBorrowLimit:           sdk.MustNewDecFromStr("100000000000"),
				priceKAVA:                 sdk.MustNewDecFromStr("2.00"),
				loanToValueKAVA:           sdk.MustNewDecFromStr("0.80"),
				priceBTCB:                 sdk.MustNewDecFromStr("10000.00"),
				loanToValueBTCB:           sdk.MustNewDecFromStr("0.10"),
				priceBNB:                  sdk.MustNewDecFromStr("0.00"),
				loanToValueBNB:            sdk.MustNewDecFromStr("0.01"),
				borrower:                  sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				depositCoins:              sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(50*KAVA_CF)), sdk.NewCoin("btcb", sdk.NewInt(0.1*BTCB_CF))),
				borrowCoins:               sdk.NewCoins(sdk.NewCoin("usdx", sdk.NewInt(181*USDX_CF))),
				expectedAccountBalance:    sdk.NewCoins(),
				expectedModAccountBalance: sdk.NewCoins(),
			},
			errArgs{
				expectPass: false,
				contains:   "exceeds the allowable amount as determined by the collateralization ratio",
			},
		},
		{
			"valid: multiple previous borrows",
			args{
				usdxBorrowLimit:           sdk.MustNewDecFromStr("100000000000"),
				priceKAVA:                 sdk.MustNewDecFromStr("2.00"),
				loanToValueKAVA:           sdk.MustNewDecFromStr("0.8"),
				priceBTCB:                 sdk.MustNewDecFromStr("0.00"),
				loanToValueBTCB:           sdk.MustNewDecFromStr("0.01"),
				priceBNB:                  sdk.MustNewDecFromStr("5.00"),
				loanToValueBNB:            sdk.MustNewDecFromStr("0.8"),
				borrower:                  sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				depositCoins:              sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(30*BNB_CF)), sdk.NewCoin("ukava", sdk.NewInt(50*KAVA_CF))), // (50 KAVA x $2.00 price = $100) + (30 BNB x $5.00 price = $150) = $250
				previousBorrowCoins:       sdk.NewCoins(sdk.NewCoin("usdx", sdk.NewInt(99*USDX_CF)), sdk.NewCoin("busd", sdk.NewInt(100*BUSD_CF))),
				borrowCoins:               sdk.NewCoins(sdk.NewCoin("usdx", sdk.NewInt(1*USDX_CF))),
				expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(50*KAVA_CF)), sdk.NewCoin("btcb", sdk.NewInt(100*BTCB_CF)), sdk.NewCoin("usdx", sdk.NewInt(100*USDX_CF)), sdk.NewCoin("busd", sdk.NewInt(100*BUSD_CF)), sdk.NewCoin("bnb", sdk.NewInt(70*BNB_CF)), sdk.NewCoin("xyz", sdk.NewInt(1))),
				expectedModAccountBalance: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1050*KAVA_CF)), sdk.NewCoin("bnb", sdk.NewInt(30*BUSD_CF)), sdk.NewCoin("usdx", sdk.NewInt(100*USDX_CF))),
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"invalid: over loan-to-value with multiple previous borrows",
			args{
				usdxBorrowLimit:           sdk.MustNewDecFromStr("100000000000"),
				priceKAVA:                 sdk.MustNewDecFromStr("2.00"),
				loanToValueKAVA:           sdk.MustNewDecFromStr("0.8"),
				priceBTCB:                 sdk.MustNewDecFromStr("0.00"),
				loanToValueBTCB:           sdk.MustNewDecFromStr("0.01"),
				priceBNB:                  sdk.MustNewDecFromStr("5.00"),
				loanToValueBNB:            sdk.MustNewDecFromStr("0.8"),
				borrower:                  sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				depositCoins:              sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(30*BNB_CF)), sdk.NewCoin("ukava", sdk.NewInt(50*KAVA_CF))), // (50 KAVA x $2.00 price = $100) + (30 BNB x $5.00 price = $150) = $250
				previousBorrowCoins:       sdk.NewCoins(sdk.NewCoin("usdx", sdk.NewInt(100*USDX_CF)), sdk.NewCoin("busd", sdk.NewInt(100*BUSD_CF))),
				borrowCoins:               sdk.NewCoins(sdk.NewCoin("usdx", sdk.NewInt(1*USDX_CF))),
				expectedAccountBalance:    sdk.NewCoins(),
				expectedModAccountBalance: sdk.NewCoins(),
			},
			errArgs{
				expectPass: false,
				contains:   "exceeds the allowable amount as determined by the collateralization ratio",
			},
		},
		{
			"invalid: no price for asset",
			args{
				usdxBorrowLimit:           sdk.MustNewDecFromStr("100000000000"),
				priceKAVA:                 sdk.MustNewDecFromStr("5.00"),
				loanToValueKAVA:           sdk.MustNewDecFromStr("0.6"),
				priceBTCB:                 sdk.MustNewDecFromStr("0.00"),
				loanToValueBTCB:           sdk.MustNewDecFromStr("0.01"),
				priceBNB:                  sdk.MustNewDecFromStr("0.00"),
				loanToValueBNB:            sdk.MustNewDecFromStr("0.01"),
				borrower:                  sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				depositCoins:              sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))),
				previousBorrowCoins:       sdk.NewCoins(),
				borrowCoins:               sdk.NewCoins(sdk.NewCoin("xyz", sdk.NewInt(1))),
				expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(20*KAVA_CF)), sdk.NewCoin("btcb", sdk.NewInt(100*BTCB_CF)), sdk.NewCoin("bnb", sdk.NewInt(100*BNB_CF)), sdk.NewCoin("xyz", sdk.NewInt(1))),
				expectedModAccountBalance: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1080*KAVA_CF)), sdk.NewCoin("usdx", sdk.NewInt(200*USDX_CF)), sdk.NewCoin("busd", sdk.NewInt(100*BUSD_CF))),
			},
			errArgs{
				expectPass: false,
				contains:   "no price found for market",
			},
		},
		{
			"invalid: borrow exceed module account balance",
			args{
				usdxBorrowLimit:           sdk.MustNewDecFromStr("100000000000"),
				priceKAVA:                 sdk.MustNewDecFromStr("2.00"),
				loanToValueKAVA:           sdk.MustNewDecFromStr("0.8"),
				priceBTCB:                 sdk.MustNewDecFromStr("0.00"),
				loanToValueBTCB:           sdk.MustNewDecFromStr("0.01"),
				priceBNB:                  sdk.MustNewDecFromStr("0.00"),
				loanToValueBNB:            sdk.MustNewDecFromStr("0.01"),
				borrower:                  sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				depositCoins:              sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))),
				previousBorrowCoins:       sdk.NewCoins(),
				borrowCoins:               sdk.NewCoins(sdk.NewCoin("busd", sdk.NewInt(101*BUSD_CF))),
				expectedAccountBalance:    sdk.NewCoins(),
				expectedModAccountBalance: sdk.NewCoins(),
			},
			errArgs{
				expectPass: false,
				contains:   "exceeds module account balance:",
			},
		},
		{
			"invalid: over global asset borrow limit",
			args{
				usdxBorrowLimit:           sdk.MustNewDecFromStr("20000000"),
				priceKAVA:                 sdk.MustNewDecFromStr("2.00"),
				loanToValueKAVA:           sdk.MustNewDecFromStr("0.8"),
				priceBTCB:                 sdk.MustNewDecFromStr("0.00"),
				loanToValueBTCB:           sdk.MustNewDecFromStr("0.01"),
				priceBNB:                  sdk.MustNewDecFromStr("0.00"),
				loanToValueBNB:            sdk.MustNewDecFromStr("0.01"),
				borrower:                  sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				depositCoins:              sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(50*KAVA_CF))),
				previousBorrowCoins:       sdk.NewCoins(),
				borrowCoins:               sdk.NewCoins(sdk.NewCoin("usdx", sdk.NewInt(25*USDX_CF))),
				expectedAccountBalance:    sdk.NewCoins(),
				expectedModAccountBalance: sdk.NewCoins(),
			},
			errArgs{
				expectPass: false,
				contains:   "fails global asset borrow limit validation",
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
				[]sdk.Coins{sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF)),
					sdk.NewCoin("btcb", sdk.NewInt(100*BTCB_CF)), sdk.NewCoin("bnb", sdk.NewInt(100*BNB_CF)),
					sdk.NewCoin("xyz", sdk.NewInt(1)))})

			// hard module genesis state
			hardGS := types.NewGenesisState(types.NewParams(
				types.MoneyMarkets{
					types.NewMoneyMarket("usdx", types.NewBorrowLimit(true, tc.args.usdxBorrowLimit, sdk.MustNewDecFromStr("1")), "usdx:usd", sdk.NewInt(USDX_CF), sdk.NewInt(USDX_CF*1000), types.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec()),
					types.NewMoneyMarket("busd", types.NewBorrowLimit(false, sdk.NewDec(100000000*BUSD_CF), sdk.MustNewDecFromStr("1")), "busd:usd", sdk.NewInt(BUSD_CF), sdk.NewInt(BUSD_CF*1000), types.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec()),
					types.NewMoneyMarket("ukava", types.NewBorrowLimit(false, sdk.NewDec(100000000*KAVA_CF), tc.args.loanToValueKAVA), "kava:usd", sdk.NewInt(KAVA_CF), sdk.NewInt(KAVA_CF*1000), types.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec()),
					types.NewMoneyMarket("btcb", types.NewBorrowLimit(false, sdk.NewDec(100000000*BTCB_CF), tc.args.loanToValueBTCB), "btcb:usd", sdk.NewInt(BTCB_CF), sdk.NewInt(BTCB_CF*1000), types.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec()),
					types.NewMoneyMarket("bnb", types.NewBorrowLimit(false, sdk.NewDec(100000000*BNB_CF), tc.args.loanToValueBNB), "bnb:usd", sdk.NewInt(BNB_CF), sdk.NewInt(BNB_CF*1000), types.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec()),
					types.NewMoneyMarket("xyz", types.NewBorrowLimit(false, sdk.NewDec(1), tc.args.loanToValueBNB), "xyz:usd", sdk.NewInt(1), sdk.NewInt(1), types.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec()),
				},
				0, // LTV counter
			), types.DefaultAccumulationTimes, types.DefaultDeposits, types.DefaultBorrows,
				types.DefaultTotalSupplied, types.DefaultTotalBorrowed, types.DefaultTotalReserves,
			)

			// Pricefeed module genesis state
			pricefeedGS := pricefeed.GenesisState{
				Params: pricefeed.Params{
					Markets: []pricefeed.Market{
						{MarketID: "usdx:usd", BaseAsset: "usdx", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
						{MarketID: "busd:usd", BaseAsset: "busd", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
						{MarketID: "kava:usd", BaseAsset: "kava", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
						{MarketID: "btcb:usd", BaseAsset: "btcb", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
						{MarketID: "bnb:usd", BaseAsset: "bnb", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
						{MarketID: "xyz:usd", BaseAsset: "xyz", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
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
						MarketID:      "busd:usd",
						OracleAddress: sdk.AccAddress{},
						Price:         sdk.MustNewDecFromStr("1.00"),
						Expiry:        time.Now().Add(1 * time.Hour),
					},
					{
						MarketID:      "kava:usd",
						OracleAddress: sdk.AccAddress{},
						Price:         tc.args.priceKAVA,
						Expiry:        time.Now().Add(1 * time.Hour),
					},
					{
						MarketID:      "btcb:usd",
						OracleAddress: sdk.AccAddress{},
						Price:         tc.args.priceBTCB,
						Expiry:        time.Now().Add(1 * time.Hour),
					},
					{
						MarketID:      "bnb:usd",
						OracleAddress: sdk.AccAddress{},
						Price:         tc.args.priceBNB,
						Expiry:        time.Now().Add(1 * time.Hour),
					},
				},
			}

			// Initialize test application
			tApp.InitializeFromGenesisStates(authGS,
				app.GenesisState{pricefeed.ModuleName: pricefeed.ModuleCdc.MustMarshalJSON(pricefeedGS)},
				app.GenesisState{types.ModuleName: types.ModuleCdc.MustMarshalJSON(hardGS)})

			// Mint coins to hard module account
			supplyKeeper := tApp.GetSupplyKeeper()
			hardMaccCoins := sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1000*KAVA_CF)),
				sdk.NewCoin("usdx", sdk.NewInt(200*USDX_CF)), sdk.NewCoin("busd", sdk.NewInt(100*BUSD_CF)))
			supplyKeeper.MintCoins(ctx, types.ModuleAccountName, hardMaccCoins)

			keeper := tApp.GetHardKeeper()
			suite.app = tApp
			suite.ctx = ctx
			suite.keeper = keeper

			var err error

			// Run BeginBlocker once to transition MoneyMarkets
			hard.BeginBlocker(suite.ctx, suite.keeper)

			err = suite.keeper.Deposit(suite.ctx, tc.args.borrower, tc.args.depositCoins)
			suite.Require().NoError(err)

			// Execute user's previous borrows
			err = suite.keeper.Borrow(suite.ctx, tc.args.borrower, tc.args.previousBorrowCoins)
			if tc.args.previousBorrowCoins.IsZero() {
				suite.Require().True(strings.Contains(err.Error(), "cannot borrow zero coins"))
			} else {
				suite.Require().NoError(err)
			}

			// Now that our state is properly set up, execute the last borrow
			err = suite.keeper.Borrow(suite.ctx, tc.args.borrower, tc.args.borrowCoins)

			if tc.errArgs.expectPass {
				suite.Require().NoError(err)

				// Check borrower balance
				acc := suite.getAccount(tc.args.borrower)
				suite.Require().Equal(tc.args.expectedAccountBalance, acc.GetCoins())

				// Check module account balance
				mAcc := suite.getModuleAccount(types.ModuleAccountName)
				suite.Require().Equal(tc.args.expectedModAccountBalance, mAcc.GetCoins())

				// Check that borrow struct is in store
				_, f := suite.keeper.GetBorrow(suite.ctx, tc.args.borrower)
				suite.Require().True(f)
			} else {
				suite.Require().Error(err)
				suite.Require().True(strings.Contains(err.Error(), tc.errArgs.contains))
			}
		})
	}
}
