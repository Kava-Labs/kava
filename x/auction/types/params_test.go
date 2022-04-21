package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParams_Validate(t *testing.T) {
	testCases := []struct {
		name string
		Params
		expectErr bool
	}{
		{
			"normal",
			DefaultParams(),
			false,
		},
		{
			"negativeForwardBidDuration",
			Params{
				MaxAuctionDuration:  24 * time.Hour,
				ForwardBidDuration:  -1 * time.Hour,
				ReverseBidDuration:  1 * time.Hour,
				IncrementSurplus:    d("0.05"),
				IncrementDebt:       d("0.05"),
				IncrementCollateral: d("0.05"),
			},
			true,
		},
		{
			"negativeReverseBidDuration",
			Params{
				MaxAuctionDuration:  24 * time.Hour,
				ForwardBidDuration:  1 * time.Hour,
				ReverseBidDuration:  -1 * time.Hour,
				IncrementSurplus:    d("0.05"),
				IncrementDebt:       d("0.05"),
				IncrementCollateral: d("0.05"),
			},
			true,
		},
		{
			"negativeBidDuration",
			Params{
				MaxAuctionDuration:  24 * time.Hour,
				ForwardBidDuration:  -1 * time.Hour,
				ReverseBidDuration:  -1 * time.Hour,
				IncrementSurplus:    d("0.05"),
				IncrementDebt:       d("0.05"),
				IncrementCollateral: d("0.05"),
			},
			true,
		},
		{
			"negativeAuction",
			Params{
				MaxAuctionDuration:  -24 * time.Hour,
				ForwardBidDuration:  1 * time.Hour,
				ReverseBidDuration:  1 * time.Hour,
				IncrementSurplus:    d("0.05"),
				IncrementDebt:       d("0.05"),
				IncrementCollateral: d("0.05"),
			},
			true,
		},
		{
			"bid>auction",
			Params{
				MaxAuctionDuration:  1 * time.Hour,
				ForwardBidDuration:  24 * time.Hour,
				ReverseBidDuration:  1 * time.Hour,
				IncrementSurplus:    d("0.05"),
				IncrementDebt:       d("0.05"),
				IncrementCollateral: d("0.05"),
			},
			true,
		},
		{
			"negative increment surplus",
			Params{
				MaxAuctionDuration:  24 * time.Hour,
				ForwardBidDuration:  1 * time.Hour,
				ReverseBidDuration:  1 * time.Hour,
				IncrementSurplus:    d("-0.05"),
				IncrementDebt:       d("0.05"),
				IncrementCollateral: d("0.05"),
			},
			true,
		},
		{
			"negative increment debt",
			Params{
				MaxAuctionDuration:  24 * time.Hour,
				ForwardBidDuration:  1 * time.Hour,
				ReverseBidDuration:  1 * time.Hour,
				IncrementSurplus:    d("0.05"),
				IncrementDebt:       d("-0.05"),
				IncrementCollateral: d("0.05"),
			},
			true,
		},
		{
			"negative increment collateral",
			Params{
				MaxAuctionDuration:  24 * time.Hour,
				ForwardBidDuration:  1 * time.Hour,
				ReverseBidDuration:  1 * time.Hour,
				IncrementSurplus:    d("0.05"),
				IncrementDebt:       d("0.05"),
				IncrementCollateral: d("-0.05"),
			},
			true,
		},
		{
			"zero value",
			Params{},
			true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.Params.Validate()
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
