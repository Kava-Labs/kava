package keeper_test

import (
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

func (suite *KeeperTestSuite) TestRepay() {
	type args struct {
		borrower             sdk.AccAddress
		repayer              sdk.AccAddress
		initialBorrowerCoins sdk.Coins
		initialRepayerCoins  sdk.Coins
		initialModuleCoins   sdk.Coins
		depositCoins         []sdk.Coin
		borrowCoins          sdk.Coins
		repayCoins           sdk.Coins
	}

	type errArgs struct {
		expectPass   bool
		expectDelete bool
		contains     string
	}

	type borrowTest struct {
		name    string
		args    args
		errArgs errArgs
	}

	model := types.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10"))

	testCases := []borrowTest{
		{
			"valid: partial repay",
			args{
				borrower:             sdk.AccAddress(crypto.AddressHash([]byte("borrower"))),
				repayer:              sdk.AccAddress(crypto.AddressHash([]byte("borrower"))),
				initialBorrowerCoins: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))),
				initialRepayerCoins:  sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))),
				initialModuleCoins:   sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1000*KAVA_CF)), sdk.NewCoin("usdx", sdk.NewInt(1000*USDX_CF))),
				depositCoins:         sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))),
				borrowCoins:          sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(50*KAVA_CF))),
				repayCoins:           sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(10*KAVA_CF))),
			},
			errArgs{
				expectPass:   true,
				expectDelete: false,
				contains:     "",
			},
		},
		{
			"valid: partial repay by non borrower",
			args{
				borrower:             sdk.AccAddress(crypto.AddressHash([]byte("borrower"))),
				repayer:              sdk.AccAddress(crypto.AddressHash([]byte("repayer"))),
				initialBorrowerCoins: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))),
				initialRepayerCoins:  sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))),
				initialModuleCoins:   sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1000*KAVA_CF)), sdk.NewCoin("usdx", sdk.NewInt(1000*USDX_CF))),
				depositCoins:         sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))),
				borrowCoins:          sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(50*KAVA_CF))),
				repayCoins:           sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(10*KAVA_CF))),
			},
			errArgs{
				expectPass:   true,
				expectDelete: false,
				contains:     "",
			},
		},
		{
			"valid: repay in full",
			args{
				borrower:             sdk.AccAddress(crypto.AddressHash([]byte("borrower"))),
				repayer:              sdk.AccAddress(crypto.AddressHash([]byte("borrower"))),
				initialBorrowerCoins: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))),
				initialRepayerCoins:  sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))),
				initialModuleCoins:   sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1000*KAVA_CF)), sdk.NewCoin("usdx", sdk.NewInt(1000*USDX_CF))),
				depositCoins:         sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))),
				borrowCoins:          sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(50*KAVA_CF))),
				repayCoins:           sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(50*KAVA_CF))),
			},
			errArgs{
				expectPass:   true,
				expectDelete: true,
				contains:     "",
			},
		},
		{
			"valid: overpayment is adjusted",
			args{
				borrower:             sdk.AccAddress(crypto.AddressHash([]byte("borrower"))),
				repayer:              sdk.AccAddress(crypto.AddressHash([]byte("borrower"))),
				initialBorrowerCoins: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))),
				initialRepayerCoins:  sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))),
				initialModuleCoins:   sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1000*KAVA_CF)), sdk.NewCoin("usdx", sdk.NewInt(1000*USDX_CF))),
				depositCoins:         sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(80*KAVA_CF))), // Deposit less so user still has some KAVA
				borrowCoins:          sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(50*KAVA_CF))),
				repayCoins:           sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(60*KAVA_CF))), // Exceeds borrowed coins but not user's balance
			},
			errArgs{
				expectPass:   true,
				expectDelete: true,
				contains:     "",
			},
		},
		{
			"invalid: attempt to repay non-supplied coin",
			args{
				borrower:             sdk.AccAddress(crypto.AddressHash([]byte("borrower"))),
				repayer:              sdk.AccAddress(crypto.AddressHash([]byte("borrower"))),
				initialBorrowerCoins: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))),
				initialRepayerCoins:  sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))),
				initialModuleCoins:   sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1000*KAVA_CF)), sdk.NewCoin("usdx", sdk.NewInt(1000*USDX_CF))),
				depositCoins:         sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))),
				borrowCoins:          sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(50*KAVA_CF))),
				repayCoins:           sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(10*KAVA_CF)), sdk.NewCoin("bnb", sdk.NewInt(10*KAVA_CF))),
			},
			errArgs{
				expectPass:   false,
				expectDelete: false,
				contains:     "no coins of this type borrowed",
			},
		},
		{
			"invalid: insufficient balance for repay",
			args{
				borrower:             sdk.AccAddress(crypto.AddressHash([]byte("borrower"))),
				repayer:              sdk.AccAddress(crypto.AddressHash([]byte("repayer"))),
				initialBorrowerCoins: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))),
				initialRepayerCoins:  sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(49*KAVA_CF))),
				initialModuleCoins:   sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1000*KAVA_CF)), sdk.NewCoin("usdx", sdk.NewInt(1000*USDX_CF))),
				depositCoins:         sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))),
				borrowCoins:          sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(50*KAVA_CF))),
				repayCoins:           sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(50*KAVA_CF))), // Exceeds repayer's balance, but not borrow amount
			},
			errArgs{
				expectPass:   false,
				expectDelete: false,
				contains:     "account can only repay up to 49000000ukava",
			},
		},
		{
			"invalid: repaying a single coin type results in borrow position below the minimum USD value",
			args{
				borrower:             sdk.AccAddress(crypto.AddressHash([]byte("borrower"))),
				repayer:              sdk.AccAddress(crypto.AddressHash([]byte("borrower"))),
				initialBorrowerCoins: sdk.NewCoins(sdk.NewCoin("usdx", sdk.NewInt(100*USDX_CF))),
				initialRepayerCoins:  sdk.NewCoins(sdk.NewCoin("usdx", sdk.NewInt(100*USDX_CF))),
				initialModuleCoins:   sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1000*KAVA_CF)), sdk.NewCoin("usdx", sdk.NewInt(1000*USDX_CF))),
				depositCoins:         sdk.NewCoins(sdk.NewCoin("usdx", sdk.NewInt(100*USDX_CF))),
				borrowCoins:          sdk.NewCoins(sdk.NewCoin("usdx", sdk.NewInt(50*USDX_CF))),
				repayCoins:           sdk.NewCoins(sdk.NewCoin("usdx", sdk.NewInt(45*USDX_CF))),
			},
			errArgs{
				expectPass:   false,
				expectDelete: false,
				contains:     "proposed borrow's USD value $5.000000000000000000 is below the minimum borrow limit",
			},
		},
		{
			"invalid: repaying multiple coin types results in borrow position below the minimum USD value",
			args{
				borrower:             sdk.AccAddress(crypto.AddressHash([]byte("borrower"))),
				repayer:              sdk.AccAddress(crypto.AddressHash([]byte("borrower"))),
				initialBorrowerCoins: sdk.NewCoins(sdk.NewCoin("usdx", sdk.NewInt(100*USDX_CF))),
				initialRepayerCoins:  sdk.NewCoins(sdk.NewCoin("usdx", sdk.NewInt(100*USDX_CF)), sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))),
				initialModuleCoins:   sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1000*KAVA_CF)), sdk.NewCoin("usdx", sdk.NewInt(1000*USDX_CF))),
				depositCoins:         sdk.NewCoins(sdk.NewCoin("usdx", sdk.NewInt(100*USDX_CF))),
				borrowCoins:          sdk.NewCoins(sdk.NewCoin("usdx", sdk.NewInt(50*USDX_CF)), sdk.NewCoin("ukava", sdk.NewInt(10*KAVA_CF))), // (50*$1)+(10*$2) = $70
				repayCoins:           sdk.NewCoins(sdk.NewCoin("usdx", sdk.NewInt(45*USDX_CF)), sdk.NewCoin("ukava", sdk.NewInt(8*KAVA_CF))),  // (45*$1)+(8*$2) = $61
			},
			errArgs{
				expectPass:   false,
				expectDelete: false,
				contains:     "proposed borrow's USD value $9.000000000000000000 is below the minimum borrow limit",
			},
		},
		{
			"invalid: overpaying multiple coin types results in borrow position below the minimum USD value",
			args{
				borrower:             sdk.AccAddress(crypto.AddressHash([]byte("borrower"))),
				repayer:              sdk.AccAddress(crypto.AddressHash([]byte("borrower"))),
				initialBorrowerCoins: sdk.NewCoins(sdk.NewCoin("usdx", sdk.NewInt(100*USDX_CF))),
				initialRepayerCoins:  sdk.NewCoins(sdk.NewCoin("usdx", sdk.NewInt(100*USDX_CF)), sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))),
				initialModuleCoins:   sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1000*KAVA_CF)), sdk.NewCoin("usdx", sdk.NewInt(1000*USDX_CF))),
				depositCoins:         sdk.NewCoins(sdk.NewCoin("usdx", sdk.NewInt(100*USDX_CF))),
				borrowCoins:          sdk.NewCoins(sdk.NewCoin("usdx", sdk.NewInt(50*USDX_CF)), sdk.NewCoin("ukava", sdk.NewInt(10*KAVA_CF))), // (50*$1)+(10*$2) = $70
				repayCoins:           sdk.NewCoins(sdk.NewCoin("usdx", sdk.NewInt(500*USDX_CF)), sdk.NewCoin("ukava", sdk.NewInt(8*KAVA_CF))), // (500*$1)+(8*$2) = $516, or capping to borrowed amount, (50*$1)+(8*$2) = $66
			},
			errArgs{
				expectPass:   false,
				expectDelete: false,
				contains:     "proposed borrow's USD value $4.000000000000000000 is below the minimum borrow limit",
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Initialize test app and set context
			tApp := app.NewTestApp()
			ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})

			// Auth module genesis state
			addrs, coinses := uniqueAddressCoins(
				[]sdk.AccAddress{tc.args.borrower, tc.args.repayer},
				[]sdk.Coins{tc.args.initialBorrowerCoins, tc.args.initialRepayerCoins},
			)
			authGS := app.NewAuthGenState(addrs, coinses)

			// Hard module genesis state
			hardGS := types.NewGenesisState(types.NewParams(
				types.MoneyMarkets{
					types.NewMoneyMarket("usdx",
						types.NewBorrowLimit(false, sdk.NewDec(100000000*USDX_CF), sdk.MustNewDecFromStr("1")), // Borrow Limit
						"usdx:usd",                     // Market ID
						sdk.NewInt(USDX_CF),            // Conversion Factor
						model,                          // Interest Rate Model
						sdk.MustNewDecFromStr("0.05"),  // Reserve Factor
						sdk.MustNewDecFromStr("0.05")), // Keeper Reward Percent
					types.NewMoneyMarket("ukava",
						types.NewBorrowLimit(false, sdk.NewDec(100000000*KAVA_CF), sdk.MustNewDecFromStr("0.8")), // Borrow Limit
						"kava:usd",                     // Market ID
						sdk.NewInt(KAVA_CF),            // Conversion Factor
						model,                          // Interest Rate Model
						sdk.MustNewDecFromStr("0.05"),  // Reserve Factor
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
						Expiry:        time.Now().Add(1 * time.Hour),
					},
					{
						MarketID:      "kava:usd",
						OracleAddress: sdk.AccAddress{},
						Price:         sdk.MustNewDecFromStr("2.00"),
						Expiry:        time.Now().Add(1 * time.Hour),
					},
				},
			}

			// Initialize test application
			tApp.InitializeFromGenesisStates(authGS,
				app.GenesisState{pricefeed.ModuleName: pricefeed.ModuleCdc.MustMarshalJSON(pricefeedGS)},
				app.GenesisState{types.ModuleName: types.ModuleCdc.MustMarshalJSON(hardGS)},
			)

			// Mint coins to Hard module account
			supplyKeeper := tApp.GetSupplyKeeper()
			supplyKeeper.MintCoins(ctx, types.ModuleAccountName, tc.args.initialModuleCoins)

			keeper := tApp.GetHardKeeper()
			suite.app = tApp
			suite.ctx = ctx
			suite.keeper = keeper

			var err error

			// Run BeginBlocker once to transition MoneyMarkets
			hard.BeginBlocker(suite.ctx, suite.keeper)

			// Deposit coins to hard
			err = suite.keeper.Deposit(suite.ctx, tc.args.borrower, tc.args.depositCoins)
			suite.Require().NoError(err)

			// Borrow coins from hard
			err = suite.keeper.Borrow(suite.ctx, tc.args.borrower, tc.args.borrowCoins)
			suite.Require().NoError(err)

			previousRepayerCoins := suite.getAccount(tc.args.repayer).GetCoins()

			err = suite.keeper.Repay(suite.ctx, tc.args.repayer, tc.args.borrower, tc.args.repayCoins)
			if tc.errArgs.expectPass {
				suite.Require().NoError(err)
				// If we overpaid expect an adjustment
				repaymentCoins, err := suite.keeper.CalculatePaymentAmount(tc.args.borrowCoins, tc.args.repayCoins)
				suite.Require().NoError(err)

				// Check repayer balance
				expectedRepayerCoins := previousRepayerCoins.Sub(repaymentCoins)
				acc := suite.getAccount(tc.args.repayer)
				suite.Require().Equal(expectedRepayerCoins, acc.GetCoins())

				// Check module account balance
				expectedModuleCoins := tc.args.initialModuleCoins.Add(tc.args.depositCoins...).Sub(tc.args.borrowCoins).Add(repaymentCoins...)
				mAcc := suite.getModuleAccount(types.ModuleAccountName)
				suite.Require().Equal(expectedModuleCoins, mAcc.GetCoins())

				// Check user's borrow object
				borrow, foundBorrow := suite.keeper.GetBorrow(suite.ctx, tc.args.borrower)
				expectedBorrowCoins := tc.args.borrowCoins.Sub(repaymentCoins)

				if tc.errArgs.expectDelete {
					suite.Require().False(foundBorrow)
				} else {
					suite.Require().True(foundBorrow)
					suite.Require().Equal(expectedBorrowCoins, borrow.Amount)
				}
			} else {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.errArgs.contains)

				// Check repayer balance (no repay coins)
				acc := suite.getAccount(tc.args.repayer)
				suite.Require().Equal(previousRepayerCoins, acc.GetCoins())

				// Check module account balance (no repay coins)
				expectedModuleCoins := tc.args.initialModuleCoins.Add(tc.args.depositCoins...).Sub(tc.args.borrowCoins)
				mAcc := suite.getModuleAccount(types.ModuleAccountName)
				suite.Require().Equal(expectedModuleCoins, mAcc.GetCoins())

				// Check user's borrow object (no repay coins)
				borrow, foundBorrow := suite.keeper.GetBorrow(suite.ctx, tc.args.borrower)
				suite.Require().True(foundBorrow)
				suite.Require().Equal(tc.args.borrowCoins, borrow.Amount)
			}
		})
	}
}

// uniqueAddressCoins removes duplicate addresses, and the corresponding elements in a list of coins.
func uniqueAddressCoins(addresses []sdk.AccAddress, coinses []sdk.Coins) ([]sdk.AccAddress, []sdk.Coins) {
	uniqueAddresses := []sdk.AccAddress{}
	filteredCoins := []sdk.Coins{}

	addrMap := map[string]bool{}
	for i, a := range addresses {
		if !addrMap[a.String()] {
			uniqueAddresses = append(uniqueAddresses, a)
			filteredCoins = append(filteredCoins, coinses[i])
		}
		addrMap[a.String()] = true
	}
	return uniqueAddresses, filteredCoins
}
