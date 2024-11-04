//nolint:lll // these tests have some long lines :D
package keeper_test

import (
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cometbft/cometbft/crypto"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtime "github.com/cometbft/cometbft/types/time"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/hard"
	"github.com/kava-labs/kava/x/hard/keeper"
	"github.com/kava-labs/kava/x/hard/types"
	pricefeedtypes "github.com/kava-labs/kava/x/pricefeed/types"
)

const (
	USDX_CF = 1000000
	KAVA_CF = 1000000
	BTCB_CF = 100000000
	BNB_CF  = 100000000
	BUSD_CF = 100000000
)

func (suite *KeeperTestSuite) TestBorrow() {
	type setupArgs struct {
		usdxBorrowLimit    sdk.Dec
		priceKAVA          sdk.Dec
		loanToValueKAVA    sdk.Dec
		priceBTCB          sdk.Dec
		loanToValueBTCB    sdk.Dec
		priceBNB           sdk.Dec
		loanToValueBNB     sdk.Dec
		borrower           sdk.AccAddress
		depositCoins       []sdk.Coin
		initialBorrowCoins sdk.Coins
	}

	type borrowArgs struct {
		borrowCoins sdk.Coins
		wantErr     string
	}
	type expected struct {
		expectedAccountBalance    sdk.Coins
		expectedModAccountBalance sdk.Coins

		expectPass bool
		contains   string
	}

	type borrowTest struct {
		name     string
		setup    setupArgs
		borrows  []borrowArgs
		expected []expected
	}

	testCases := []borrowTest{
		{
			name: "valid",
			setup: setupArgs{
				usdxBorrowLimit:    sdk.MustNewDecFromStr("100000000000"),
				priceKAVA:          sdk.MustNewDecFromStr("5.00"),
				loanToValueKAVA:    sdk.MustNewDecFromStr("0.6"),
				priceBTCB:          sdk.MustNewDecFromStr("0.00"),
				loanToValueBTCB:    sdk.MustNewDecFromStr("0.01"),
				priceBNB:           sdk.MustNewDecFromStr("0.00"),
				loanToValueBNB:     sdk.MustNewDecFromStr("0.01"),
				borrower:           sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				depositCoins:       []sdk.Coin{sdk.NewCoin("ukava", sdkmath.NewInt(100*KAVA_CF))},
				initialBorrowCoins: sdk.NewCoins(),
			},
			expected: []expected{
				{
					expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(20*KAVA_CF)), sdk.NewCoin("btcb", sdkmath.NewInt(100*BTCB_CF)), sdk.NewCoin("bnb", sdkmath.NewInt(100*BNB_CF)), sdk.NewCoin("xyz", sdkmath.NewInt(1))),
					expectedModAccountBalance: sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(1080*KAVA_CF)), sdk.NewCoin("usdx", sdkmath.NewInt(200*USDX_CF)), sdk.NewCoin("busd", sdkmath.NewInt(100*BUSD_CF))),

					expectPass: true,
				},
			},
			borrows: []borrowArgs{
				{
					borrowCoins: sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(20*KAVA_CF))),
					wantErr:     "",
				},
			},
		},
		{
			name: "invalid: loan-to-value limited",
			setup: setupArgs{
				usdxBorrowLimit:    sdk.MustNewDecFromStr("100000000000"),
				priceKAVA:          sdk.MustNewDecFromStr("5.00"),
				loanToValueKAVA:    sdk.MustNewDecFromStr("0.6"),
				priceBTCB:          sdk.MustNewDecFromStr("0.00"),
				loanToValueBTCB:    sdk.MustNewDecFromStr("0.01"),
				priceBNB:           sdk.MustNewDecFromStr("0.00"),
				loanToValueBNB:     sdk.MustNewDecFromStr("0.01"),
				borrower:           sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				depositCoins:       []sdk.Coin{sdk.NewCoin("ukava", sdkmath.NewInt(20*KAVA_CF))}, // 20 KAVA x $5.00 price = $100
				initialBorrowCoins: sdk.NewCoins(),
			},
			expected: []expected{
				{
					expectedAccountBalance:    sdk.NewCoins(),
					expectedModAccountBalance: sdk.NewCoins(),

					expectPass: false,
					contains:   "exceeds the allowable amount as determined by the collateralization ratio",
				},
			},
			borrows: []borrowArgs{
				{
					borrowCoins: sdk.NewCoins(sdk.NewCoin("usdx", sdkmath.NewInt(61*USDX_CF))), // 61 USDX x $1 price = $61
					wantErr:     "",
				},
			},
		},
		{
			name: "valid: multiple deposits",
			setup: setupArgs{
				usdxBorrowLimit:    sdk.MustNewDecFromStr("100000000000"),
				priceKAVA:          sdk.MustNewDecFromStr("2.00"),
				loanToValueKAVA:    sdk.MustNewDecFromStr("0.80"),
				priceBTCB:          sdk.MustNewDecFromStr("10000.00"),
				loanToValueBTCB:    sdk.MustNewDecFromStr("0.10"),
				priceBNB:           sdk.MustNewDecFromStr("0.00"),
				loanToValueBNB:     sdk.MustNewDecFromStr("0.01"),
				borrower:           sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				depositCoins:       sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(50*KAVA_CF)), sdk.NewCoin("btcb", sdkmath.NewInt(0.1*BTCB_CF))),
				initialBorrowCoins: sdk.NewCoins(),
			},
			expected: []expected{
				{
					expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(50*KAVA_CF)), sdk.NewCoin("btcb", sdkmath.NewInt(99.9*BTCB_CF)), sdk.NewCoin("usdx", sdkmath.NewInt(180*USDX_CF)), sdk.NewCoin("bnb", sdkmath.NewInt(100*BNB_CF)), sdk.NewCoin("xyz", sdkmath.NewInt(1))),
					expectedModAccountBalance: sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(1050*KAVA_CF)), sdk.NewCoin("usdx", sdkmath.NewInt(20*USDX_CF)), sdk.NewCoin("btcb", sdkmath.NewInt(0.1*BTCB_CF)), sdk.NewCoin("busd", sdkmath.NewInt(100*BUSD_CF))),

					expectPass: true,

					contains: "exceeds the allowable amount as determined by the collateralization ratio",
				},
			},
			borrows: []borrowArgs{
				{
					borrowCoins: sdk.NewCoins(sdk.NewCoin("usdx", sdkmath.NewInt(180*USDX_CF))),
					wantErr:     "",
				},
			},
		},
		{
			name: "invalid: multiple deposits",
			setup: setupArgs{
				usdxBorrowLimit:    sdk.MustNewDecFromStr("100000000000"),
				priceKAVA:          sdk.MustNewDecFromStr("2.00"),
				loanToValueKAVA:    sdk.MustNewDecFromStr("0.80"),
				priceBTCB:          sdk.MustNewDecFromStr("10000.00"),
				loanToValueBTCB:    sdk.MustNewDecFromStr("0.10"),
				priceBNB:           sdk.MustNewDecFromStr("0.00"),
				loanToValueBNB:     sdk.MustNewDecFromStr("0.01"),
				borrower:           sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				depositCoins:       sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(50*KAVA_CF)), sdk.NewCoin("btcb", sdkmath.NewInt(0.1*BTCB_CF))),
				initialBorrowCoins: sdk.NewCoins(),
			},
			expected: []expected{
				{
					expectedAccountBalance:    sdk.NewCoins(),
					expectedModAccountBalance: sdk.NewCoins(),

					expectPass: false,
					contains:   "exceeds the allowable amount as determined by the collateralization ratio",
				},
			},
			borrows: []borrowArgs{
				{
					borrowCoins: sdk.NewCoins(sdk.NewCoin("usdx", sdkmath.NewInt(181*USDX_CF))),
					wantErr:     "",
				},
			},
		},
		{
			name: "valid: multiple previous borrows",
			setup: setupArgs{
				usdxBorrowLimit:    sdk.MustNewDecFromStr("100000000000"),
				priceKAVA:          sdk.MustNewDecFromStr("2.00"),
				loanToValueKAVA:    sdk.MustNewDecFromStr("0.8"),
				priceBTCB:          sdk.MustNewDecFromStr("0.00"),
				loanToValueBTCB:    sdk.MustNewDecFromStr("0.01"),
				priceBNB:           sdk.MustNewDecFromStr("5.00"),
				loanToValueBNB:     sdk.MustNewDecFromStr("0.8"),
				borrower:           sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				depositCoins:       sdk.NewCoins(sdk.NewCoin("bnb", sdkmath.NewInt(30*BNB_CF)), sdk.NewCoin("ukava", sdkmath.NewInt(50*KAVA_CF))), // (50 KAVA x $2.00 price = $100) + (30 BNB x $5.00 price = $150) = $250
				initialBorrowCoins: sdk.NewCoins(sdk.NewCoin("usdx", sdkmath.NewInt(99*USDX_CF)), sdk.NewCoin("busd", sdkmath.NewInt(100*BUSD_CF))),
			},
			expected: []expected{
				{
					expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(50*KAVA_CF)), sdk.NewCoin("btcb", sdkmath.NewInt(100*BTCB_CF)), sdk.NewCoin("usdx", sdkmath.NewInt(100*USDX_CF)), sdk.NewCoin("busd", sdkmath.NewInt(100*BUSD_CF)), sdk.NewCoin("bnb", sdkmath.NewInt(70*BNB_CF)), sdk.NewCoin("xyz", sdkmath.NewInt(1))),
					expectedModAccountBalance: sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(1050*KAVA_CF)), sdk.NewCoin("bnb", sdkmath.NewInt(30*BUSD_CF)), sdk.NewCoin("usdx", sdkmath.NewInt(100*USDX_CF))),

					expectPass: true,
				},
			},
			borrows: []borrowArgs{
				{
					borrowCoins: sdk.NewCoins(sdk.NewCoin("usdx", sdkmath.NewInt(1*USDX_CF))),
					wantErr:     "",
				},
			},
		},
		{
			name: "invalid: over loan-to-value with multiple previous borrows",
			setup: setupArgs{
				usdxBorrowLimit:    sdk.MustNewDecFromStr("100000000000"),
				priceKAVA:          sdk.MustNewDecFromStr("2.00"),
				loanToValueKAVA:    sdk.MustNewDecFromStr("0.8"),
				priceBTCB:          sdk.MustNewDecFromStr("0.00"),
				loanToValueBTCB:    sdk.MustNewDecFromStr("0.01"),
				priceBNB:           sdk.MustNewDecFromStr("5.00"),
				loanToValueBNB:     sdk.MustNewDecFromStr("0.8"),
				borrower:           sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				depositCoins:       sdk.NewCoins(sdk.NewCoin("bnb", sdkmath.NewInt(30*BNB_CF)), sdk.NewCoin("ukava", sdkmath.NewInt(50*KAVA_CF))), // (50 KAVA x $2.00 price = $100) + (30 BNB x $5.00 price = $150) = $250
				initialBorrowCoins: sdk.NewCoins(sdk.NewCoin("usdx", sdkmath.NewInt(100*USDX_CF)), sdk.NewCoin("busd", sdkmath.NewInt(100*BUSD_CF))),
			},
			expected: []expected{
				{
					expectedAccountBalance:    sdk.NewCoins(),
					expectedModAccountBalance: sdk.NewCoins(),
					expectPass:                false,
					contains:                  "exceeds the allowable amount as determined by the collateralization ratio",
				},
			},
			borrows: []borrowArgs{
				{
					borrowCoins: sdk.NewCoins(sdk.NewCoin("usdx", sdkmath.NewInt(1*USDX_CF))),
					wantErr:     "",
				},
			},
		},
		{
			name: "invalid: no price for asset",
			setup: setupArgs{
				usdxBorrowLimit:    sdk.MustNewDecFromStr("100000000000"),
				priceKAVA:          sdk.MustNewDecFromStr("5.00"),
				loanToValueKAVA:    sdk.MustNewDecFromStr("0.6"),
				priceBTCB:          sdk.MustNewDecFromStr("0.00"),
				loanToValueBTCB:    sdk.MustNewDecFromStr("0.01"),
				priceBNB:           sdk.MustNewDecFromStr("0.00"),
				loanToValueBNB:     sdk.MustNewDecFromStr("0.01"),
				borrower:           sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				depositCoins:       sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(100*KAVA_CF))),
				initialBorrowCoins: sdk.NewCoins(),
			},
			expected: []expected{
				{
					expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(20*KAVA_CF)), sdk.NewCoin("btcb", sdkmath.NewInt(100*BTCB_CF)), sdk.NewCoin("bnb", sdkmath.NewInt(100*BNB_CF)), sdk.NewCoin("xyz", sdkmath.NewInt(1))),
					expectedModAccountBalance: sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(1080*KAVA_CF)), sdk.NewCoin("usdx", sdkmath.NewInt(200*USDX_CF)), sdk.NewCoin("busd", sdkmath.NewInt(100*BUSD_CF))),

					expectPass: false,
					contains:   "no price found for market",
				},
			},
			borrows: []borrowArgs{
				{
					borrowCoins: sdk.NewCoins(sdk.NewCoin("xyz", sdkmath.NewInt(1))),
					wantErr:     "",
				},
			},
		},
		{
			name: "invalid: borrow exceed module account balance",
			setup: setupArgs{
				usdxBorrowLimit:    sdk.MustNewDecFromStr("100000000000"),
				priceKAVA:          sdk.MustNewDecFromStr("2.00"),
				loanToValueKAVA:    sdk.MustNewDecFromStr("0.8"),
				priceBTCB:          sdk.MustNewDecFromStr("0.00"),
				loanToValueBTCB:    sdk.MustNewDecFromStr("0.01"),
				priceBNB:           sdk.MustNewDecFromStr("0.00"),
				loanToValueBNB:     sdk.MustNewDecFromStr("0.01"),
				borrower:           sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				depositCoins:       sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(100*KAVA_CF))),
				initialBorrowCoins: sdk.NewCoins(),
			},
			expected: []expected{
				{
					expectedAccountBalance:    sdk.NewCoins(),
					expectedModAccountBalance: sdk.NewCoins(),
					expectPass:                false,
					contains:                  "exceeds borrowable module account balance",
				},
			},
			borrows: []borrowArgs{
				{
					borrowCoins: sdk.NewCoins(sdk.NewCoin("busd", sdkmath.NewInt(101*BUSD_CF))),
					wantErr:     "",
				},
			},
		},
		{
			name: "invalid: over global asset borrow limit",
			setup: setupArgs{
				usdxBorrowLimit:    sdk.MustNewDecFromStr("20000000"),
				priceKAVA:          sdk.MustNewDecFromStr("2.00"),
				loanToValueKAVA:    sdk.MustNewDecFromStr("0.8"),
				priceBTCB:          sdk.MustNewDecFromStr("0.00"),
				loanToValueBTCB:    sdk.MustNewDecFromStr("0.01"),
				priceBNB:           sdk.MustNewDecFromStr("0.00"),
				loanToValueBNB:     sdk.MustNewDecFromStr("0.01"),
				borrower:           sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				depositCoins:       sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(50*KAVA_CF))),
				initialBorrowCoins: sdk.NewCoins(),
			},
			expected: []expected{
				{
					expectedAccountBalance:    sdk.NewCoins(),
					expectedModAccountBalance: sdk.NewCoins(),
					expectPass:                false,
					contains:                  "fails global asset borrow limit validation",
				},
			},
			borrows: []borrowArgs{
				{
					borrowCoins: sdk.NewCoins(sdk.NewCoin("usdx", sdkmath.NewInt(25*USDX_CF))),
					wantErr:     "",
				},
			},
		},
		{
			name: "invalid: borrowing an individual coin type results in a borrow that's under the minimum USD borrow limit",
			setup: setupArgs{
				usdxBorrowLimit:    sdk.MustNewDecFromStr("20000000"),
				priceKAVA:          sdk.MustNewDecFromStr("2.00"),
				loanToValueKAVA:    sdk.MustNewDecFromStr("0.8"),
				priceBTCB:          sdk.MustNewDecFromStr("0.00"),
				loanToValueBTCB:    sdk.MustNewDecFromStr("0.01"),
				priceBNB:           sdk.MustNewDecFromStr("0.00"),
				loanToValueBNB:     sdk.MustNewDecFromStr("0.01"),
				borrower:           sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				depositCoins:       sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(50*KAVA_CF))),
				initialBorrowCoins: sdk.NewCoins(),
			},
			expected: []expected{
				{
					expectedAccountBalance:    sdk.NewCoins(),
					expectedModAccountBalance: sdk.NewCoins(),
					expectPass:                false,
					contains:                  "below the minimum borrow limit",
				},
			},
			borrows: []borrowArgs{
				{
					borrowCoins: sdk.NewCoins(sdk.NewCoin("usdx", sdkmath.NewInt(5*USDX_CF))),
					wantErr:     "",
				},
			},
		},
		{
			name: "invalid: borrowing multiple coins results in a borrow that's under the minimum USD borrow limit",
			setup: setupArgs{
				usdxBorrowLimit:    sdk.MustNewDecFromStr("20000000"),
				priceKAVA:          sdk.MustNewDecFromStr("2.00"),
				loanToValueKAVA:    sdk.MustNewDecFromStr("0.8"),
				priceBTCB:          sdk.MustNewDecFromStr("0.00"),
				loanToValueBTCB:    sdk.MustNewDecFromStr("0.01"),
				priceBNB:           sdk.MustNewDecFromStr("0.00"),
				loanToValueBNB:     sdk.MustNewDecFromStr("0.01"),
				borrower:           sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				depositCoins:       sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(50*KAVA_CF))),
				initialBorrowCoins: sdk.NewCoins(),
			},
			expected: []expected{
				{
					expectedAccountBalance:    sdk.NewCoins(),
					expectedModAccountBalance: sdk.NewCoins(),
					expectPass:                false,
					contains:                  "below the minimum borrow limit",
				},
			},
			borrows: []borrowArgs{
				{
					borrowCoins: sdk.NewCoins(sdk.NewCoin("usdx", sdkmath.NewInt(5*USDX_CF)), sdk.NewCoin("ukava", sdkmath.NewInt(2*USDX_CF))),
					wantErr:     "",
				},
			},
		},
		{
			name: "valid borrow multiple blocks",
			setup: setupArgs{
				usdxBorrowLimit:    sdk.MustNewDecFromStr("100000000000"),
				priceKAVA:          sdk.MustNewDecFromStr("5.00"),
				loanToValueKAVA:    sdk.MustNewDecFromStr("0.6"),
				priceBTCB:          sdk.MustNewDecFromStr("0.00"),
				loanToValueBTCB:    sdk.MustNewDecFromStr("0.01"),
				priceBNB:           sdk.MustNewDecFromStr("0.00"),
				loanToValueBNB:     sdk.MustNewDecFromStr("0.01"),
				borrower:           sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				depositCoins:       []sdk.Coin{sdk.NewCoin("ukava", sdkmath.NewInt(100*KAVA_CF))},
				initialBorrowCoins: sdk.NewCoins(),
			},
			expected: []expected{
				{
					expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(20*KAVA_CF)), sdk.NewCoin("btcb", sdkmath.NewInt(100*BTCB_CF)), sdk.NewCoin("bnb", sdkmath.NewInt(100*BNB_CF)), sdk.NewCoin("xyz", sdkmath.NewInt(1))),
					expectedModAccountBalance: sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(1080*KAVA_CF)), sdk.NewCoin("usdx", sdkmath.NewInt(200*USDX_CF)), sdk.NewCoin("busd", sdkmath.NewInt(100*BUSD_CF))),

					expectPass: true,
				},
				{
					expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(40*KAVA_CF)), sdk.NewCoin("btcb", sdkmath.NewInt(100*BTCB_CF)), sdk.NewCoin("bnb", sdkmath.NewInt(100*BNB_CF)), sdk.NewCoin("xyz", sdkmath.NewInt(1))),
					expectedModAccountBalance: sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(1060*KAVA_CF)), sdk.NewCoin("usdx", sdkmath.NewInt(200*USDX_CF)), sdk.NewCoin("busd", sdkmath.NewInt(100*BUSD_CF))),
					expectPass:                true,
				},
			},
			borrows: []borrowArgs{
				{
					borrowCoins: sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(20*KAVA_CF))),
					wantErr:     "",
				},
				{
					borrowCoins: sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(20*KAVA_CF))),
					wantErr:     "",
				},
			},
		},
		{
			name: "valid borrow followed by protocol reserves exceed available cash for busd when borrowing from ukava",
			setup: setupArgs{
				usdxBorrowLimit:    sdk.MustNewDecFromStr("100000000000"),
				priceKAVA:          sdk.MustNewDecFromStr("5.00"),
				loanToValueKAVA:    sdk.MustNewDecFromStr("0.8"),
				priceBTCB:          sdk.MustNewDecFromStr("0.00"),
				loanToValueBTCB:    sdk.MustNewDecFromStr("0.01"),
				priceBNB:           sdk.MustNewDecFromStr("5.00"),
				loanToValueBNB:     sdk.MustNewDecFromStr("0.8"),
				borrower:           sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				depositCoins:       sdk.NewCoins(sdk.NewCoin("bnb", sdkmath.NewInt(30*BNB_CF)), sdk.NewCoin("ukava", sdkmath.NewInt(50*KAVA_CF))),
				initialBorrowCoins: sdk.NewCoins(sdk.NewCoin("usdx", sdkmath.NewInt(99*USDX_CF)), sdk.NewCoin("busd", sdkmath.NewInt(100*BUSD_CF))),
			},
			expected: []expected{
				{
					expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(50*KAVA_CF)), sdk.NewCoin("btcb", sdkmath.NewInt(100*BTCB_CF)), sdk.NewCoin("usdx", sdkmath.NewInt(100*USDX_CF)), sdk.NewCoin("busd", sdkmath.NewInt(100*BUSD_CF)), sdk.NewCoin("bnb", sdkmath.NewInt(70*BNB_CF)), sdk.NewCoin("xyz", sdkmath.NewInt(1))),
					expectedModAccountBalance: sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(1050*KAVA_CF)), sdk.NewCoin("bnb", sdkmath.NewInt(30*BUSD_CF)), sdk.NewCoin("usdx", sdkmath.NewInt(100*USDX_CF))),

					expectPass: true,
				},
				{
					expectedAccountBalance: sdk.NewCoins(
						sdk.NewCoin("ukava", sdkmath.NewInt(51*KAVA_CF)), // now should be 1 ukava more
						sdk.NewCoin("btcb", sdkmath.NewInt(100*BTCB_CF)),
						sdk.NewCoin("usdx", sdkmath.NewInt(100*USDX_CF)),
						sdk.NewCoin("busd", sdkmath.NewInt(100*BUSD_CF)),
						sdk.NewCoin("bnb", sdkmath.NewInt(70*BNB_CF)),
						sdk.NewCoin("xyz", sdkmath.NewInt(1)),
					),
					expectedModAccountBalance: sdk.NewCoins(
						sdk.NewCoin("ukava", sdkmath.NewInt(1049*KAVA_CF)), // now should be 1 ukava less
						sdk.NewCoin("bnb", sdkmath.NewInt(30*BUSD_CF)),
						sdk.NewCoin("usdx", sdkmath.NewInt(100*USDX_CF)),
					),
					expectPass: true,
				},
			},
			borrows: []borrowArgs{
				{
					borrowCoins: sdk.NewCoins(sdk.NewCoin("usdx", sdkmath.NewInt(1*USDX_CF))),
					wantErr:     "",
				},
				{
					borrowCoins: sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(1*KAVA_CF))),
					wantErr:     "",
				},
			},
		},
		{
			name: "valid borrow followed by protocol reserves exceed available cash for busd when borrowing from busd",
			setup: setupArgs{
				usdxBorrowLimit:    sdk.MustNewDecFromStr("100000000000"),
				priceKAVA:          sdk.MustNewDecFromStr("2.00"),
				loanToValueKAVA:    sdk.MustNewDecFromStr("0.8"),
				priceBTCB:          sdk.MustNewDecFromStr("0.00"),
				loanToValueBTCB:    sdk.MustNewDecFromStr("0.01"),
				priceBNB:           sdk.MustNewDecFromStr("5.00"),
				loanToValueBNB:     sdk.MustNewDecFromStr("0.8"),
				borrower:           sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				depositCoins:       sdk.NewCoins(sdk.NewCoin("bnb", sdkmath.NewInt(30*BNB_CF)), sdk.NewCoin("ukava", sdkmath.NewInt(50*KAVA_CF))),
				initialBorrowCoins: sdk.NewCoins(sdk.NewCoin("usdx", sdkmath.NewInt(99*USDX_CF)), sdk.NewCoin("busd", sdkmath.NewInt(100*BUSD_CF))),
			},
			expected: []expected{
				{
					expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(50*KAVA_CF)), sdk.NewCoin("btcb", sdkmath.NewInt(100*BTCB_CF)), sdk.NewCoin("usdx", sdkmath.NewInt(100*USDX_CF)), sdk.NewCoin("busd", sdkmath.NewInt(100*BUSD_CF)), sdk.NewCoin("bnb", sdkmath.NewInt(70*BNB_CF)), sdk.NewCoin("xyz", sdkmath.NewInt(1))),
					expectedModAccountBalance: sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(1050*KAVA_CF)), sdk.NewCoin("bnb", sdkmath.NewInt(30*BUSD_CF)), sdk.NewCoin("usdx", sdkmath.NewInt(100*USDX_CF))),

					expectPass: true,
				},
				{
					expectedAccountBalance:    sdk.NewCoins(),
					expectedModAccountBalance: sdk.NewCoins(),
					expectPass:                false,
					contains:                  "insolvency - protocol reserves exceed available cash",
				},
			},
			borrows: []borrowArgs{
				{
					borrowCoins: sdk.NewCoins(sdk.NewCoin("usdx", sdkmath.NewInt(1*USDX_CF))),
					wantErr:     "",
				},
				{
					borrowCoins: sdk.NewCoins(sdk.NewCoin("busd", sdkmath.NewInt(100*BUSD_CF))),
					wantErr:     "",
				},
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Initialize test app and set context
			tApp := app.NewTestApp()
			ctx := tApp.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})

			// Auth module genesis state
			authGS := app.NewFundedGenStateWithCoins(
				tApp.AppCodec(),
				[]sdk.Coins{
					sdk.NewCoins(
						sdk.NewCoin("ukava", sdkmath.NewInt(100*KAVA_CF)),
						sdk.NewCoin("btcb", sdkmath.NewInt(100*BTCB_CF)),
						sdk.NewCoin("bnb", sdkmath.NewInt(100*BNB_CF)),
						sdk.NewCoin("xyz", sdkmath.NewInt(1)),
					),
				},
				[]sdk.AccAddress{tc.setup.borrower},
			)

			// hard module genesis state
			hardGS := types.NewGenesisState(types.NewParams(
				types.MoneyMarkets{
					types.NewMoneyMarket("usdx", types.NewBorrowLimit(true, tc.setup.usdxBorrowLimit, sdk.MustNewDecFromStr("1")), "usdx:usd", sdkmath.NewInt(USDX_CF), types.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec()),
					types.NewMoneyMarket("busd", types.NewBorrowLimit(false, sdk.NewDec(100000000*BUSD_CF), sdk.MustNewDecFromStr("1")), "busd:usd", sdkmath.NewInt(BUSD_CF), types.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec()),
					types.NewMoneyMarket("ukava", types.NewBorrowLimit(false, sdk.NewDec(100000000*KAVA_CF), tc.setup.loanToValueKAVA), "kava:usd", sdkmath.NewInt(KAVA_CF), types.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.99"), sdk.ZeroDec()),
					types.NewMoneyMarket("btcb", types.NewBorrowLimit(false, sdk.NewDec(100000000*BTCB_CF), tc.setup.loanToValueBTCB), "btcb:usd", sdkmath.NewInt(BTCB_CF), types.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec()),
					types.NewMoneyMarket("bnb", types.NewBorrowLimit(false, sdk.NewDec(100000000*BNB_CF), tc.setup.loanToValueBNB), "bnb:usd", sdkmath.NewInt(BNB_CF), types.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec()),
					types.NewMoneyMarket("xyz", types.NewBorrowLimit(false, sdk.NewDec(1), tc.setup.loanToValueBNB), "xyz:usd", sdkmath.NewInt(1), types.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec()),
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
						{MarketID: "busd:usd", BaseAsset: "busd", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
						{MarketID: "kava:usd", BaseAsset: "kava", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
						{MarketID: "btcb:usd", BaseAsset: "btcb", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
						{MarketID: "bnb:usd", BaseAsset: "bnb", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
						{MarketID: "xyz:usd", BaseAsset: "xyz", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
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
						MarketID:      "busd:usd",
						OracleAddress: sdk.AccAddress{},
						Price:         sdk.MustNewDecFromStr("1.00"),
						Expiry:        time.Now().Add(1 * time.Hour),
					},
					{
						MarketID:      "kava:usd",
						OracleAddress: sdk.AccAddress{},
						Price:         tc.setup.priceKAVA,
						Expiry:        time.Now().Add(1 * time.Hour),
					},
					{
						MarketID:      "btcb:usd",
						OracleAddress: sdk.AccAddress{},
						Price:         tc.setup.priceBTCB,
						Expiry:        time.Now().Add(1 * time.Hour),
					},
					{
						MarketID:      "bnb:usd",
						OracleAddress: sdk.AccAddress{},
						Price:         tc.setup.priceBNB,
						Expiry:        time.Now().Add(1 * time.Hour),
					},
				},
			}

			// Initialize test application
			tApp.InitializeFromGenesisStates(authGS,
				app.GenesisState{pricefeedtypes.ModuleName: tApp.AppCodec().MustMarshalJSON(&pricefeedGS)},
				app.GenesisState{types.ModuleName: tApp.AppCodec().MustMarshalJSON(&hardGS)})
			// Mint coins to hard module account
			bankKeeper := tApp.GetBankKeeper()
			hardMaccCoins := sdk.NewCoins(
				sdk.NewCoin("ukava", sdkmath.NewInt(1000*KAVA_CF)),
				sdk.NewCoin("usdx", sdkmath.NewInt(200*USDX_CF)),
				sdk.NewCoin("busd", sdkmath.NewInt(100*BUSD_CF)),
			)
			err := bankKeeper.MintCoins(ctx, types.ModuleAccountName, hardMaccCoins)
			suite.Require().NoError(err)

			keeper := tApp.GetHardKeeper()
			suite.app = tApp
			suite.ctx = ctx
			suite.keeper = keeper
			// Run BeginBlocker once to transition MoneyMarkets
			hard.BeginBlocker(suite.ctx, suite.keeper)
			suite.Require().NoError(suite.keeper.Deposit(suite.ctx, tc.setup.borrower, tc.setup.depositCoins))
			// Execute user's previous borrows
			if err = suite.keeper.Borrow(suite.ctx, tc.setup.borrower, tc.setup.initialBorrowCoins); tc.setup.initialBorrowCoins.IsZero() {
				suite.Require().ErrorContains(err, "cannot borrow zero coins")
			} else {
				suite.Require().NoError(err)
			}

			for i, borrow := range tc.borrows {
				// Now that our state is properly set up, execute the last borrow
				err = suite.keeper.Borrow(suite.ctx, tc.setup.borrower, borrow.borrowCoins)
				if tc.expected[i].expectPass {
					suite.Require().NoError(err)

					// Check borrower balance
					acc := suite.getAccount(tc.setup.borrower)
					suite.Require().Equal(tc.expected[i].expectedAccountBalance, suite.getAccountCoins(acc))

					// Check module account balance
					mAcc := suite.getModuleAccount(types.ModuleAccountName)
					suite.Require().Equal(tc.expected[i].expectedModAccountBalance, suite.getAccountCoins(mAcc))

					// Check that borrow struct is in store
					_, f := suite.keeper.GetBorrow(suite.ctx, tc.setup.borrower)
					suite.Require().True(f)
				} else {
					suite.Require().Error(err)
					suite.Require().ErrorContains(err, tc.expected[i].contains)
				}
				blockDuration := time.Second * 3600 * 30 // long blocks to accumulate larger interest
				runAtTime := suite.ctx.BlockTime().Add(blockDuration)
				suite.ctx = suite.ctx.WithBlockTime(runAtTime)
				// Run BeginBlocker once to transition MoneyMarkets
				hard.BeginBlocker(suite.ctx, suite.keeper)

			}
		})
	}
}

func (suite *KeeperTestSuite) TestValidateBorrow() {
	blockDuration := time.Second * 3600 * 24 // long blocks to accumulate larger interest

	_, addrs := app.GeneratePrivKeyAddressPairs(5)
	borrower := addrs[0]
	initialBorrowerBalance := sdk.NewCoins(
		sdk.NewCoin("ukava", sdkmath.NewInt(1000*KAVA_CF)),
		sdk.NewCoin("usdx", sdkmath.NewInt(1000*KAVA_CF)),
	)

	model := types.NewInterestRateModel(sdk.MustNewDecFromStr("1.0"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10"))

	// Initialize test app and set context
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})

	// Auth module genesis state
	authGS := app.NewFundedGenStateWithSameCoins(
		tApp.AppCodec(),
		initialBorrowerBalance,
		[]sdk.AccAddress{borrower},
	)

	// Hard module genesis state
	hardGS := types.NewGenesisState(
		types.NewParams(
			types.MoneyMarkets{
				types.NewMoneyMarket("usdx",
					types.NewBorrowLimit(false, sdk.NewDec(100000000*USDX_CF), sdk.MustNewDecFromStr("1")), // Borrow Limit
					"usdx:usd",                     // Market ID
					sdkmath.NewInt(USDX_CF),        // Conversion Factor
					model,                          // Interest Rate Model
					sdk.MustNewDecFromStr("1.0"),   // Reserve Factor (high)
					sdk.MustNewDecFromStr("0.05")), // Keeper Reward Percent
				types.NewMoneyMarket("ukava",
					types.NewBorrowLimit(false, sdk.NewDec(100000000*KAVA_CF), sdk.MustNewDecFromStr("0.8")), // Borrow Limit
					"kava:usd",                     // Market ID
					sdkmath.NewInt(KAVA_CF),        // Conversion Factor
					model,                          // Interest Rate Model
					sdk.MustNewDecFromStr("1.0"),   // Reserve Factor (high)
					sdk.MustNewDecFromStr("0.05")), // Keeper Reward Percent
			},
			sdk.NewDec(10),
		),
		types.DefaultAccumulationTimes,
		types.DefaultDeposits,
		types.DefaultBorrows,
		types.DefaultTotalSupplied,
		types.DefaultTotalBorrowed,
		types.DefaultTotalReserves,
	)

	// Pricefeed module genesis state
	pricefeedGS := pricefeedtypes.GenesisState{
		Params: pricefeedtypes.Params{
			Markets: []pricefeedtypes.Market{
				{MarketID: "usdx:usd", BaseAsset: "usdx", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
				{MarketID: "kava:usd", BaseAsset: "kava", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
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
		},
	}

	// Initialize test application
	tApp.InitializeFromGenesisStates(
		authGS,
		app.GenesisState{pricefeedtypes.ModuleName: tApp.AppCodec().MustMarshalJSON(&pricefeedGS)},
		app.GenesisState{types.ModuleName: tApp.AppCodec().MustMarshalJSON(&hardGS)},
	)

	keeper := tApp.GetHardKeeper()
	suite.app = tApp
	suite.ctx = ctx
	suite.keeper = keeper

	var err error

	// Run BeginBlocker once to transition MoneyMarkets
	hard.BeginBlocker(suite.ctx, suite.keeper)

	// Setup borrower with some collateral to borrow against, and some reserve in the protocol.
	depositCoins := sdk.NewCoins(
		sdk.NewCoin("ukava", sdkmath.NewInt(100*KAVA_CF)),
		sdk.NewCoin("usdx", sdkmath.NewInt(100*USDX_CF)),
	)
	err = suite.keeper.Deposit(suite.ctx, borrower, depositCoins)
	suite.Require().NoError(err)

	initialBorrowCoins := sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(70*KAVA_CF)))
	err = suite.keeper.Borrow(suite.ctx, borrower, initialBorrowCoins)
	suite.Require().NoError(err)

	runAtTime := suite.ctx.BlockTime().Add(blockDuration)
	suite.ctx = suite.ctx.WithBlockTime(runAtTime)
	hard.BeginBlocker(suite.ctx, suite.keeper)

	repayCoins := sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(100*KAVA_CF))) // repay everything including accumulated interest
	err = suite.keeper.Repay(suite.ctx, borrower, borrower, repayCoins)
	suite.Require().NoError(err)

	// Get the total borrowable amount from the protocol, taking into account the reserves.
	modAccBalance := suite.getAccountCoins(suite.getModuleAccountAtCtx(types.ModuleAccountName, suite.ctx))
	reserves, found := suite.keeper.GetTotalReserves(suite.ctx)
	suite.Require().True(found)
	availableToBorrow := modAccBalance.Sub(reserves...)

	// Test borrowing one over the available amount (try to borrow from the reserves)
	err = suite.keeper.Borrow(
		suite.ctx,
		borrower,
		sdk.NewCoins(sdk.NewCoin("ukava", availableToBorrow.AmountOf("ukava").Add(sdk.OneInt()))),
	)
	suite.Require().Error(err)

	// Test borrowing exactly the limit
	err = suite.keeper.Borrow(
		suite.ctx,
		borrower,
		sdk.NewCoins(sdk.NewCoin("ukava", availableToBorrow.AmountOf("ukava"))),
	)
	suite.Require().NoError(err)

	// now that it's all that you can borrow, we shouldn't be able to borrow anything
	err = suite.keeper.Borrow(
		suite.ctx,
		borrower,
		sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(2*USDX_CF))),
	)
	suite.Require().Error(err)
	suite.Require().ErrorContains(err, "available to borrow: exceeds module account balance")

	suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(blockDuration))
	hard.BeginBlocker(suite.ctx, suite.keeper)

	// Should error since ukava has become insolvent at this point
	err = suite.keeper.Borrow(
		suite.ctx,
		borrower,
		sdk.NewCoins(sdk.NewCoin("ukava", availableToBorrow.AmountOf("ukava"))),
	)
	suite.Require().Error(err)
	suite.Require().ErrorContains(err, "protocol reserves exceed available cash")

	suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(blockDuration))
	hard.BeginBlocker(suite.ctx, suite.keeper)

	err = suite.keeper.Borrow(
		suite.ctx,
		borrower,
		sdk.NewCoins(sdk.NewCoin("usdx", sdkmath.NewInt(25*USDX_CF))),
	)
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestFilterCoinsByDenoms() {
	type args struct {
		coins         sdk.Coins
		filterByCoins sdk.Coins
	}
	tests := []struct {
		name string
		args args
		want sdk.Coins
	}{
		{
			name: "more coins than filtered coins",
			args: args{
				coins: sdk.NewCoins(
					sdk.NewCoin("ukava", sdkmath.NewInt(1000*KAVA_CF)),
					sdk.NewCoin("usdx", sdkmath.NewInt(200*USDX_CF)),
					sdk.NewCoin("busd", sdkmath.NewInt(100*BUSD_CF)),
				),
				filterByCoins: sdk.NewCoins(
					sdk.NewCoin("usdx", sdkmath.NewInt(25*USDX_CF)),
					sdk.NewCoin("ukava", sdkmath.NewInt(25*KAVA_CF)),
				),
			},
			want: sdk.NewCoins(
				sdk.NewCoin("usdx", sdkmath.NewInt(200*USDX_CF)),
				sdk.NewCoin("ukava", sdkmath.NewInt(1000*KAVA_CF)),
			),
		},
		{
			name: "less coins than filtered coins",
			args: args{
				coins: sdk.NewCoins(
					sdk.NewCoin("ukava", sdkmath.NewInt(1000*KAVA_CF)),
				),
				filterByCoins: sdk.NewCoins(
					sdk.NewCoin("usdx", sdkmath.NewInt(25*USDX_CF)),
					sdk.NewCoin("ukava", sdkmath.NewInt(25*KAVA_CF)),
				),
			},
			want: sdk.NewCoins(
				sdk.NewCoin("ukava", sdkmath.NewInt(1000*KAVA_CF)),
			),
		},
		{
			name: "no filter coins ",
			args: args{
				coins: sdk.NewCoins(
					sdk.NewCoin("ukava", sdkmath.NewInt(1000*KAVA_CF)),
				),
				filterByCoins: sdk.NewCoins(),
			},
			want: sdk.NewCoins(),
		},
		{
			name: "no coins ",
			args: args{
				coins: sdk.NewCoins(),
				filterByCoins: sdk.NewCoins(
					sdk.NewCoin("usdx", sdkmath.NewInt(25*USDX_CF))),
			},
			want: sdk.NewCoins(),
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			got := keeper.FilterCoinsByDenoms(tt.args.coins, tt.args.filterByCoins)
			suite.Require().Equal(tt.want, got)
		})
	}
}
