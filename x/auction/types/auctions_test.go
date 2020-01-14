package types

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
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
	TestAccAddress1         = "kava1qcfdf69js922qrdr4yaww3ax7gjml6pd39p8lj"
	TestAccAddress2         = "kava1pdfav2cjhry9k79nu6r8kgknnjtq6a7rcr0qlr"
)

func TestNewWeightedAddresses(t *testing.T) {

	tests := []struct {
		name       string
		addresses  []sdk.AccAddress
		weights    []sdk.Int
		expectpass bool
	}{
		{
			"normal",
			[]sdk.AccAddress{
				sdk.AccAddress([]byte(TestAccAddress1)),
				sdk.AccAddress([]byte(TestAccAddress2)),
			},
			[]sdk.Int{
				sdk.NewInt(6),
				sdk.NewInt(8),
			},
			true,
		},
		{
			"mismatched",
			[]sdk.AccAddress{
				sdk.AccAddress([]byte(TestAccAddress1)),
				sdk.AccAddress([]byte(TestAccAddress2)),
			},
			[]sdk.Int{
				sdk.NewInt(6),
			},
			false,
		},
		{
			"negativeWeight",
			[]sdk.AccAddress{
				sdk.AccAddress([]byte(TestAccAddress1)),
				sdk.AccAddress([]byte(TestAccAddress2)),
			},
			[]sdk.Int{
				sdk.NewInt(6),
				sdk.NewInt(-8),
			},
			false,
		},
	}

	// Run NewWeightedAdresses tests
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Attempt to instantiate new WeightedAddresses
			weightedAddresses, err := NewWeightedAddresses(tc.addresses, tc.weights)

			if tc.expectpass {
				// Confirm there is no error
				require.Nil(t, err)

				// Check addresses, weights
				require.Equal(t, tc.addresses, weightedAddresses.Addresses)
				require.Equal(t, tc.weights, weightedAddresses.Weights)
			} else {
				// Confirm that there is an error
				require.NotNil(t, err)

				switch tc.name {
				case "mismatched":
					require.Contains(t, err.Error(), "number of addresses doesn't match number of weights")
				case "negativeWeight":
					require.Contains(t, err.Error(), "weights contain a negative amount")
				default:
					// Unexpected error state
					t.Fail()
				}
			}
		})
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
	addresses := []sdk.AccAddress{
		sdk.AccAddress([]byte(TestAccAddress1)),
		sdk.AccAddress([]byte(TestAccAddress2)),
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

	require.Equal(t, collateralAuction.BaseAuction.Initiator, TestInitiatorModuleName)
	require.Equal(t, collateralAuction.BaseAuction.Lot, c(TestLotDenom, TestLotAmount))
	require.Equal(t, collateralAuction.BaseAuction.Bid, c(TestBidDenom, 0))
	require.Equal(t, collateralAuction.BaseAuction.EndTime, endTime)
	require.Equal(t, collateralAuction.BaseAuction.MaxEndTime, endTime)
	require.Equal(t, collateralAuction.MaxBid, c(TestBidDenom, TestBidAmount))
	require.Equal(t, collateralAuction.LotReturns, weightedAddresses)
	require.Equal(t, collateralAuction.CorrespondingDebt, c(TestDebtDenom, TestDebtAmount2))
}
