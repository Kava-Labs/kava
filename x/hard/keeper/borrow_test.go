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
		wantDepositErr            string
	}
	type errArgs struct {
		expectPass bool
		contains   string
	}
	type borrowTest struct {
		name    string
		args    []args
		errArgs []errArgs
	}
	testCases := []borrowTest{
		{
			name: "valid",
			args: []args{
				{
					usdxBorrowLimit:           sdk.MustNewDecFromStr("100000000000"),
					priceKAVA:                 sdk.MustNewDecFromStr("5.00"),
					loanToValueKAVA:           sdk.MustNewDecFromStr("0.6"),
					priceBTCB:                 sdk.MustNewDecFromStr("0.00"),
					loanToValueBTCB:           sdk.MustNewDecFromStr("0.01"),
					priceBNB:                  sdk.MustNewDecFromStr("0.00"),
					loanToValueBNB:            sdk.MustNewDecFromStr("0.01"),
					borrower:                  sdk.AccAddress(crypto.AddressHash([]byte("test"))),
					depositCoins:              []sdk.Coin{sdk.NewCoin("ukava", sdkmath.NewInt(100*KAVA_CF))},
					previousBorrowCoins:       sdk.NewCoins(),
					borrowCoins:               sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(20*KAVA_CF))),
					expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(20*KAVA_CF)), sdk.NewCoin("btcb", sdkmath.NewInt(100*BTCB_CF)), sdk.NewCoin("bnb", sdkmath.NewInt(100*BNB_CF)), sdk.NewCoin("xyz", sdkmath.NewInt(1))),
					expectedModAccountBalance: sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(1080*KAVA_CF)), sdk.NewCoin("usdx", sdkmath.NewInt(200*USDX_CF)), sdk.NewCoin("busd", sdkmath.NewInt(100*BUSD_CF))),
				},
			},
			errArgs: []errArgs{
				{
					expectPass: true,
				},
			},
		},
		{
			name: "invalid: loan-to-value limited",
			args: []args{
				{
					usdxBorrowLimit:           sdk.MustNewDecFromStr("100000000000"),
					priceKAVA:                 sdk.MustNewDecFromStr("5.00"),
					loanToValueKAVA:           sdk.MustNewDecFromStr("0.6"),
					priceBTCB:                 sdk.MustNewDecFromStr("0.00"),
					loanToValueBTCB:           sdk.MustNewDecFromStr("0.01"),
					priceBNB:                  sdk.MustNewDecFromStr("0.00"),
					loanToValueBNB:            sdk.MustNewDecFromStr("0.01"),
					borrower:                  sdk.AccAddress(crypto.AddressHash([]byte("test"))),
					depositCoins:              []sdk.Coin{sdk.NewCoin("ukava", sdkmath.NewInt(20*KAVA_CF))},  // 20 KAVA x $5.00 price = $100
					borrowCoins:               sdk.NewCoins(sdk.NewCoin("usdx", sdkmath.NewInt(61*USDX_CF))), // 61 USDX x $1 price = $61
					expectedAccountBalance:    sdk.NewCoins(),
					expectedModAccountBalance: sdk.NewCoins(),
				},
			},
			errArgs: []errArgs{
				{
					expectPass: false,
					contains:   "exceeds the allowable amount as determined by the collateralization ratio",
				},
			},
		},
		{
			name: "valid: multiple deposits",
			args: []args{
				{
					usdxBorrowLimit:           sdk.MustNewDecFromStr("100000000000"),
					priceKAVA:                 sdk.MustNewDecFromStr("2.00"),
					loanToValueKAVA:           sdk.MustNewDecFromStr("0.80"),
					priceBTCB:                 sdk.MustNewDecFromStr("10000.00"),
					loanToValueBTCB:           sdk.MustNewDecFromStr("0.10"),
					priceBNB:                  sdk.MustNewDecFromStr("0.00"),
					loanToValueBNB:            sdk.MustNewDecFromStr("0.01"),
					borrower:                  sdk.AccAddress(crypto.AddressHash([]byte("test"))),
					depositCoins:              sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(50*KAVA_CF)), sdk.NewCoin("btcb", sdkmath.NewInt(0.1*BTCB_CF))),
					borrowCoins:               sdk.NewCoins(sdk.NewCoin("usdx", sdkmath.NewInt(180*USDX_CF))),
					expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(50*KAVA_CF)), sdk.NewCoin("btcb", sdkmath.NewInt(99.9*BTCB_CF)), sdk.NewCoin("usdx", sdkmath.NewInt(180*USDX_CF)), sdk.NewCoin("bnb", sdkmath.NewInt(100*BNB_CF)), sdk.NewCoin("xyz", sdkmath.NewInt(1))),
					expectedModAccountBalance: sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(1050*KAVA_CF)), sdk.NewCoin("usdx", sdkmath.NewInt(20*USDX_CF)), sdk.NewCoin("btcb", sdkmath.NewInt(0.1*BTCB_CF)), sdk.NewCoin("busd", sdkmath.NewInt(100*BUSD_CF))),
				},
			},
			errArgs: []errArgs{
				{
					expectPass: true,
				},
			},
		},
		{
			name: "invalid: multiple deposits",
			args: []args{
				{
					usdxBorrowLimit:           sdk.MustNewDecFromStr("100000000000"),
					priceKAVA:                 sdk.MustNewDecFromStr("2.00"),
					loanToValueKAVA:           sdk.MustNewDecFromStr("0.80"),
					priceBTCB:                 sdk.MustNewDecFromStr("10000.00"),
					loanToValueBTCB:           sdk.MustNewDecFromStr("0.10"),
					priceBNB:                  sdk.MustNewDecFromStr("0.00"),
					loanToValueBNB:            sdk.MustNewDecFromStr("0.01"),
					borrower:                  sdk.AccAddress(crypto.AddressHash([]byte("test"))),
					depositCoins:              sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(50*KAVA_CF)), sdk.NewCoin("btcb", sdkmath.NewInt(0.1*BTCB_CF))),
					borrowCoins:               sdk.NewCoins(sdk.NewCoin("usdx", sdkmath.NewInt(181*USDX_CF))),
					expectedAccountBalance:    sdk.NewCoins(),
					expectedModAccountBalance: sdk.NewCoins(),
				},
			},
			errArgs: []errArgs{
				{
					expectPass: false,
					contains:   "exceeds the allowable amount as determined by the collateralization ratio",
				},
			},
		},
		{
			name: "valid: multiple previous borrows",
			args: []args{
				{
					usdxBorrowLimit:           sdk.MustNewDecFromStr("100000000000"),
					priceKAVA:                 sdk.MustNewDecFromStr("2.00"),
					loanToValueKAVA:           sdk.MustNewDecFromStr("0.8"),
					priceBTCB:                 sdk.MustNewDecFromStr("0.00"),
					loanToValueBTCB:           sdk.MustNewDecFromStr("0.01"),
					priceBNB:                  sdk.MustNewDecFromStr("5.00"),
					loanToValueBNB:            sdk.MustNewDecFromStr("0.8"),
					borrower:                  sdk.AccAddress(crypto.AddressHash([]byte("test"))),
					depositCoins:              sdk.NewCoins(sdk.NewCoin("bnb", sdkmath.NewInt(30*BNB_CF)), sdk.NewCoin("ukava", sdkmath.NewInt(50*KAVA_CF))), // (50 KAVA x $2.00 price = $100) + (30 BNB x $5.00 price = $150) = $250
					previousBorrowCoins:       sdk.NewCoins(sdk.NewCoin("usdx", sdkmath.NewInt(99*USDX_CF)), sdk.NewCoin("busd", sdkmath.NewInt(100*BUSD_CF))),
					borrowCoins:               sdk.NewCoins(sdk.NewCoin("usdx", sdkmath.NewInt(1*USDX_CF))),
					expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(50*KAVA_CF)), sdk.NewCoin("btcb", sdkmath.NewInt(100*BTCB_CF)), sdk.NewCoin("usdx", sdkmath.NewInt(100*USDX_CF)), sdk.NewCoin("busd", sdkmath.NewInt(100*BUSD_CF)), sdk.NewCoin("bnb", sdkmath.NewInt(70*BNB_CF)), sdk.NewCoin("xyz", sdkmath.NewInt(1))),
					expectedModAccountBalance: sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(1050*KAVA_CF)), sdk.NewCoin("bnb", sdkmath.NewInt(30*BUSD_CF)), sdk.NewCoin("usdx", sdkmath.NewInt(100*USDX_CF))),
				},
			},
			errArgs: []errArgs{
				{
					expectPass: true,
				},
			},
		},
		{
			name: "invalid: over loan-to-value with multiple previous borrows",
			args: []args{
				{
					usdxBorrowLimit:           sdk.MustNewDecFromStr("100000000000"),
					priceKAVA:                 sdk.MustNewDecFromStr("2.00"),
					loanToValueKAVA:           sdk.MustNewDecFromStr("0.8"),
					priceBTCB:                 sdk.MustNewDecFromStr("0.00"),
					loanToValueBTCB:           sdk.MustNewDecFromStr("0.01"),
					priceBNB:                  sdk.MustNewDecFromStr("5.00"),
					loanToValueBNB:            sdk.MustNewDecFromStr("0.8"),
					borrower:                  sdk.AccAddress(crypto.AddressHash([]byte("test"))),
					depositCoins:              sdk.NewCoins(sdk.NewCoin("bnb", sdkmath.NewInt(30*BNB_CF)), sdk.NewCoin("ukava", sdkmath.NewInt(50*KAVA_CF))), // (50 KAVA x $2.00 price = $100) + (30 BNB x $5.00 price = $150) = $250
					previousBorrowCoins:       sdk.NewCoins(sdk.NewCoin("usdx", sdkmath.NewInt(100*USDX_CF)), sdk.NewCoin("busd", sdkmath.NewInt(100*BUSD_CF))),
					borrowCoins:               sdk.NewCoins(sdk.NewCoin("usdx", sdkmath.NewInt(1*USDX_CF))),
					expectedAccountBalance:    sdk.NewCoins(),
					expectedModAccountBalance: sdk.NewCoins(),
				},
			},
			errArgs: []errArgs{
				{
					expectPass: false,
					contains:   "exceeds the allowable amount as determined by the collateralization ratio",
				},
			},
		},
		{
			name: "invalid: no price for asset",
			args: []args{
				{
					usdxBorrowLimit:           sdk.MustNewDecFromStr("100000000000"),
					priceKAVA:                 sdk.MustNewDecFromStr("5.00"),
					loanToValueKAVA:           sdk.MustNewDecFromStr("0.6"),
					priceBTCB:                 sdk.MustNewDecFromStr("0.00"),
					loanToValueBTCB:           sdk.MustNewDecFromStr("0.01"),
					priceBNB:                  sdk.MustNewDecFromStr("0.00"),
					loanToValueBNB:            sdk.MustNewDecFromStr("0.01"),
					borrower:                  sdk.AccAddress(crypto.AddressHash([]byte("test"))),
					depositCoins:              sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(100*KAVA_CF))),
					previousBorrowCoins:       sdk.NewCoins(),
					borrowCoins:               sdk.NewCoins(sdk.NewCoin("xyz", sdkmath.NewInt(1))),
					expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(20*KAVA_CF)), sdk.NewCoin("btcb", sdkmath.NewInt(100*BTCB_CF)), sdk.NewCoin("bnb", sdkmath.NewInt(100*BNB_CF)), sdk.NewCoin("xyz", sdkmath.NewInt(1))),
					expectedModAccountBalance: sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(1080*KAVA_CF)), sdk.NewCoin("usdx", sdkmath.NewInt(200*USDX_CF)), sdk.NewCoin("busd", sdkmath.NewInt(100*BUSD_CF))),
				},
			},
			errArgs: []errArgs{
				{
					expectPass: false,
					contains:   "no price found for market",
				},
			},
		},
		{
			name: "invalid: borrow exceed module account balance",
			args: []args{
				{
					usdxBorrowLimit:           sdk.MustNewDecFromStr("100000000000"),
					priceKAVA:                 sdk.MustNewDecFromStr("2.00"),
					loanToValueKAVA:           sdk.MustNewDecFromStr("0.8"),
					priceBTCB:                 sdk.MustNewDecFromStr("0.00"),
					loanToValueBTCB:           sdk.MustNewDecFromStr("0.01"),
					priceBNB:                  sdk.MustNewDecFromStr("0.00"),
					loanToValueBNB:            sdk.MustNewDecFromStr("0.01"),
					borrower:                  sdk.AccAddress(crypto.AddressHash([]byte("test"))),
					depositCoins:              sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(100*KAVA_CF))),
					previousBorrowCoins:       sdk.NewCoins(),
					borrowCoins:               sdk.NewCoins(sdk.NewCoin("busd", sdkmath.NewInt(101*BUSD_CF))),
					expectedAccountBalance:    sdk.NewCoins(),
					expectedModAccountBalance: sdk.NewCoins(),
				},
			},
			errArgs: []errArgs{
				{
					expectPass: false,
					contains:   "exceeds borrowable module account balance",
				},
			},
		},
		{
			name: "invalid: over global asset borrow limit",
			args: []args{
				{
					usdxBorrowLimit:           sdk.MustNewDecFromStr("20000000"),
					priceKAVA:                 sdk.MustNewDecFromStr("2.00"),
					loanToValueKAVA:           sdk.MustNewDecFromStr("0.8"),
					priceBTCB:                 sdk.MustNewDecFromStr("0.00"),
					loanToValueBTCB:           sdk.MustNewDecFromStr("0.01"),
					priceBNB:                  sdk.MustNewDecFromStr("0.00"),
					loanToValueBNB:            sdk.MustNewDecFromStr("0.01"),
					borrower:                  sdk.AccAddress(crypto.AddressHash([]byte("test"))),
					depositCoins:              sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(50*KAVA_CF))),
					previousBorrowCoins:       sdk.NewCoins(),
					borrowCoins:               sdk.NewCoins(sdk.NewCoin("usdx", sdkmath.NewInt(25*USDX_CF))),
					expectedAccountBalance:    sdk.NewCoins(),
					expectedModAccountBalance: sdk.NewCoins(),
				},
			},
			errArgs: []errArgs{
				{
					expectPass: false,
					contains:   "fails global asset borrow limit validation",
				},
			},
		},
		{
			name: "invalid: borrowing an individual coin type results in a borrow that's under the minimum USD borrow limit",
			args: []args{
				{
					usdxBorrowLimit:           sdk.MustNewDecFromStr("20000000"),
					priceKAVA:                 sdk.MustNewDecFromStr("2.00"),
					loanToValueKAVA:           sdk.MustNewDecFromStr("0.8"),
					priceBTCB:                 sdk.MustNewDecFromStr("0.00"),
					loanToValueBTCB:           sdk.MustNewDecFromStr("0.01"),
					priceBNB:                  sdk.MustNewDecFromStr("0.00"),
					loanToValueBNB:            sdk.MustNewDecFromStr("0.01"),
					borrower:                  sdk.AccAddress(crypto.AddressHash([]byte("test"))),
					depositCoins:              sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(50*KAVA_CF))),
					previousBorrowCoins:       sdk.NewCoins(),
					borrowCoins:               sdk.NewCoins(sdk.NewCoin("usdx", sdkmath.NewInt(5*USDX_CF))),
					expectedAccountBalance:    sdk.NewCoins(),
					expectedModAccountBalance: sdk.NewCoins(),
				},
			},
			errArgs: []errArgs{
				{
					expectPass: false,
					contains:   "below the minimum borrow limit",
				},
			},
		},
		{
			name: "invalid: borrowing multiple coins results in a borrow that's under the minimum USD borrow limit",
			args: []args{
				{
					usdxBorrowLimit:           sdk.MustNewDecFromStr("20000000"),
					priceKAVA:                 sdk.MustNewDecFromStr("2.00"),
					loanToValueKAVA:           sdk.MustNewDecFromStr("0.8"),
					priceBTCB:                 sdk.MustNewDecFromStr("0.00"),
					loanToValueBTCB:           sdk.MustNewDecFromStr("0.01"),
					priceBNB:                  sdk.MustNewDecFromStr("0.00"),
					loanToValueBNB:            sdk.MustNewDecFromStr("0.01"),
					borrower:                  sdk.AccAddress(crypto.AddressHash([]byte("test"))),
					depositCoins:              sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(50*KAVA_CF))),
					previousBorrowCoins:       sdk.NewCoins(),
					borrowCoins:               sdk.NewCoins(sdk.NewCoin("usdx", sdkmath.NewInt(5*USDX_CF)), sdk.NewCoin("ukava", sdkmath.NewInt(2*USDX_CF))),
					expectedAccountBalance:    sdk.NewCoins(),
					expectedModAccountBalance: sdk.NewCoins(),
				},
			},
			errArgs: []errArgs{
				{
					expectPass: false,
					contains:   "below the minimum borrow limit",
				},
			},
		},
		{
			name: "valid borrow multiple blocks",
			args: []args{
				{
					usdxBorrowLimit:           sdk.MustNewDecFromStr("100000000000"),
					priceKAVA:                 sdk.MustNewDecFromStr("5.00"),
					loanToValueKAVA:           sdk.MustNewDecFromStr("0.6"),
					priceBTCB:                 sdk.MustNewDecFromStr("0.00"),
					loanToValueBTCB:           sdk.MustNewDecFromStr("0.01"),
					priceBNB:                  sdk.MustNewDecFromStr("0.00"),
					loanToValueBNB:            sdk.MustNewDecFromStr("0.01"),
					borrower:                  sdk.AccAddress(crypto.AddressHash([]byte("test"))),
					depositCoins:              []sdk.Coin{sdk.NewCoin("ukava", sdkmath.NewInt(100*KAVA_CF))},
					previousBorrowCoins:       sdk.NewCoins(),
					borrowCoins:               sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(20*KAVA_CF))),
					expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(20*KAVA_CF)), sdk.NewCoin("btcb", sdkmath.NewInt(100*BTCB_CF)), sdk.NewCoin("bnb", sdkmath.NewInt(100*BNB_CF)), sdk.NewCoin("xyz", sdkmath.NewInt(1))),
					expectedModAccountBalance: sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(1080*KAVA_CF)), sdk.NewCoin("usdx", sdkmath.NewInt(200*USDX_CF)), sdk.NewCoin("busd", sdkmath.NewInt(100*BUSD_CF))),
				},
				{
					borrower:                  sdk.AccAddress(crypto.AddressHash([]byte("test"))),
					depositCoins:              []sdk.Coin{sdk.NewCoin("ukava", sdkmath.NewInt(10*KAVA_CF))},
					borrowCoins:               sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(20*KAVA_CF))),
					expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(40*KAVA_CF)), sdk.NewCoin("btcb", sdkmath.NewInt(100*BTCB_CF)), sdk.NewCoin("bnb", sdkmath.NewInt(100*BNB_CF)), sdk.NewCoin("xyz", sdkmath.NewInt(1))),
					expectedModAccountBalance: sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(1060*KAVA_CF)), sdk.NewCoin("usdx", sdkmath.NewInt(200*USDX_CF)), sdk.NewCoin("busd", sdkmath.NewInt(100*BUSD_CF))),
				},
			},
			errArgs: []errArgs{
				{
					expectPass: true,
				},
				{
					expectPass: true,
				},
			},
		},
		{
			name: "first valid borrow, second insufficient funds blocks",
			args: []args{
				{
					usdxBorrowLimit:           sdk.MustNewDecFromStr("100000000000"),
					priceKAVA:                 sdk.MustNewDecFromStr("5.00"),
					loanToValueKAVA:           sdk.MustNewDecFromStr("0.6"),
					priceBTCB:                 sdk.MustNewDecFromStr("0.00"),
					loanToValueBTCB:           sdk.MustNewDecFromStr("0.01"),
					priceBNB:                  sdk.MustNewDecFromStr("0.00"),
					loanToValueBNB:            sdk.MustNewDecFromStr("0.01"),
					borrower:                  sdk.AccAddress(crypto.AddressHash([]byte("test"))),
					depositCoins:              []sdk.Coin{sdk.NewCoin("ukava", sdkmath.NewInt(100*KAVA_CF))},
					previousBorrowCoins:       sdk.NewCoins(),
					borrowCoins:               sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(20*KAVA_CF))),
					expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(20*KAVA_CF)), sdk.NewCoin("btcb", sdkmath.NewInt(100*BTCB_CF)), sdk.NewCoin("bnb", sdkmath.NewInt(100*BNB_CF)), sdk.NewCoin("xyz", sdkmath.NewInt(1))),
					expectedModAccountBalance: sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(1080*KAVA_CF)), sdk.NewCoin("usdx", sdkmath.NewInt(200*USDX_CF)), sdk.NewCoin("busd", sdkmath.NewInt(100*BUSD_CF))),
				},
				{
					borrower:                  sdk.AccAddress(crypto.AddressHash([]byte("test"))),
					depositCoins:              []sdk.Coin{sdk.NewCoin("ukava", sdkmath.NewInt(100*KAVA_CF))}, // too much
					borrowCoins:               sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(20*KAVA_CF))),
					expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(40*KAVA_CF)), sdk.NewCoin("btcb", sdkmath.NewInt(100*BTCB_CF)), sdk.NewCoin("bnb", sdkmath.NewInt(100*BNB_CF)), sdk.NewCoin("xyz", sdkmath.NewInt(1))),
					expectedModAccountBalance: sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(1060*KAVA_CF)), sdk.NewCoin("usdx", sdkmath.NewInt(200*USDX_CF)), sdk.NewCoin("busd", sdkmath.NewInt(100*BUSD_CF))),
					wantDepositErr:            "insufficient funds: the requested deposit amount of 100000000ukava exceeds the total available account funds of 20000000ukava: exceeds module account balance",
				},
			},
			errArgs: []errArgs{
				{
					expectPass: true,
				},
				{
					expectPass: true,
				},
			},
		},
		{
			name: "valid borrow followed by protocol reserves exceed available cash for busd when borrowing from ukava(incorrect)",
			args: []args{
				{
					usdxBorrowLimit:           sdk.MustNewDecFromStr("100000000000"),
					priceKAVA:                 sdk.MustNewDecFromStr("5.00"),
					loanToValueKAVA:           sdk.MustNewDecFromStr("0.8"),
					priceBTCB:                 sdk.MustNewDecFromStr("0.00"),
					loanToValueBTCB:           sdk.MustNewDecFromStr("0.01"),
					priceBNB:                  sdk.MustNewDecFromStr("5.00"),
					loanToValueBNB:            sdk.MustNewDecFromStr("0.8"),
					borrower:                  sdk.AccAddress(crypto.AddressHash([]byte("test"))),
					depositCoins:              sdk.NewCoins(sdk.NewCoin("bnb", sdkmath.NewInt(30*BNB_CF)), sdk.NewCoin("ukava", sdkmath.NewInt(50*KAVA_CF))),
					previousBorrowCoins:       sdk.NewCoins(sdk.NewCoin("usdx", sdkmath.NewInt(99*USDX_CF)), sdk.NewCoin("busd", sdkmath.NewInt(100*BUSD_CF))),
					borrowCoins:               sdk.NewCoins(sdk.NewCoin("usdx", sdkmath.NewInt(1*USDX_CF))),
					expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(50*KAVA_CF)), sdk.NewCoin("btcb", sdkmath.NewInt(100*BTCB_CF)), sdk.NewCoin("usdx", sdkmath.NewInt(100*USDX_CF)), sdk.NewCoin("busd", sdkmath.NewInt(100*BUSD_CF)), sdk.NewCoin("bnb", sdkmath.NewInt(70*BNB_CF)), sdk.NewCoin("xyz", sdkmath.NewInt(1))),
					expectedModAccountBalance: sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(1050*KAVA_CF)), sdk.NewCoin("bnb", sdkmath.NewInt(30*BUSD_CF)), sdk.NewCoin("usdx", sdkmath.NewInt(100*USDX_CF))),
				},
				{
					borrower:    sdk.AccAddress(crypto.AddressHash([]byte("test"))),
					borrowCoins: sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(1*KAVA_CF))),
				},
			},
			errArgs: []errArgs{
				{
					expectPass: true,
				},
				{
					// THIS SHOULDN'T FAIL
					expectPass: false,
					contains:   "reserves 2638559busd,12306usdx > cash 3000000000bnb,1050000000ukava,100000000usdx: insolvency - protocol reserves exceed available cash",
				},
			},
		},
		{
			name: "valid borrow followed by protocol reserves exceed available cash for busd when borrowing from busd",
			args: []args{
				{
					usdxBorrowLimit:           sdk.MustNewDecFromStr("100000000000"),
					priceKAVA:                 sdk.MustNewDecFromStr("2.00"),
					loanToValueKAVA:           sdk.MustNewDecFromStr("0.8"),
					priceBTCB:                 sdk.MustNewDecFromStr("0.00"),
					loanToValueBTCB:           sdk.MustNewDecFromStr("0.01"),
					priceBNB:                  sdk.MustNewDecFromStr("5.00"),
					loanToValueBNB:            sdk.MustNewDecFromStr("0.8"),
					borrower:                  sdk.AccAddress(crypto.AddressHash([]byte("test"))),
					depositCoins:              sdk.NewCoins(sdk.NewCoin("bnb", sdkmath.NewInt(30*BNB_CF)), sdk.NewCoin("ukava", sdkmath.NewInt(50*KAVA_CF))),
					previousBorrowCoins:       sdk.NewCoins(sdk.NewCoin("usdx", sdkmath.NewInt(99*USDX_CF)), sdk.NewCoin("busd", sdkmath.NewInt(100*BUSD_CF))),
					borrowCoins:               sdk.NewCoins(sdk.NewCoin("usdx", sdkmath.NewInt(1*USDX_CF))),
					expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(50*KAVA_CF)), sdk.NewCoin("btcb", sdkmath.NewInt(100*BTCB_CF)), sdk.NewCoin("usdx", sdkmath.NewInt(100*USDX_CF)), sdk.NewCoin("busd", sdkmath.NewInt(100*BUSD_CF)), sdk.NewCoin("bnb", sdkmath.NewInt(70*BNB_CF)), sdk.NewCoin("xyz", sdkmath.NewInt(1))),
					expectedModAccountBalance: sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(1050*KAVA_CF)), sdk.NewCoin("bnb", sdkmath.NewInt(30*BUSD_CF)), sdk.NewCoin("usdx", sdkmath.NewInt(100*USDX_CF))),
				},
				{
					borrower:    sdk.AccAddress(crypto.AddressHash([]byte("test"))),
					borrowCoins: sdk.NewCoins(sdk.NewCoin("busd", sdkmath.NewInt(100*BUSD_CF))),
				},
			},
			errArgs: []errArgs{
				{
					expectPass: true,
				},
				{
					expectPass: false,
					contains:   "reserves 2638559busd,12306usdx > cash 3000000000bnb,1050000000ukava,100000000usdx: insolvency - protocol reserves exceed available cash",
				},
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.Require().Len(tc.args, len(tc.errArgs), "malformed test setup, args should have same number of indexes as errArgs")
			var (
				ctx  sdk.Context
				tApp app.TestApp
			)
			for i, arg := range tc.args {
				if i == 0 {
					// Initialize test app and set context
					tApp = app.NewTestApp()
					ctx = tApp.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})

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
						[]sdk.AccAddress{arg.borrower},
					)

					// hard module genesis state
					hardGS := types.NewGenesisState(types.NewParams(
						types.MoneyMarkets{
							types.NewMoneyMarket("usdx", types.NewBorrowLimit(true, arg.usdxBorrowLimit, sdk.MustNewDecFromStr("1")), "usdx:usd", sdkmath.NewInt(USDX_CF), types.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec()),
							types.NewMoneyMarket("busd", types.NewBorrowLimit(false, sdk.NewDec(100000000*BUSD_CF), sdk.MustNewDecFromStr("1")), "busd:usd", sdkmath.NewInt(BUSD_CF), types.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec()),
							types.NewMoneyMarket("ukava", types.NewBorrowLimit(false, sdk.NewDec(100000000*KAVA_CF), arg.loanToValueKAVA), "kava:usd", sdkmath.NewInt(KAVA_CF), types.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.99"), sdk.ZeroDec()),
							types.NewMoneyMarket("btcb", types.NewBorrowLimit(false, sdk.NewDec(100000000*BTCB_CF), arg.loanToValueBTCB), "btcb:usd", sdkmath.NewInt(BTCB_CF), types.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec()),
							types.NewMoneyMarket("bnb", types.NewBorrowLimit(false, sdk.NewDec(100000000*BNB_CF), arg.loanToValueBNB), "bnb:usd", sdkmath.NewInt(BNB_CF), types.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec()),
							types.NewMoneyMarket("xyz", types.NewBorrowLimit(false, sdk.NewDec(1), arg.loanToValueBNB), "xyz:usd", sdkmath.NewInt(1), types.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec()),
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
								Price:         arg.priceKAVA,
								Expiry:        time.Now().Add(1 * time.Hour),
							},
							{
								MarketID:      "btcb:usd",
								OracleAddress: sdk.AccAddress{},
								Price:         arg.priceBTCB,
								Expiry:        time.Now().Add(1 * time.Hour),
							},
							{
								MarketID:      "bnb:usd",
								OracleAddress: sdk.AccAddress{},
								Price:         arg.priceBNB,
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
					hardMaccCoins := sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(1000*KAVA_CF)),
						sdk.NewCoin("usdx", sdkmath.NewInt(200*USDX_CF)), sdk.NewCoin("busd", sdkmath.NewInt(100*BUSD_CF)))
					err := bankKeeper.MintCoins(ctx, types.ModuleAccountName, hardMaccCoins)
					suite.Require().NoError(err)

					keeper := tApp.GetHardKeeper()
					suite.app = tApp
					suite.ctx = ctx
					suite.keeper = keeper
					// Run BeginBlocker once to transition MoneyMarkets
					hard.BeginBlocker(suite.ctx, suite.keeper)
					if err = suite.keeper.Deposit(suite.ctx, arg.borrower, arg.depositCoins); err != nil {
						suite.Require().Equal(arg.wantDepositErr, err.Error())
					}
					// Execute user's previous borrows
					if err = suite.keeper.Borrow(suite.ctx, arg.borrower, arg.previousBorrowCoins); arg.previousBorrowCoins.IsZero() {
						suite.Require().True(strings.Contains(err.Error(), "cannot borrow zero coins"))
					} else {
						suite.Require().NoError(err)
					}
					// keep running original first case
				} else {
					//suite.Require().NoError(suite.keeper.Withdraw(suite.ctx, arg.borrower, arg.depositCoins))
					blockDuration := time.Second * 3600 * 30 // long blocks to accumulate larger interest
					runAtTime := suite.ctx.BlockTime().Add(blockDuration)
					suite.ctx = suite.ctx.WithBlockTime(runAtTime)
					// Run BeginBlocker once to transition MoneyMarkets
					hard.BeginBlocker(suite.ctx, suite.keeper)
				}
				// Now that our state is properly set up, execute the last borrow
				err := suite.keeper.Borrow(suite.ctx, arg.borrower, arg.borrowCoins)
				if tc.errArgs[i].expectPass {
					suite.Require().NoError(err)

					// Check borrower balance
					acc := suite.getAccount(arg.borrower)
					suite.Require().Equal(arg.expectedAccountBalance, suite.getAccountCoins(acc))

					// Check module account balance
					mAcc := suite.getModuleAccount(types.ModuleAccountName)
					suite.Require().Equal(arg.expectedModAccountBalance, suite.getAccountCoins(mAcc))

					// Check that borrow struct is in store
					_, f := suite.keeper.GetBorrow(suite.ctx, arg.borrower)
					suite.Require().True(f)
				} else {
					suite.Require().Error(err)
					suite.Require().True(strings.Contains(err.Error(), tc.errArgs[i].contains))
				}
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

	suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(blockDuration))
	hard.BeginBlocker(suite.ctx, suite.keeper)

	// Should error since ukava has become insolvent at this point
	suite.Require().Error(suite.keeper.Borrow(
		suite.ctx,
		borrower,
		sdk.NewCoins(sdk.NewCoin("ukava", availableToBorrow.AmountOf("ukava"))),
	))

	suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(blockDuration))
	hard.BeginBlocker(suite.ctx, suite.keeper)

	// TODO: this is wrong, since usdk is not insolvent, ukava is.
	suite.Require().Error(suite.keeper.Borrow(
		suite.ctx,
		borrower,
		sdk.NewCoins(sdk.NewCoin("usdx", sdkmath.NewInt(25*USDX_CF))),
	))
}
