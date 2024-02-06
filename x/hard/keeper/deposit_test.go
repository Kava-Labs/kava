package keeper_test

import (
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cometbft/cometbft/crypto"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtime "github.com/cometbft/cometbft/types/time"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/hard"
	"github.com/kava-labs/kava/x/hard/types"
	pricefeedtypes "github.com/kava-labs/kava/x/pricefeed/types"
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
				amount:                    sdk.NewCoins(sdk.NewCoin("bnb", sdkmath.NewInt(100))),
				numberDeposits:            1,
				expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("bnb", sdkmath.NewInt(900)), sdk.NewCoin("btcb", sdkmath.NewInt(1000))),
				expectedModAccountBalance: sdk.NewCoins(sdk.NewCoin("bnb", sdkmath.NewInt(100))),
				expectedDepositCoins:      sdk.NewCoins(sdk.NewCoin("bnb", sdkmath.NewInt(100))),
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
				amount:                    sdk.NewCoins(sdk.NewCoin("bnb", sdkmath.NewInt(100))),
				numberDeposits:            2,
				expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("bnb", sdkmath.NewInt(800)), sdk.NewCoin("btcb", sdkmath.NewInt(1000))),
				expectedModAccountBalance: sdk.NewCoins(sdk.NewCoin("bnb", sdkmath.NewInt(200))),
				expectedDepositCoins:      sdk.NewCoins(sdk.NewCoin("bnb", sdkmath.NewInt(200))),
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
				amount:                    sdk.NewCoins(sdk.NewCoin("fake", sdkmath.NewInt(100))),
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
				amount:                    sdk.NewCoins(sdk.NewCoin("bnb", sdkmath.NewInt(10000))),
				numberDeposits:            1,
				expectedAccountBalance:    sdk.Coins{},
				expectedModAccountBalance: sdk.Coins{},
				expectedDepositCoins:      sdk.Coins{},
			},
			errArgs{
				expectPass: false,
				contains:   "insufficient funds: the requested deposit amount",
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// create new app with one funded account

			// Initialize test app and set context
			tApp := app.NewTestApp()
			ctx := tApp.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})
			authGS := app.NewFundedGenStateWithCoins(
				tApp.AppCodec(),
				[]sdk.Coins{
					sdk.NewCoins(
						sdk.NewCoin("bnb", sdkmath.NewInt(1000)),
						sdk.NewCoin("btcb", sdkmath.NewInt(1000)),
					),
				},
				[]sdk.AccAddress{tc.args.depositor},
			)
			loanToValue, _ := sdk.NewDecFromStr("0.6")
			hardGS := types.NewGenesisState(types.NewParams(
				types.MoneyMarkets{
					types.NewMoneyMarket("usdx", types.NewBorrowLimit(false, sdk.NewDec(1000000000000000), loanToValue), "usdx:usd", sdkmath.NewInt(1000000), types.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec()),
					types.NewMoneyMarket("ukava", types.NewBorrowLimit(false, sdk.NewDec(1000000000000000), loanToValue), "kava:usd", sdkmath.NewInt(1000000), types.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec()),
					types.NewMoneyMarket("bnb", types.NewBorrowLimit(false, sdk.NewDec(1000000000000000), loanToValue), "bnb:usd", sdkmath.NewInt(1000000), types.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec()),
					types.NewMoneyMarket("btcb", types.NewBorrowLimit(false, sdk.NewDec(1000000000000000), loanToValue), "btcb:usd", sdkmath.NewInt(1000000), types.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec()),
				},
				sdk.NewDec(10),
			), types.DefaultAccumulationTimes, types.DefaultDeposits, types.DefaultBorrows,
				types.DefaultTotalSupplied, types.DefaultTotalBorrowed, types.DefaultTotalReserves,
			)

			// Pricefeed module genesis state
			pricefeedGS := pricefeedtypes.GenesisState{
				Params: pricefeedtypes.Params{
					Markets: []pricefeedtypes.Market{
						{MarketID: "usdx:usd", BaseAsset: "usdx", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
						{MarketID: "kava:usd", BaseAsset: "kava", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
						{MarketID: "btcb:usd", BaseAsset: "btcb", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
						{MarketID: "bnb:usd", BaseAsset: "bnb", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
					},
				},
				PostedPrices: []pricefeedtypes.PostedPrice{
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
				app.GenesisState{pricefeedtypes.ModuleName: tApp.AppCodec().MustMarshalJSON(&pricefeedGS)},
				app.GenesisState{types.ModuleName: tApp.AppCodec().MustMarshalJSON(&hardGS)},
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
				suite.Require().Equal(tc.args.expectedAccountBalance, suite.getAccountCoins(acc))
				mAcc := suite.getModuleAccount(types.ModuleAccountName)
				suite.Require().Equal(tc.args.expectedModAccountBalance, suite.getAccountCoins(mAcc))
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

func (suite *KeeperTestSuite) TestDecrementSuppliedCoins() {
	type args struct {
		suppliedInitial       sdk.Coins
		decrementCoins        sdk.Coins
		expectedSuppliedFinal sdk.Coins
	}
	type errArgs struct {
		expectPass bool
		contains   string
	}
	type decrementTest struct {
		name    string
		args    args
		errArgs errArgs
	}
	testCases := []decrementTest{
		{
			"valid",
			args{
				suppliedInitial:       cs(c("bnb", 10000000000000), c("busd", 3000000000000), c("xrpb", 2500000000000)),
				decrementCoins:        cs(c("bnb", 5000000000000)),
				expectedSuppliedFinal: cs(c("bnb", 5000000000000), c("busd", 3000000000000), c("xrpb", 2500000000000)),
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"valid-negative",
			args{
				suppliedInitial:       cs(c("bnb", 10000000000000), c("busd", 3000000000000), c("xrpb", 2500000000000)),
				decrementCoins:        cs(c("bnb", 10000000000001)),
				expectedSuppliedFinal: cs(c("busd", 3000000000000), c("xrpb", 2500000000000)),
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"valid-multiple negative",
			args{
				suppliedInitial:       cs(c("bnb", 10000000000000), c("busd", 3000000000000), c("xrpb", 2500000000000)),
				decrementCoins:        cs(c("bnb", 10000000000001), c("busd", 5000000000000)),
				expectedSuppliedFinal: cs(c("xrpb", 2500000000000)),
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"valid-absent coin denom",
			args{
				suppliedInitial:       cs(c("bnb", 10000000000000), c("xrpb", 2500000000000)),
				decrementCoins:        cs(c("busd", 5)),
				expectedSuppliedFinal: cs(c("bnb", 10000000000000), c("xrpb", 2500000000000)),
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Initialize test app and set context
			tApp := app.NewTestApp()
			ctx := tApp.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})
			loanToValue, _ := sdk.NewDecFromStr("0.6")
			depositor := sdk.AccAddress(crypto.AddressHash([]byte("test")))
			authGS := app.NewFundedGenStateWithCoins(
				tApp.AppCodec(),
				[]sdk.Coins{tc.args.suppliedInitial},
				[]sdk.AccAddress{depositor},
			)
			hardGS := types.NewGenesisState(types.NewParams(
				types.MoneyMarkets{
					types.NewMoneyMarket("bnb", types.NewBorrowLimit(false, sdk.NewDec(1000000000000000), loanToValue), "bnb:usd", sdkmath.NewInt(100000000), types.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec()),
					types.NewMoneyMarket("busd", types.NewBorrowLimit(false, sdk.NewDec(1000000000000000), loanToValue), "busd:usd", sdkmath.NewInt(100000000), types.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec()),
					types.NewMoneyMarket("xrpb", types.NewBorrowLimit(false, sdk.NewDec(1000000000000000), loanToValue), "xrpb:usd", sdkmath.NewInt(100000000), types.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec()),
				},
				sdk.MustNewDecFromStr("10"),
			), types.DefaultAccumulationTimes, types.DefaultDeposits, types.DefaultBorrows,
				types.DefaultTotalSupplied, types.DefaultTotalBorrowed, types.DefaultTotalReserves,
			)
			// Pricefeed module genesis state
			pricefeedGS := pricefeedtypes.GenesisState{
				Params: pricefeedtypes.Params{
					Markets: []pricefeedtypes.Market{
						{MarketID: "xrpb:usd", BaseAsset: "kava", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
						{MarketID: "busd:usd", BaseAsset: "btcb", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
						{MarketID: "bnb:usd", BaseAsset: "bnb", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
					},
				},
				PostedPrices: []pricefeedtypes.PostedPrice{
					{
						MarketID:      "busd:usd",
						OracleAddress: sdk.AccAddress{},
						Price:         sdk.MustNewDecFromStr("1.00"),
						Expiry:        time.Now().Add(1 * time.Hour),
					},
					{
						MarketID:      "xrpb:usd",
						OracleAddress: sdk.AccAddress{},
						Price:         sdk.MustNewDecFromStr("2.00"),
						Expiry:        time.Now().Add(1 * time.Hour),
					},
					{
						MarketID:      "bnb:usd",
						OracleAddress: sdk.AccAddress{},
						Price:         sdk.MustNewDecFromStr("200.00"),
						Expiry:        time.Now().Add(1 * time.Hour),
					},
				},
			}
			tApp.InitializeFromGenesisStates(authGS,
				app.GenesisState{pricefeedtypes.ModuleName: tApp.AppCodec().MustMarshalJSON(&pricefeedGS)},
				app.GenesisState{types.ModuleName: tApp.AppCodec().MustMarshalJSON(&hardGS)},
			)
			keeper := tApp.GetHardKeeper()
			suite.app = tApp
			suite.ctx = ctx
			suite.keeper = keeper

			// Run BeginBlocker once to transition MoneyMarkets
			hard.BeginBlocker(suite.ctx, suite.keeper)

			err := suite.keeper.Deposit(suite.ctx, depositor, tc.args.suppliedInitial)
			suite.Require().NoError(err)
			err = suite.keeper.DecrementSuppliedCoins(ctx, tc.args.decrementCoins)
			suite.Require().NoError(err)
			totalSuppliedActual, found := suite.keeper.GetSuppliedCoins(suite.ctx)
			suite.Require().True(found)
			suite.Require().Equal(totalSuppliedActual, tc.args.expectedSuppliedFinal)
		})
	}
}

func c(denom string, amount int64) sdk.Coin { return sdk.NewInt64Coin(denom, amount) }
func cs(coins ...sdk.Coin) sdk.Coins        { return sdk.NewCoins(coins...) }
