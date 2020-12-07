package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/harvest"
	"github.com/kava-labs/kava/x/harvest/types"
	"github.com/kava-labs/kava/x/pricefeed"
)

// type LiquidationTestSuite struct {
// 	suite.Suite
// }

func (suite *KeeperTestSuite) TestKeeperLiquidation() {
	type args struct {
		borrower             sdk.AccAddress
		keeper               sdk.AccAddress
		keeperRewardPercent  sdk.Dec
		initialModuleCoins   sdk.Coins
		initialBorrowerCoins sdk.Coins
		initialKeeperCoins   sdk.Coins
		depositCoins         []sdk.Coin
		borrowCoins          sdk.Coins
		liquidateAfter       int64
		auctionSize          sdk.Int
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

	model := types.NewInterestRateModel(sdk.MustNewDecFromStr("0"), sdk.MustNewDecFromStr("0.1"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("0.5"))
	reserveFactor := sdk.MustNewDecFromStr("0.05")

	// oneDayInSeconds := int64(86400)
	// oneWeekInSeconds := int64(604800)
	oneMonthInSeconds := int64(2592000)
	// oneYearInSeconds := int64(31536000)

	borrower := sdk.AccAddress(crypto.AddressHash([]byte("testborrower")))
	keeper := sdk.AccAddress(crypto.AddressHash([]byte("testkeeper")))

	testCases := []liqTest{
		{
			"valid: keeper liquidates borrow",
			args{
				borrower:             borrower,
				keeper:               keeper,
				keeperRewardPercent:  sdk.MustNewDecFromStr("0.05"),
				initialModuleCoins:   sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))),
				initialBorrowerCoins: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))),
				initialKeeperCoins:   sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))),
				depositCoins:         []sdk.Coin{sdk.NewCoin("ukava", sdk.NewInt(10*KAVA_CF))},
				borrowCoins:          sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(8*KAVA_CF))),
				liquidateAfter:       oneMonthInSeconds,
				auctionSize:          sdk.NewInt(KAVA_CF * 1000),
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
			ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})

			// Auth module genesis state
			authGS := app.NewAuthGenState(
				[]sdk.AccAddress{tc.args.borrower, tc.args.keeper},
				[]sdk.Coins{tc.args.initialBorrowerCoins, tc.args.initialKeeperCoins},
			)

			// Harvest module genesis state
			harvestGS := types.NewGenesisState(types.NewParams(
				true,
				types.DistributionSchedules{
					types.NewDistributionSchedule(true, "ukava", time.Date(2020, 10, 8, 14, 0, 0, 0, time.UTC), time.Date(2020, 11, 22, 14, 0, 0, 0, time.UTC), sdk.NewCoin("hard", sdk.NewInt(5000)), time.Date(2021, 11, 22, 14, 0, 0, 0, time.UTC), types.Multipliers{types.NewMultiplier(types.Small, 0, sdk.MustNewDecFromStr("0.33")), types.NewMultiplier(types.Medium, 6, sdk.MustNewDecFromStr("0.5")), types.NewMultiplier(types.Medium, 24, sdk.OneDec())}),
				},
				types.DelegatorDistributionSchedules{types.NewDelegatorDistributionSchedule(
					types.NewDistributionSchedule(true, "usdx", time.Date(2020, 10, 8, 14, 0, 0, 0, time.UTC), time.Date(2025, 10, 8, 14, 0, 0, 0, time.UTC), sdk.NewCoin("hard", sdk.NewInt(500)), time.Date(2026, 10, 8, 14, 0, 0, 0, time.UTC), types.Multipliers{types.NewMultiplier(types.Small, 0, sdk.MustNewDecFromStr("0.33")), types.NewMultiplier(types.Medium, 6, sdk.MustNewDecFromStr("0.5")), types.NewMultiplier(types.Medium, 24, sdk.OneDec())}),
					time.Hour*24,
				),
				},
				types.MoneyMarkets{
					types.NewMoneyMarket("ukava",
						types.NewBorrowLimit(false, sdk.NewDec(100000000*KAVA_CF), sdk.MustNewDecFromStr("0.8")), // Borrow Limit
						"kava:usd",          // Market ID
						sdk.NewInt(KAVA_CF), // Conversion Factor
						tc.args.auctionSize, // Auction Size
						model,               // Interest Rate Model
						reserveFactor),      // Reserve Factor
				},
				tc.args.keeperRewardPercent,
			), types.DefaultPreviousBlockTime, types.DefaultDistributionTimes)

			// Pricefeed module genesis state
			pricefeedGS := pricefeed.GenesisState{
				Params: pricefeed.Params{
					Markets: []pricefeed.Market{
						{MarketID: "kava:usd", BaseAsset: "kava", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
					},
				},
				PostedPrices: []pricefeed.PostedPrice{
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

			keeper := tApp.GetHarvestKeeper()
			suite.app = tApp
			suite.ctx = ctx
			suite.keeper = keeper

			var err error

			// Run begin blocker to set up state
			harvest.BeginBlocker(suite.ctx, suite.keeper)

			// Deposit coins
			for _, coin := range tc.args.depositCoins {
				err = suite.keeper.Deposit(suite.ctx, tc.args.borrower, coin)
				suite.Require().NoError(err)
			}

			// Borrow coins
			err = suite.keeper.Borrow(suite.ctx, tc.args.borrower, tc.args.borrowCoins)
			suite.Require().NoError(err)

			// Set up liquidation chain context and run begin blocker
			runAtTime := time.Unix(suite.ctx.BlockTime().Unix()+(tc.args.liquidateAfter), 0)
			liqCtx := suite.ctx.WithBlockTime(runAtTime)
			harvest.BeginBlocker(liqCtx, suite.keeper)

			// Check borrow exists before liquidation
			_, foundBorrowBefore := suite.keeper.GetBorrow(liqCtx, tc.args.borrower)
			suite.Require().True(foundBorrowBefore)
			// Check that the user's deposits exist before liquidation
			for _, coin := range tc.args.depositCoins {
				_, foundDepositBefore := suite.keeper.GetDeposit(liqCtx, tc.args.borrower, coin.Denom)
				suite.Require().True(foundDepositBefore)
			}

			// Attempt to liquidate
			err = suite.keeper.AttemptKeeperLiquidation(liqCtx, tc.args.keeper, tc.args.borrower)
			if tc.errArgs.expectPass {
				suite.Require().NoError(err)

				// Check borrow does not exist after liquidation
				_, foundBorrowAfter := suite.keeper.GetBorrow(liqCtx, tc.args.borrower)
				suite.Require().False(foundBorrowAfter)
				// Check deposits do not exist after liquidation
				for _, coin := range tc.args.depositCoins {
					_, foundDepositAfter := suite.keeper.GetDeposit(liqCtx, tc.args.borrower, coin.Denom)
					suite.Require().False(foundDepositAfter)
				}

				// Check that the keeper's balance increased by reward % of all the borrowed coins
				rewardCoins := sdk.Coins{}
				for _, coin := range tc.args.depositCoins {
					reward := tc.args.keeperRewardPercent.MulInt(coin.Amount).TruncateInt()
					rewardCoins = append(rewardCoins, sdk.NewCoin(coin.Denom, reward))
				}
				accKeeper := suite.getAccountAtCtx(tc.args.keeper, liqCtx)
				suite.Require().Equal(tc.args.initialKeeperCoins.Add(rewardCoins...), accKeeper.GetCoins())

			} else {
				suite.Require().Error(err)

				// Check that the user's borrow exists
				_, foundBorrowAfter := suite.keeper.GetBorrow(liqCtx, tc.args.borrower)
				suite.Require().True(foundBorrowAfter)
				// Check that the user's deposits exist before liquidation
				for _, coin := range tc.args.depositCoins {
					_, foundDepositAfter := suite.keeper.GetDeposit(liqCtx, tc.args.borrower, coin.Denom)
					suite.Require().True(foundDepositAfter)
				}
			}
		})
	}
}

// func TestLiquidationTestSuite(t *testing.T) {
// 	suite.Run(t, new(LiquidationTestSuite))
// }
