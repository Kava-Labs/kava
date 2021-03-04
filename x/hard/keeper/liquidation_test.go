package keeper_test

import (
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	auctypes "github.com/kava-labs/kava/x/auction/types"
	"github.com/kava-labs/kava/x/hard"
	"github.com/kava-labs/kava/x/hard/types"
	"github.com/kava-labs/kava/x/pricefeed"
)

func (suite *KeeperTestSuite) TestKeeperLiquidation() {
	type args struct {
		borrower                   sdk.AccAddress
		keeper                     sdk.AccAddress
		keeperRewardPercent        sdk.Dec
		initialModuleCoins         sdk.Coins
		initialBorrowerCoins       sdk.Coins
		initialKeeperCoins         sdk.Coins
		depositCoins               []sdk.Coin
		borrowCoins                sdk.Coins
		liquidateAfter             int64
		expectedTotalSuppliedCoins sdk.Coins
		expectedTotalBorrowedCoins sdk.Coins
		expectedKeeperCoins        sdk.Coins         // coins keeper address should have after successfully liquidating position
		expectedBorrowerCoins      sdk.Coins         // additional coins (if any) the borrower address should have after successfully liquidating position
		expectedAuctions           auctypes.Auctions // the auctions we should expect to find have been started
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
				borrower:                   borrower,
				keeper:                     keeper,
				keeperRewardPercent:        sdk.MustNewDecFromStr("0.05"),
				initialModuleCoins:         sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))),
				initialBorrowerCoins:       sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))),
				initialKeeperCoins:         sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))),
				depositCoins:               sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(10*KAVA_CF))),
				borrowCoins:                sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(8*KAVA_CF))),
				liquidateAfter:             oneMonthInSeconds,
				expectedTotalSuppliedCoins: sdk.NewCoins(sdk.NewInt64Coin("ukava", 504138)),
				expectedTotalBorrowedCoins: nil,
				expectedKeeperCoins:        sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100500020))),
				expectedBorrowerCoins:      sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(98000001))), // initial - deposit + borrow + liquidation leftovers
				expectedAuctions: auctypes.Auctions{
					auctypes.CollateralAuction{
						BaseAuction: auctypes.BaseAuction{
							ID:              1,
							Initiator:       "hard",
							Lot:             sdk.NewInt64Coin("ukava", 9500390),
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
			"valid: single deposit, multiple borrows",
			args{
				borrower:             borrower,
				keeper:               keeper,
				keeperRewardPercent:  sdk.MustNewDecFromStr("0.05"),
				initialModuleCoins:   sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1000*KAVA_CF)), sdk.NewCoin("usdc", sdk.NewInt(1000*KAVA_CF)), sdk.NewCoin("bnb", sdk.NewInt(1000*BNB_CF)), sdk.NewCoin("btc", sdk.NewInt(1000*BTCB_CF))),
				initialBorrowerCoins: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))),
				initialKeeperCoins:   sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))),
				depositCoins:         sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(50*KAVA_CF))),                                                                                                                                     // $100 * 0.8 = $80 borrowable
				borrowCoins:          sdk.NewCoins(sdk.NewCoin("usdc", sdk.NewInt(20*KAVA_CF)), sdk.NewCoin("ukava", sdk.NewInt(10*KAVA_CF)), sdk.NewCoin("bnb", sdk.NewInt(2*BNB_CF)), sdk.NewCoin("btc", sdk.NewInt(0.2*BTCB_CF))), // $20+$20+$20 = $80 borrowed
				liquidateAfter:       oneMonthInSeconds,
				expectedTotalSuppliedCoins: sdk.NewCoins(
					sdk.NewInt64Coin("ukava", 2500711),
					sdk.NewInt64Coin("bnb", 3123),
					sdk.NewInt64Coin("btc", 31),
					sdk.NewInt64Coin("usdc", 3120),
				),
				expectedTotalBorrowedCoins: nil,
				expectedKeeperCoins:        sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(102500001))),
				expectedBorrowerCoins:      sdk.NewCoins(sdk.NewCoin("usdc", sdk.NewInt(20*KAVA_CF)), sdk.NewCoin("ukava", sdk.NewInt(60000002)), sdk.NewCoin("bnb", sdk.NewInt(2*BNB_CF)), sdk.NewCoin("btc", sdk.NewInt(0.2*BTCB_CF))), // initial - deposit + borrow + liquidation leftovers
				expectedAuctions: auctypes.Auctions{
					auctypes.CollateralAuction{
						BaseAuction: auctypes.BaseAuction{
							ID:              1,
							Initiator:       "hard",
							Lot:             sdk.NewInt64Coin("ukava", 11874430),
							Bidder:          nil,
							Bid:             sdk.NewInt64Coin("bnb", 0),
							HasReceivedBids: false,
							EndTime:         endTime,
							MaxEndTime:      endTime,
						},
						CorrespondingDebt: sdk.NewInt64Coin("debt", 0),
						MaxBid:            sdk.NewInt64Coin("bnb", 200003287),
						LotReturns:        lotReturns,
					},
					auctypes.CollateralAuction{
						BaseAuction: auctypes.BaseAuction{
							ID:              2,
							Initiator:       "hard",
							Lot:             sdk.NewInt64Coin("ukava", 11874254),
							Bidder:          nil,
							Bid:             sdk.NewInt64Coin("btc", 0),
							HasReceivedBids: false,
							EndTime:         endTime,
							MaxEndTime:      endTime,
						},
						CorrespondingDebt: sdk.NewInt64Coin("debt", 0),
						MaxBid:            sdk.NewInt64Coin("btc", 20000032),
						LotReturns:        lotReturns,
					},
					auctypes.CollateralAuction{
						BaseAuction: auctypes.BaseAuction{
							ID:              3,
							Initiator:       "hard",
							Lot:             sdk.NewInt64Coin("ukava", 11875163),
							Bidder:          nil,
							Bid:             sdk.NewInt64Coin("ukava", 0),
							HasReceivedBids: false,
							EndTime:         endTime,
							MaxEndTime:      endTime,
						},
						CorrespondingDebt: sdk.NewInt64Coin("debt", 0),
						MaxBid:            sdk.NewInt64Coin("ukava", 10000782),
						LotReturns:        lotReturns,
					},
					auctypes.CollateralAuction{
						BaseAuction: auctypes.BaseAuction{
							ID:              4,
							Initiator:       "hard",
							Lot:             sdk.NewInt64Coin("ukava", 11876185),
							Bidder:          nil,
							Bid:             sdk.NewInt64Coin("usdc", 0),
							HasReceivedBids: false,
							EndTime:         endTime,
							MaxEndTime:      endTime,
						},
						CorrespondingDebt: sdk.NewInt64Coin("debt", 0),
						MaxBid:            sdk.NewInt64Coin("usdc", 20003284),
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
			"valid: multiple deposits, single borrow",
			args{
				borrower:             borrower,
				keeper:               keeper,
				keeperRewardPercent:  sdk.MustNewDecFromStr("0.05"),
				initialModuleCoins:   sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1000*KAVA_CF))),
				initialBorrowerCoins: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF)), sdk.NewCoin("bnb", sdk.NewInt(100*BNB_CF)), sdk.NewCoin("btc", sdk.NewInt(100*BTCB_CF))),
				initialKeeperCoins:   sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))),
				depositCoins:         sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(50*KAVA_CF)), sdk.NewCoin("bnb", sdk.NewInt(10*BNB_CF)), sdk.NewCoin("btc", sdk.NewInt(1*BTCB_CF))), // $100 + $100 + $100 = $300 * 0.8 = $240 borrowable                                                                                                                                       // $100 * 0.8 = $80 borrowable
				borrowCoins:          sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(120*KAVA_CF))),                                                                                      // $240 borrowed
				liquidateAfter:       oneMonthInSeconds,
				expectedTotalSuppliedCoins: sdk.NewCoins(
					sdk.NewInt64Coin("btc", 5000000),
					sdk.NewInt64Coin("bnb", 50000000),
					sdk.NewInt64Coin("ukava", 2601709),
				),
				expectedTotalBorrowedCoins: nil,
				expectedKeeperCoins:        sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(102500253)), sdk.NewCoin("bnb", sdk.NewInt(0.5*BNB_CF)), sdk.NewCoin("btc", sdk.NewInt(0.05*BTCB_CF))), // 5% of each seized coin + initial balances
				expectedBorrowerCoins:      sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(170.000001*KAVA_CF)), sdk.NewCoin("bnb", sdk.NewInt(90*BNB_CF)), sdk.NewCoin("btc", sdk.NewInt(99*BTCB_CF))),
				expectedAuctions: auctypes.Auctions{
					auctypes.CollateralAuction{
						BaseAuction: auctypes.BaseAuction{
							ID:              1,
							Initiator:       "hard",
							Lot:             sdk.NewInt64Coin("bnb", 950000000),
							Bidder:          nil,
							Bid:             sdk.NewInt64Coin("ukava", 0),
							HasReceivedBids: false,
							EndTime:         endTime,
							MaxEndTime:      endTime,
						},
						CorrespondingDebt: sdk.NewInt64Coin("debt", 0),
						MaxBid:            sdk.NewInt64Coin("ukava", 40036023),
						LotReturns:        lotReturns,
					},
					auctypes.CollateralAuction{
						BaseAuction: auctypes.BaseAuction{
							ID:              2,
							Initiator:       "hard",
							Lot:             sdk.NewInt64Coin("btc", 95000000),
							Bidder:          nil,
							Bid:             sdk.NewInt64Coin("ukava", 0),
							HasReceivedBids: false,
							EndTime:         endTime,
							MaxEndTime:      endTime,
						},
						CorrespondingDebt: sdk.NewInt64Coin("debt", 0),
						MaxBid:            sdk.NewInt64Coin("ukava", 40036023),
						LotReturns:        lotReturns,
					},
					auctypes.CollateralAuction{
						BaseAuction: auctypes.BaseAuction{
							ID:              3,
							Initiator:       "hard",
							Lot:             sdk.NewInt64Coin("ukava", 47504818),
							Bidder:          nil,
							Bid:             sdk.NewInt64Coin("ukava", 0),
							HasReceivedBids: false,
							EndTime:         endTime,
							MaxEndTime:      endTime,
						},
						CorrespondingDebt: sdk.NewInt64Coin("debt", 0),
						MaxBid:            sdk.NewInt64Coin("ukava", 40040087),
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
			"valid: mutliple stablecoin deposits, multiple variable coin borrows",
			// Auctions: total lot value = $285 ($300 of deposits - $15 keeper reward), total max bid value = $270
			args{
				borrower:             borrower,
				keeper:               keeper,
				keeperRewardPercent:  sdk.MustNewDecFromStr("0.05"),
				initialModuleCoins:   sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1000*KAVA_CF)), sdk.NewCoin("bnb", sdk.NewInt(1000*BNB_CF)), sdk.NewCoin("btc", sdk.NewInt(1000*BTCB_CF))),
				initialBorrowerCoins: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF)), sdk.NewCoin("usdc", sdk.NewInt(100*KAVA_CF)), sdk.NewCoin("usdt", sdk.NewInt(100*KAVA_CF)), sdk.NewCoin("usdx", sdk.NewInt(100*KAVA_CF))),
				initialKeeperCoins:   sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))),
				depositCoins:         sdk.NewCoins(sdk.NewCoin("usdc", sdk.NewInt(100*KAVA_CF)), sdk.NewCoin("usdt", sdk.NewInt(100*KAVA_CF)), sdk.NewCoin("usdx", sdk.NewInt(100*KAVA_CF))), // $100 + $100 + $100 = $300 * 0.9 = $270 borrowable
				borrowCoins:          sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(35*KAVA_CF)), sdk.NewCoin("bnb", sdk.NewInt(10*BNB_CF)), sdk.NewCoin("btc", sdk.NewInt(1*BTCB_CF))),       // $270 borrowed
				liquidateAfter:       oneMonthInSeconds,
				expectedTotalSuppliedCoins: sdk.NewCoins(
					sdk.NewInt64Coin("bnb", 78047),
					sdk.NewInt64Coin("btc", 780),
					sdk.NewInt64Coin("ukava", 9550),
					sdk.NewInt64Coin("usdc", 5000000),
					sdk.NewInt64Coin("usdt", 5000000),
					sdk.NewInt64Coin("usdx", 5000001),
				),
				expectedTotalBorrowedCoins: nil,
				expectedKeeperCoins:        sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF)), sdk.NewCoin("usdc", sdk.NewInt(5*KAVA_CF)), sdk.NewCoin("usdt", sdk.NewInt(5*KAVA_CF)), sdk.NewCoin("usdx", sdk.NewInt(5*KAVA_CF))), // 5% of each seized coin + initial balances
				expectedBorrowerCoins:      sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(135*KAVA_CF)), sdk.NewCoin("bnb", sdk.NewInt(10*BNB_CF)), sdk.NewCoin("btc", sdk.NewInt(1*BTCB_CF)), sdk.NewCoin("usdx", sdk.NewInt(0.000001*KAVA_CF))),
				expectedAuctions: auctypes.Auctions{
					auctypes.CollateralAuction{
						BaseAuction: auctypes.BaseAuction{
							ID:              1,
							Initiator:       "hard",
							Lot:             sdk.NewInt64Coin("usdc", 95000000), // $95.00
							Bidder:          nil,
							Bid:             sdk.NewInt64Coin("bnb", 0),
							HasReceivedBids: false,
							EndTime:         endTime,
							MaxEndTime:      endTime,
						},
						CorrespondingDebt: sdk.NewInt64Coin("debt", 0),
						MaxBid:            sdk.NewInt64Coin("bnb", 900097134), // $90.00
						LotReturns:        lotReturns,
					},
					auctypes.CollateralAuction{
						BaseAuction: auctypes.BaseAuction{
							ID:              2,
							Initiator:       "hard",
							Lot:             sdk.NewInt64Coin("usdt", 10552835), // $10.55
							Bidder:          nil,
							Bid:             sdk.NewInt64Coin("bnb", 0),
							HasReceivedBids: false,
							EndTime:         endTime,
							MaxEndTime:      endTime,
						},
						CorrespondingDebt: sdk.NewInt64Coin("debt", 0),
						MaxBid:            sdk.NewInt64Coin("bnb", 99985020), // $10.00
						LotReturns:        lotReturns,
					},
					auctypes.CollateralAuction{
						BaseAuction: auctypes.BaseAuction{
							ID:              3,
							Initiator:       "hard",
							Lot:             sdk.NewInt64Coin("usdt", 84447165), // $84.45
							Bidder:          nil,
							Bid:             sdk.NewInt64Coin("btc", 0),
							HasReceivedBids: false,
							EndTime:         endTime,
							MaxEndTime:      endTime,
						},
						CorrespondingDebt: sdk.NewInt64Coin("debt", 0),
						MaxBid:            sdk.NewInt64Coin("btc", 80011211), // $80.01
						LotReturns:        lotReturns,
					},
					auctypes.CollateralAuction{
						BaseAuction: auctypes.BaseAuction{
							ID:              4,
							Initiator:       "hard",
							Lot:             sdk.NewInt64Coin("usdx", 21097866), // $21.10
							Bidder:          nil,
							Bid:             sdk.NewInt64Coin("btc", 0),
							HasReceivedBids: false,
							EndTime:         endTime,
							MaxEndTime:      endTime,
						},
						CorrespondingDebt: sdk.NewInt64Coin("debt", 0),
						MaxBid:            sdk.NewInt64Coin("btc", 19989610), // $19.99
						LotReturns:        lotReturns,
					},
					auctypes.CollateralAuction{
						BaseAuction: auctypes.BaseAuction{
							ID:              5,
							Initiator:       "hard",
							Lot:             sdk.NewInt64Coin("usdx", 73902133), //$73.90
							Bidder:          nil,
							Bid:             sdk.NewInt64Coin("ukava", 0),
							HasReceivedBids: false,
							EndTime:         endTime,
							MaxEndTime:      endTime,
						},
						CorrespondingDebt: sdk.NewInt64Coin("debt", 0),
						MaxBid:            sdk.NewInt64Coin("ukava", 35010052), // $70.02
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
			"valid: multiple stablecoin deposits, multiple stablecoin borrows",
			args{
				borrower:             borrower,
				keeper:               keeper,
				keeperRewardPercent:  sdk.MustNewDecFromStr("0.05"),
				initialModuleCoins:   sdk.NewCoins(sdk.NewCoin("usdx", sdk.NewInt(1000*KAVA_CF)), sdk.NewCoin("usdt", sdk.NewInt(1000*KAVA_CF)), sdk.NewCoin("dai", sdk.NewInt(1000*KAVA_CF)), sdk.NewCoin("usdc", sdk.NewInt(1000*KAVA_CF))),
				initialBorrowerCoins: sdk.NewCoins(sdk.NewCoin("usdx", sdk.NewInt(1000*KAVA_CF)), sdk.NewCoin("usdt", sdk.NewInt(1000*KAVA_CF)), sdk.NewCoin("dai", sdk.NewInt(1000*KAVA_CF)), sdk.NewCoin("usdc", sdk.NewInt(1000*KAVA_CF))),
				initialKeeperCoins:   sdk.NewCoins(sdk.NewCoin("usdx", sdk.NewInt(1000*KAVA_CF)), sdk.NewCoin("usdt", sdk.NewInt(1000*KAVA_CF)), sdk.NewCoin("dai", sdk.NewInt(1000*KAVA_CF)), sdk.NewCoin("usdc", sdk.NewInt(1000*KAVA_CF))),
				depositCoins:         sdk.NewCoins(sdk.NewCoin("dai", sdk.NewInt(350*KAVA_CF)), sdk.NewCoin("usdc", sdk.NewInt(200*KAVA_CF))),
				borrowCoins:          sdk.NewCoins(sdk.NewCoin("usdt", sdk.NewInt(250*KAVA_CF)), sdk.NewCoin("usdx", sdk.NewInt(245*KAVA_CF))),
				liquidateAfter:       oneMonthInSeconds,
				expectedTotalSuppliedCoins: sdk.NewCoins(
					sdk.NewInt64Coin("dai", 17500000),
					sdk.NewInt64Coin("usdc", 10000001),
					sdk.NewInt64Coin("usdt", 482503),
					sdk.NewInt64Coin("usdx", 463500),
				),
				expectedTotalBorrowedCoins: nil,
				expectedKeeperCoins:        sdk.NewCoins(sdk.NewCoin("dai", sdk.NewInt(1017.50*KAVA_CF)), sdk.NewCoin("usdt", sdk.NewInt(1000*KAVA_CF)), sdk.NewCoin("usdc", sdk.NewInt(1010*KAVA_CF)), sdk.NewCoin("usdx", sdk.NewInt(1000*KAVA_CF))),
				expectedBorrowerCoins:      sdk.NewCoins(sdk.NewCoin("dai", sdk.NewInt(650*KAVA_CF)), sdk.NewCoin("usdc", sdk.NewInt(800000001)), sdk.NewCoin("usdt", sdk.NewInt(1250*KAVA_CF)), sdk.NewCoin("usdx", sdk.NewInt(1245*KAVA_CF))),
				expectedAuctions: auctypes.Auctions{
					auctypes.CollateralAuction{
						BaseAuction: auctypes.BaseAuction{
							ID:              1,
							Initiator:       "hard",
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
							Initiator:       "hard",
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
							Initiator:       "hard",
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
		{
			"invalid: borrow not liquidatable",
			args{
				borrower:                   borrower,
				keeper:                     keeper,
				keeperRewardPercent:        sdk.MustNewDecFromStr("0.05"),
				initialModuleCoins:         sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))),
				initialBorrowerCoins:       sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))),
				initialKeeperCoins:         sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))),
				depositCoins:               sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(20*KAVA_CF))), // Deposit 20 KAVA
				borrowCoins:                sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(5*KAVA_CF))),  // Borrow 5 KAVA
				liquidateAfter:             oneMonthInSeconds,
				expectedTotalSuppliedCoins: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(20001624))),
				expectedTotalBorrowedCoins: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(5001709))),
				expectedKeeperCoins:        sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100.5*KAVA_CF))),
				expectedBorrowerCoins:      sdk.NewCoins(),
				expectedAuctions:           auctypes.Auctions{},
			},
			errArgs{
				expectPass: false,
				contains:   "borrow not liquidatable",
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

			// Hard module genesis state
			hardGS := types.NewGenesisState(types.NewParams(
				types.MoneyMarkets{
					types.NewMoneyMarket("usdx",
						types.NewBorrowLimit(false, sdk.NewDec(100000000*KAVA_CF), sdk.MustNewDecFromStr("0.9")), // Borrow Limit
						"usdx:usd",                   // Market ID
						sdk.NewInt(KAVA_CF),          // Conversion Factor
						model,                        // Interest Rate Model
						reserveFactor,                // Reserve Factor
						tc.args.keeperRewardPercent), // Keeper Reward Percent
					types.NewMoneyMarket("usdt",
						types.NewBorrowLimit(false, sdk.NewDec(100000000*KAVA_CF), sdk.MustNewDecFromStr("0.9")), // Borrow Limit
						"usdt:usd",                   // Market ID
						sdk.NewInt(KAVA_CF),          // Conversion Factor
						model,                        // Interest Rate Model
						reserveFactor,                // Reserve Factor
						tc.args.keeperRewardPercent), // Keeper Reward Percent
					types.NewMoneyMarket("usdc",
						types.NewBorrowLimit(false, sdk.NewDec(100000000*KAVA_CF), sdk.MustNewDecFromStr("0.9")), // Borrow Limit
						"usdc:usd",                   // Market ID
						sdk.NewInt(KAVA_CF),          // Conversion Factor
						model,                        // Interest Rate Model
						reserveFactor,                // Reserve Factor
						tc.args.keeperRewardPercent), // Keeper Reward Percent
					types.NewMoneyMarket("dai",
						types.NewBorrowLimit(false, sdk.NewDec(100000000*KAVA_CF), sdk.MustNewDecFromStr("0.9")), // Borrow Limit
						"dai:usd",                    // Market ID
						sdk.NewInt(KAVA_CF),          // Conversion Factor
						model,                        // Interest Rate Model
						reserveFactor,                // Reserve Factor
						tc.args.keeperRewardPercent), // Keeper Reward Percent
					types.NewMoneyMarket("ukava",
						types.NewBorrowLimit(false, sdk.NewDec(100000000*KAVA_CF), sdk.MustNewDecFromStr("0.8")), // Borrow Limit
						"kava:usd",                   // Market ID
						sdk.NewInt(KAVA_CF),          // Conversion Factor
						model,                        // Interest Rate Model
						reserveFactor,                // Reserve Factor
						tc.args.keeperRewardPercent), // Keeper Reward Percent
					types.NewMoneyMarket("bnb",
						types.NewBorrowLimit(false, sdk.NewDec(100000000*BNB_CF), sdk.MustNewDecFromStr("0.8")), // Borrow Limit
						"bnb:usd",                    // Market ID
						sdk.NewInt(BNB_CF),           // Conversion Factor
						model,                        // Interest Rate Model
						reserveFactor,                // Reserve Factor
						tc.args.keeperRewardPercent), // Keeper Reward Percent
					types.NewMoneyMarket("btc",
						types.NewBorrowLimit(false, sdk.NewDec(100000000*BTCB_CF), sdk.MustNewDecFromStr("0.8")), // Borrow Limit
						"btc:usd",                    // Market ID
						sdk.NewInt(BTCB_CF),          // Conversion Factor
						model,                        // Interest Rate Model
						reserveFactor,                // Reserve Factor
						tc.args.keeperRewardPercent), // Keeper Reward Percent
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
						{MarketID: "usdt:usd", BaseAsset: "usdt", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
						{MarketID: "usdc:usd", BaseAsset: "usdc", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
						{MarketID: "dai:usd", BaseAsset: "dai", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
						{MarketID: "kava:usd", BaseAsset: "kava", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
						{MarketID: "bnb:usd", BaseAsset: "bnb", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
						{MarketID: "btc:usd", BaseAsset: "btc", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
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
					{
						MarketID:      "bnb:usd",
						OracleAddress: sdk.AccAddress{},
						Price:         sdk.MustNewDecFromStr("10.00"),
						Expiry:        time.Now().Add(100 * time.Hour),
					},
					{
						MarketID:      "btc:usd",
						OracleAddress: sdk.AccAddress{},
						Price:         sdk.MustNewDecFromStr("100.00"),
						Expiry:        time.Now().Add(100 * time.Hour),
					},
				},
			}

			// Initialize test application
			tApp.InitializeFromGenesisStates(authGS,
				app.GenesisState{pricefeed.ModuleName: pricefeed.ModuleCdc.MustMarshalJSON(pricefeedGS)},
				app.GenesisState{types.ModuleName: types.ModuleCdc.MustMarshalJSON(hardGS)})

			// Mint coins to Hard module account
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

			// Deposit coins
			err = suite.keeper.Deposit(suite.ctx, tc.args.borrower, tc.args.depositCoins)
			suite.Require().NoError(err)

			// Borrow coins
			err = suite.keeper.Borrow(suite.ctx, tc.args.borrower, tc.args.borrowCoins)
			suite.Require().NoError(err)

			// Set up liquidation chain context and run begin blocker
			runAtTime := time.Unix(suite.ctx.BlockTime().Unix()+(tc.args.liquidateAfter), 0)
			liqCtx := suite.ctx.WithBlockTime(runAtTime)
			hard.BeginBlocker(liqCtx, suite.keeper)

			// Check borrow exists before liquidation
			_, foundBorrowBefore := suite.keeper.GetBorrow(liqCtx, tc.args.borrower)
			suite.Require().True(foundBorrowBefore)
			// Check that the user's deposit exists before liquidation
			_, foundDepositBefore := suite.keeper.GetDeposit(liqCtx, tc.args.borrower)
			suite.Require().True(foundDepositBefore)

			// Fetch supplied and borrowed coins pre-liquidation
			suppliedCoinsPre, foundSuppliedCoinsPre := suite.keeper.GetSuppliedCoins(liqCtx)
			suite.Require().True(foundSuppliedCoinsPre)
			borrowedCoinsPre, foundBorrowedCoinsPre := suite.keeper.GetBorrowedCoins(liqCtx)
			suite.Require().True(foundBorrowedCoinsPre)

			// Attempt to liquidate
			err = suite.keeper.AttemptKeeperLiquidation(liqCtx, tc.args.keeper, tc.args.borrower)
			if tc.errArgs.expectPass {
				suite.Require().NoError(err)

				// Check borrow does not exist after liquidation
				_, foundBorrowAfter := suite.keeper.GetBorrow(liqCtx, tc.args.borrower)
				suite.Require().False(foundBorrowAfter)
				// Check deposits do not exist after liquidation
				_, foundDepositAfter := suite.keeper.GetDeposit(liqCtx, tc.args.borrower)
				suite.Require().False(foundDepositAfter)

				// Check that the keeper's balance increased by reward % of all the borrowed coins
				accKeeper := suite.getAccountAtCtx(tc.args.keeper, liqCtx)
				suite.Require().Equal(tc.args.expectedKeeperCoins, accKeeper.GetCoins())

				// Check that borrower's balance contains the expected coins
				accBorrower := suite.getAccountAtCtx(tc.args.borrower, liqCtx)
				suite.Require().Equal(tc.args.expectedBorrowerCoins, accBorrower.GetCoins())

				// Check that the expected auctions have been created
				auctions := suite.auctionKeeper.GetAllAuctions(liqCtx)
				suite.Require().True(len(auctions) > 0)
				suite.Require().Equal(tc.args.expectedAuctions, auctions)

				// Check that supplied and borrowed coins have been updated post-liquidation
				suppliedCoinsPost, _ := suite.keeper.GetSuppliedCoins(liqCtx)
				if !suppliedCoinsPost.IsEqual(tc.args.expectedTotalSuppliedCoins) {
				}
				suite.Require().Equal(tc.args.expectedTotalSuppliedCoins, suppliedCoinsPost)
				borrowedCoinsPost, _ := suite.keeper.GetBorrowedCoins(liqCtx)
				suite.Require().Equal(tc.args.expectedTotalBorrowedCoins, borrowedCoinsPost)
			} else {
				suite.Require().Error(err)
				suite.Require().True(strings.Contains(err.Error(), tc.errArgs.contains))

				// Check that the user's borrow exists
				_, foundBorrowAfter := suite.keeper.GetBorrow(liqCtx, tc.args.borrower)
				suite.Require().True(foundBorrowAfter)
				// Check that the user's deposits exist
				_, foundDepositAfter := suite.keeper.GetDeposit(liqCtx, tc.args.borrower)
				suite.Require().True(foundDepositAfter)

				// Check that no auctions have been created
				auctions := suite.auctionKeeper.GetAllAuctions(liqCtx)
				suite.Require().True(len(auctions) == 0)

				// Check that supplied and borrowed coins have not been updated post-liquidation
				suite.Require().Equal(tc.args.expectedTotalSuppliedCoins, suppliedCoinsPre)
				suite.Require().Equal(tc.args.expectedTotalBorrowedCoins, borrowedCoinsPre)
			}
		})
	}
}
