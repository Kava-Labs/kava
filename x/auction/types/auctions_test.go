package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	TestInitiatorModuleName = "liquidator"
	TestLotDenom            = "usdx"
	TestLotAmount           = 100
	TestBidDenom            = "kava"
	TestBidAmount           = 20
	TestDebtDenom           = "debt"
	TestDebtAmount1         = 20
	TestDebtAmount2         = 15
	TestExtraEndTime        = 10000
	TestAuctionID           = 9999123
	testAccAddress1         = "kava1qcfdf69js922qrdr4yaww3ax7gjml6pd39p8lj"
	testAccAddress2         = "kava1pdfav2cjhry9k79nu6r8kgknnjtq6a7rcr0qlr"
)

func init() {
	sdk.GetConfig().SetBech32PrefixForAccount("kava", "kava"+sdk.PrefixPublic)
}

func d(amount string) sdk.Dec               { return sdk.MustNewDecFromStr(amount) }
func c(denom string, amount int64) sdk.Coin { return sdk.NewInt64Coin(denom, amount) }
func i(n int64) sdk.Int                     { return sdk.NewInt(n) }
func is(ns ...int64) (is []sdk.Int) {
	for _, n := range ns {
		is = append(is, sdk.NewInt(n))
	}
	return
}

func TestNewWeightedAddresses(t *testing.T) {
	tests := []struct {
		name      string
		addresses []string
		weights   []sdk.Int
		expPass   bool
	}{
		{
			"normal",
			[]string{testAccAddress1, testAccAddress2},
			[]sdk.Int{sdk.NewInt(6), sdk.NewInt(8)},
			true,
		},
		{
			"empty address",
			[]string{"", ""},
			[]sdk.Int{sdk.NewInt(6), sdk.NewInt(8)},
			false,
		},
		{
			"mismatched",
			[]string{testAccAddress1, testAccAddress2},
			[]sdk.Int{sdk.NewInt(6)},
			false,
		},
		{
			"negative weight",
			[]string{testAccAddress1, testAccAddress2},
			is(6, -8),
			false,
		},
		{
			"zero weight",
			[]string{testAccAddress1, testAccAddress2},
			is(0, 0),
			false,
		},
	}

	// Run NewWeightedAdresses tests
	for _, tc := range tests {
		// Attempt to instantiate new WeightedAddresses
		weightedAddresses, err := NewWeightedAddresses(tc.addresses, tc.weights)

		if tc.expPass {
			require.NoError(t, err)
			require.Equal(t, tc.addresses, weightedAddresses.Addresses)
			require.Equal(t, tc.weights, weightedAddresses.Weights)
		} else {
			require.Error(t, err)
		}
	}
}

func TestDebtAuctionValidate(t *testing.T) {
	addr1, err := sdk.AccAddressFromBech32(testAccAddress1)
	require.NoError(t, err)

	now := time.Now()

	tests := []struct {
		msg     string
		auction DebtAuction
		expPass bool
	}{
		{
			"valid auction",
			DebtAuction{
				BaseAuction: BaseAuction{
					ID:              1,
					Initiator:       testAccAddress1,
					Lot:             c("kava", 1),
					Bidder:          addr1.String(),
					Bid:             c("kava", 1),
					EndTime:         now,
					MaxEndTime:      now,
					HasReceivedBids: true,
				},
				CorrespondingDebt: c("kava", 1),
			},
			true,
		},
		{
			"invalid corresponding debt",
			DebtAuction{
				BaseAuction: BaseAuction{
					ID:              1,
					Initiator:       testAccAddress1,
					Lot:             c("kava", 1),
					Bidder:          addr1.String(),
					Bid:             c("kava", 1),
					EndTime:         now,
					MaxEndTime:      now,
					HasReceivedBids: true,
				},
				CorrespondingDebt: sdk.Coin{Denom: "", Amount: sdk.NewInt(1)},
			},
			false,
		},
	}

	for _, tc := range tests {

		err := tc.auction.Validate()

		if tc.expPass {
			require.NoError(t, err, tc.msg)
		} else {
			require.Error(t, err, tc.msg)
		}
	}
}

