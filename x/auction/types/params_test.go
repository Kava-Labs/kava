package types

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestParams_Validate(t *testing.T) {
	type fields struct {
	}
	testCases := []struct {
		name               string
		MaxAuctionDuration time.Duration
		BidDuration        time.Duration
		expectErr          bool
	}{
		{"normal", 24 * time.Hour, 1 * time.Hour, false},
		{"negativeBid", 24 * time.Hour, -1 * time.Hour, true},
		{"negativeAuction", -24 * time.Hour, 1 * time.Hour, true},
		{"bid>auction", 1 * time.Hour, 24 * time.Hour, true},
		{"zeros", 0, 0, false},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := Params{
				MaxAuctionDuration: tc.MaxAuctionDuration,
				BidDuration:        tc.BidDuration,
			}
			err := p.Validate()
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
