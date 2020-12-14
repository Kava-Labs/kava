package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	auctypes "github.com/kava-labs/kava/x/auction/types"
	"github.com/kava-labs/kava/x/harvest"
	"github.com/kava-labs/kava/x/harvest/types"
	"github.com/kava-labs/kava/x/pricefeed"
)

func (suite *KeeperTestSuite) TestKeeperLiquidation() {
	type args struct {
		borrower              sdk.AccAddress
		keeper                sdk.AccAddress
		keeperRewardPercent   sdk.Dec
		initialModuleCoins    sdk.Coins
		initialBorrowerCoins  sdk.Coins
		initialKeeperCoins    sdk.Coins
		depositCoins          []sdk.Coin
		borrowCoins           sdk.Coins
		liquidateAfter        int64
		auctionSize           sdk.Int
		expectedKeeperCoins   sdk.Coins         // coins keeper address should have after successfully liquidating position
		expectedBorrowerCoins sdk.Coins         // additional coins (if any) the borrower address should have after successfully liquidating position
		expectedAuctions      auctypes.Auctions // the auctions we should expect to find have been started
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

	// Set up auction constants
	layout := "2006-01-02T15:04:05.000Z"
	endTimeStr := "9000-01-01T00:00:00.000Z"
	endTime, _ := time.Parse(layout, endTimeStr)

	lotReturns, _ := auctypes.NewWeightedAddresses([]sdk.AccAddress{borrower}, []sdk.Int{sdk.NewInt(100)})

	testCases := []liqTest{
		{
			"valid: keeper liquidates borrow",
			args{
				borrower:              borrower,
				keeper:                keeper,
				keeperRewardPercent:   sdk.MustNewDecFromStr("0.05"),
				initialModuleCoins:    sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))),
				initialBorrowerCoins:  sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))),
				initialKeeperCoins:    sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))),
				depositCoins:          []sdk.Coin{sdk.NewCoin("ukava", sdk.NewInt(10*KAVA_CF))},
				borrowCoins:           sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(8*KAVA_CF))),
				liquidateAfter:        oneMonthInSeconds,
				auctionSize:           sdk.NewInt(KAVA_CF * 1000),
				expectedKeeperCoins:   sdk.Coins{},
				expectedBorrowerCoins: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(98000001))), // initial - deposit + borrow + liquidation leftovers
				expectedAuctions: auctypes.Auctions{
					auctypes.CollateralAuction{
						BaseAuction: auctypes.BaseAuction{
							ID:              1,
							Initiator:       "harvest_liquidator",
							Lot:             sdk.NewInt64Coin("ukava", 9499999),
							Bidder:          nil,
							Bid:             sdk.NewInt64Coin("ukava", 0),
							HasReceivedBids: false,
							EndTime:         endTime,
							MaxEndTime:      endTime,
						},
						CorrespondingDebt: sdk.NewInt64Coin("debt", 0),
						MaxBid:            sdk.NewInt64Coin("ukava", 8004766),
						LotReturns:        lotReturns,
					},
				},
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"valid: multiple deposits, multiple borrows",
			args{
				borrower:              borrower,
				keeper:                keeper,
				keeperRewardPercent:   sdk.MustNewDecFromStr("0.05"),
				initialModuleCoins:    sdk.NewCoins(sdk.NewCoin("usdx", sdk.NewInt(1000*KAVA_CF)), sdk.NewCoin("usdt", sdk.NewInt(1000*KAVA_CF)), sdk.NewCoin("dai", sdk.NewInt(1000*KAVA_CF)), sdk.NewCoin("usdc", sdk.NewInt(1000*KAVA_CF))),
				initialBorrowerCoins:  sdk.NewCoins(sdk.NewCoin("usdx", sdk.NewInt(1000*KAVA_CF)), sdk.NewCoin("usdt", sdk.NewInt(1000*KAVA_CF)), sdk.NewCoin("dai", sdk.NewInt(1000*KAVA_CF)), sdk.NewCoin("usdc", sdk.NewInt(1000*KAVA_CF))),
				initialKeeperCoins:    sdk.NewCoins(sdk.NewCoin("usdx", sdk.NewInt(1000*KAVA_CF)), sdk.NewCoin("usdt", sdk.NewInt(1000*KAVA_CF)), sdk.NewCoin("dai", sdk.NewInt(1000*KAVA_CF)), sdk.NewCoin("usdc", sdk.NewInt(1000*KAVA_CF))),
				depositCoins:          []sdk.Coin{sdk.NewCoin("dai", sdk.NewInt(350*KAVA_CF)), sdk.NewCoin("usdc", sdk.NewInt(200*KAVA_CF))},
				borrowCoins:           sdk.NewCoins(sdk.NewCoin("usdt", sdk.NewInt(250*KAVA_CF)), sdk.NewCoin("usdx", sdk.NewInt(245*KAVA_CF))),
				liquidateAfter:        oneMonthInSeconds,
				auctionSize:           sdk.NewInt(KAVA_CF * 100000),
				expectedKeeperCoins:   sdk.Coins{},
				expectedBorrowerCoins: sdk.NewCoins(sdk.NewCoin("dai", sdk.NewInt(650*KAVA_CF)), sdk.NewCoin("usdc", sdk.NewInt(800000001)), sdk.NewCoin("usdt", sdk.NewInt(1250*KAVA_CF)), sdk.NewCoin("usdx", sdk.NewInt(1245*KAVA_CF))),
				expectedAuctions: auctypes.Auctions{
					auctypes.CollateralAuction{
						BaseAuction: auctypes.BaseAuction{
							ID:              1,
							Initiator:       "harvest_liquidator",
							Lot:             sdk.NewInt64Coin("dai", 263894126),
							Bidder:          nil,
							Bid:             sdk.NewInt64Coin("usdt", 0),
							HasReceivedBids: false,
							EndTime:         endTime,
							MaxEndTime:      endTime,
						},
						CorrespondingDebt: sdk.NewInt64Coin("debt", 0),
						MaxBid:            sdk.NewInt64Coin("usdt", 250507897),
						LotReturns:        lotReturns,
					},
					auctypes.CollateralAuction{
						BaseAuction: auctypes.BaseAuction{
							ID:              2,
							Initiator:       "harvest_liquidator",
							Lot:             sdk.NewInt64Coin("dai", 68605874),
							Bidder:          nil,
							Bid:             sdk.NewInt64Coin("usdx", 0),
							HasReceivedBids: false,
							EndTime:         endTime,
							MaxEndTime:      endTime,
						},
						CorrespondingDebt: sdk.NewInt64Coin("debt", 0),
						MaxBid:            sdk.NewInt64Coin("usdx", 65125788),
						LotReturns:        lotReturns,
					},
					auctypes.CollateralAuction{
						BaseAuction: auctypes.BaseAuction{
							ID:              3,
							Initiator:       "harvest_liquidator",
							Lot:             sdk.NewInt64Coin("usdc", 189999999),
							Bidder:          nil,
							Bid:             sdk.NewInt64Coin("usdx", 0),
							HasReceivedBids: false,
							EndTime:         endTime,
							MaxEndTime:      endTime,
						},
						CorrespondingDebt: sdk.NewInt64Coin("debt", 0),
						MaxBid:            sdk.NewInt64Coin("usdx", 180362106),
						LotReturns:        lotReturns,
					},
				},
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
					types.NewDistributionSchedule(true, "usdx", time.Date(2020, 10, 8, 14, 0, 0, 0, time.UTC), time.Date(2020, 11, 22, 14, 0, 0, 0, time.UTC), sdk.NewCoin("hard", sdk.NewInt(5000)), time.Date(2021, 11, 22, 14, 0, 0, 0, time.UTC), types.Multipliers{types.NewMultiplier(types.Small, 0, sdk.MustNewDecFromStr("0.33")), types.NewMultiplier(types.Medium, 6, sdk.MustNewDecFromStr("0.5")), types.NewMultiplier(types.Medium, 24, sdk.OneDec())}),
					types.NewDistributionSchedule(true, "usdc", time.Date(2020, 10, 8, 14, 0, 0, 0, time.UTC), time.Date(2020, 11, 22, 14, 0, 0, 0, time.UTC), sdk.NewCoin("hard", sdk.NewInt(5000)), time.Date(2021, 11, 22, 14, 0, 0, 0, time.UTC), types.Multipliers{types.NewMultiplier(types.Small, 0, sdk.MustNewDecFromStr("0.33")), types.NewMultiplier(types.Medium, 6, sdk.MustNewDecFromStr("0.5")), types.NewMultiplier(types.Medium, 24, sdk.OneDec())}),
					types.NewDistributionSchedule(true, "usdt", time.Date(2020, 10, 8, 14, 0, 0, 0, time.UTC), time.Date(2020, 11, 22, 14, 0, 0, 0, time.UTC), sdk.NewCoin("hard", sdk.NewInt(5000)), time.Date(2021, 11, 22, 14, 0, 0, 0, time.UTC), types.Multipliers{types.NewMultiplier(types.Small, 0, sdk.MustNewDecFromStr("0.33")), types.NewMultiplier(types.Medium, 6, sdk.MustNewDecFromStr("0.5")), types.NewMultiplier(types.Medium, 24, sdk.OneDec())}),
					types.NewDistributionSchedule(true, "dai", time.Date(2020, 10, 8, 14, 0, 0, 0, time.UTC), time.Date(2020, 11, 22, 14, 0, 0, 0, time.UTC), sdk.NewCoin("hard", sdk.NewInt(5000)), time.Date(2021, 11, 22, 14, 0, 0, 0, time.UTC), types.Multipliers{types.NewMultiplier(types.Small, 0, sdk.MustNewDecFromStr("0.33")), types.NewMultiplier(types.Medium, 6, sdk.MustNewDecFromStr("0.5")), types.NewMultiplier(types.Medium, 24, sdk.OneDec())}),
					types.NewDistributionSchedule(true, "ukava", time.Date(2020, 10, 8, 14, 0, 0, 0, time.UTC), time.Date(2020, 11, 22, 14, 0, 0, 0, time.UTC), sdk.NewCoin("hard", sdk.NewInt(5000)), time.Date(2021, 11, 22, 14, 0, 0, 0, time.UTC), types.Multipliers{types.NewMultiplier(types.Small, 0, sdk.MustNewDecFromStr("0.33")), types.NewMultiplier(types.Medium, 6, sdk.MustNewDecFromStr("0.5")), types.NewMultiplier(types.Medium, 24, sdk.OneDec())}),
				},
				types.DelegatorDistributionSchedules{types.NewDelegatorDistributionSchedule(
					types.NewDistributionSchedule(true, "usdx", time.Date(2020, 10, 8, 14, 0, 0, 0, time.UTC), time.Date(2025, 10, 8, 14, 0, 0, 0, time.UTC), sdk.NewCoin("hard", sdk.NewInt(500)), time.Date(2026, 10, 8, 14, 0, 0, 0, time.UTC), types.Multipliers{types.NewMultiplier(types.Small, 0, sdk.MustNewDecFromStr("0.33")), types.NewMultiplier(types.Medium, 6, sdk.MustNewDecFromStr("0.5")), types.NewMultiplier(types.Medium, 24, sdk.OneDec())}),
					time.Hour*24,
				),
				},
				types.MoneyMarkets{
					types.NewMoneyMarket("usdx",
						types.NewBorrowLimit(false, sdk.NewDec(100000000*KAVA_CF), sdk.MustNewDecFromStr("0.9")), // Borrow Limit
						"usdx:usd",                   // Market ID
						sdk.NewInt(KAVA_CF),          // Conversion Factor
						tc.args.auctionSize,          // Auction Size
						model,                        // Interest Rate Model
						reserveFactor,                // Reserve Factor
						tc.args.keeperRewardPercent), // Keeper Reward Percent
					types.NewMoneyMarket("usdt",
						types.NewBorrowLimit(false, sdk.NewDec(100000000*KAVA_CF), sdk.MustNewDecFromStr("0.9")), // Borrow Limit
						"usdt:usd",                   // Market ID
						sdk.NewInt(KAVA_CF),          // Conversion Factor
						tc.args.auctionSize,          // Auction Size
						model,                        // Interest Rate Model
						reserveFactor,                // Reserve Factor
						tc.args.keeperRewardPercent), // Keeper Reward Percent
					types.NewMoneyMarket("usdc",
						types.NewBorrowLimit(false, sdk.NewDec(100000000*KAVA_CF), sdk.MustNewDecFromStr("0.9")), // Borrow Limit
						"usdc:usd",                   // Market ID
						sdk.NewInt(KAVA_CF),          // Conversion Factor
						tc.args.auctionSize,          // Auction Size
						model,                        // Interest Rate Model
						reserveFactor,                // Reserve Factor
						tc.args.keeperRewardPercent), // Keeper Reward Percent
					types.NewMoneyMarket("dai",
						types.NewBorrowLimit(false, sdk.NewDec(100000000*KAVA_CF), sdk.MustNewDecFromStr("0.9")), // Borrow Limit
						"dai:usd",                    // Market ID
						sdk.NewInt(KAVA_CF),          // Conversion Factor
						tc.args.auctionSize,          // Auction Size
						model,                        // Interest Rate Model
						reserveFactor,                // Reserve Factor
						tc.args.keeperRewardPercent), // Keeper Reward Percent
					types.NewMoneyMarket("ukava",
						types.NewBorrowLimit(false, sdk.NewDec(100000000*KAVA_CF), sdk.MustNewDecFromStr("0.8")), // Borrow Limit
						"kava:usd",                   // Market ID
						sdk.NewInt(KAVA_CF),          // Conversion Factor
						tc.args.auctionSize,          // Auction Size
						model,                        // Interest Rate Model
						reserveFactor,                // Reserve Factor
						tc.args.keeperRewardPercent), // Keeper Reward Percent
				},
			), types.DefaultPreviousBlockTime, types.DefaultDistributionTimes)

			// Pricefeed module genesis state
			pricefeedGS := pricefeed.GenesisState{
				Params: pricefeed.Params{
					Markets: []pricefeed.Market{
						{MarketID: "usdx:usd", BaseAsset: "usdx", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
						{MarketID: "usdt:usd", BaseAsset: "usdt", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
						{MarketID: "usdc:usd", BaseAsset: "usdc", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
						{MarketID: "dai:usd", BaseAsset: "dai", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
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
						MarketID:      "usdt:usd",
						OracleAddress: sdk.AccAddress{},
						Price:         sdk.MustNewDecFromStr("1.00"),
						Expiry:        time.Now().Add(100 * time.Hour),
					},
					{
						MarketID:      "usdc:usd",
						OracleAddress: sdk.AccAddress{},
						Price:         sdk.MustNewDecFromStr("1.00"),
						Expiry:        time.Now().Add(100 * time.Hour),
					},
					{
						MarketID:      "dai:usd",
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

			keeper := tApp.GetHarvestKeeper()
			suite.app = tApp
			suite.ctx = ctx
			suite.keeper = keeper
			suite.auctionKeeper = auctionKeeper

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

				// Check that borrower's balance contains the expected coins
				accBorrower := suite.getAccountAtCtx(tc.args.borrower, liqCtx)
				suite.Require().Equal(tc.args.expectedBorrowerCoins, accBorrower.GetCoins())

				// Check that the expected auctions have been created
				auctions := suite.auctionKeeper.GetAllAuctions(liqCtx)
				suite.Require().True(len(auctions) > 0)
				suite.Require().Equal(tc.args.expectedAuctions, auctions)
			} else {
				suite.Require().Error(err)

				// Check that the user's borrow exists
				_, foundBorrowAfter := suite.keeper.GetBorrow(liqCtx, tc.args.borrower)
				suite.Require().True(foundBorrowAfter)
				// Check that the user's deposits exist
				for _, coin := range tc.args.depositCoins {
					_, foundDepositAfter := suite.keeper.GetDeposit(liqCtx, tc.args.borrower, coin.Denom)
					suite.Require().True(foundDepositAfter)
				}

				// Check that no auctions have been created
				auctions := suite.auctionKeeper.GetAllAuctions(liqCtx)
				suite.Require().True(len(auctions) == 0)
			}
		})
	}
}