func TestCollateralAuctionValidate(t *testing.T) {
	addr1, err := sdk.AccAddressFromBech32(testAccAddress1)
	require.NoError(t, err)

	now := time.Now()

	tests := []struct {
		msg     string
		auction CollateralAuction
		expPass bool
	}{
		{
			"valid auction",
			CollateralAuction{
				BaseAuction: BaseAuction{
					ID:              1,
					Initiator:       testAccAddress1,
					Lot:             c("kava", 1),
					Bidder:          addr1.String(),
					Bid:             c("kava", 1),
					EndTime:         now,
					MaxEndTime:      now,
					HasReceivedBids: true,
				},
				CorrespondingDebt: c("kava", 1),
				MaxBid:            c("kava", 1),
				LotReturns: WeightedAddresses{
					Addresses: []string{testAccAddress1},
					Weights:   []sdk.Int{sdk.NewInt(1)},
				},
			},
			true,
		},
		{
			"invalid corresponding debt",
			CollateralAuction{
				BaseAuction: BaseAuction{
					ID:              1,
					Initiator:       testAccAddress1,
					Lot:             c("kava", 1),
					Bidder:          addr1.String(),
					Bid:             c("kava", 1),
					EndTime:         now,
					MaxEndTime:      now,
					HasReceivedBids: true,
				},
				CorrespondingDebt: sdk.Coin{Denom: "DENOM", Amount: sdk.NewInt(1)},
			},
			false,
		},
		{
			"invalid max bid",
			CollateralAuction{
				BaseAuction: BaseAuction{
					ID:              1,
					Initiator:       testAccAddress1,
					Lot:             c("kava", 1),
					Bidder:          addr1.String(),
					Bid:             c("kava", 1),
					EndTime:         now,
					MaxEndTime:      now,
					HasReceivedBids: true,
				},
				CorrespondingDebt: c("kava", 1),
				MaxBid:            sdk.Coin{Denom: "DENOM", Amount: sdk.NewInt(1)},
			},
			false,
		},
		{
			"invalid lot returns",
			CollateralAuction{
				BaseAuction: BaseAuction{
					ID:              1,
					Initiator:       testAccAddress1,
					Lot:             c("kava", 1),
					Bidder:          addr1.String(),
					Bid:             c("kava", 1),
					EndTime:         now,
					MaxEndTime:      now,
					HasReceivedBids: true,
				},
				CorrespondingDebt: c("kava", 1),
				MaxBid:            c("kava", 1),
				LotReturns: WeightedAddresses{
					Addresses: []string{""},
					Weights:   []sdk.Int{sdk.NewInt(1)},
				},
			},
			false,
		},
	}

	for _, tc := range tests {

		err := tc.auction.Validate()

		if tc.expPass {
			require.NoError(t, err, tc.msg)
		} else {
			require.Error(t, err, tc.msg)
		}
	}
}

func TestBaseAuctionGetters(t *testing.T) {
	endTime := time.Now().Add(TestExtraEndTime)

	// Create a new BaseAuction (via SurplusAuction)
	auction := NewSurplusAuction(
		TestInitiatorModuleName,
		c(TestLotDenom, TestLotAmount),
		TestBidDenom, endTime,
	)

	auctionID := auction.GetID()
	auctionBid := auction.GetBid()
	auctionLot := auction.GetLot()
	auctionEndTime := auction.GetEndTime()
	auctionString := auction.String()

	require.Equal(t, auction.ID, auctionID)
	require.Equal(t, auction.Bid, auctionBid)
	require.Equal(t, auction.Lot, auctionLot)
	require.Equal(t, auction.EndTime, auctionEndTime)
	require.NotNil(t, auctionString)
}

func TestNewSurplusAuction(t *testing.T) {
	endTime := time.Now().Add(TestExtraEndTime)

	// Create a new SurplusAuction
	surplusAuction := NewSurplusAuction(
		TestInitiatorModuleName,
		c(TestLotDenom, TestLotAmount),
		TestBidDenom, endTime,
	)

	require.Equal(t, surplusAuction.Initiator, TestInitiatorModuleName)
	require.Equal(t, surplusAuction.Lot, c(TestLotDenom, TestLotAmount))
	require.Equal(t, surplusAuction.Bid, c(TestBidDenom, 0))
	require.Equal(t, surplusAuction.EndTime, endTime)
	require.Equal(t, surplusAuction.MaxEndTime, endTime)
}

func TestNewDebtAuction(t *testing.T) {
	endTime := time.Now().Add(TestExtraEndTime)

	// Create a new DebtAuction
	debtAuction := NewDebtAuction(
		TestInitiatorModuleName,
		c(TestBidDenom, TestBidAmount),
		c(TestLotDenom, TestLotAmount),
		endTime,
		c(TestDebtDenom, TestDebtAmount1),
	)

	require.Equal(t, debtAuction.Initiator, TestInitiatorModuleName)
	require.Equal(t, debtAuction.Lot, c(TestLotDenom, TestLotAmount))
	require.Equal(t, debtAuction.Bid, c(TestBidDenom, TestBidAmount))
	require.Equal(t, debtAuction.EndTime, endTime)
	require.Equal(t, debtAuction.MaxEndTime, endTime)
	require.Equal(t, debtAuction.CorrespondingDebt, c(TestDebtDenom, TestDebtAmount1))
}

func TestNewCollateralAuction(t *testing.T) {
	// Set up WeightedAddresses
	addresses := []string{
		testAccAddress1,
		testAccAddress2,
	}

	weights := []sdk.Int{
		sdk.NewInt(6),
		sdk.NewInt(8),
	}

	weightedAddresses, _ := NewWeightedAddresses(addresses, weights)

	endTime := time.Now().Add(TestExtraEndTime)

	collateralAuction := NewCollateralAuction(
		TestInitiatorModuleName,
		c(TestLotDenom, TestLotAmount),
		endTime,
		c(TestBidDenom, TestBidAmount),
		weightedAddresses,
		c(TestDebtDenom, TestDebtAmount2),
	)

	require.Equal(t, collateralAuction.Initiator, TestInitiatorModuleName)
	require.Equal(t, collateralAuction.Lot, c(TestLotDenom, TestLotAmount))
	require.Equal(t, collateralAuction.Bid, c(TestBidDenom, 0))
	require.Equal(t, collateralAuction.EndTime, endTime)
	require.Equal(t, collateralAuction.MaxEndTime, endTime)
	require.Equal(t, collateralAuction.MaxBid, c(TestBidDenom, TestBidAmount))
	require.Equal(t, collateralAuction.LotReturns, weightedAddresses)
	require.Equal(t, collateralAuction.CorrespondingDebt, c(TestDebtDenom, TestDebtAmount2))
}
