package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	tmtypes "github.com/tendermint/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestMarketValidate(t *testing.T) {
	mockPrivKey := tmtypes.NewMockPV()
	pubkey, err := mockPrivKey.GetPubKey()
	require.NoError(t, err)
	addr := sdk.AccAddress(pubkey.Address())

	testCases := []struct {
		msg     string
		market  Market
		expPass bool
	}{
		{
			"valid market",
			Market{
				MarketID:   "market",
				BaseAsset:  "xrp",
				QuoteAsset: "bnb",
				Oracles:    []sdk.AccAddress{addr},
				Active:     true,
			},
			true,
		},
		{
			"invalid id",
			Market{
				MarketID: " ",
			},
			false,
		},
		{
			"invalid base asset",
			Market{
				MarketID:  "market",
				BaseAsset: "XRP",
			},
			false,
		},
		{
			"invalid market",
			Market{
				MarketID:  "market",
				BaseAsset: "xrp",
				// Denoms can be uppercase in v0.44
				QuoteAsset: "BNB~",
			},
			false,
		},
		{
			"empty oracle address ",
			Market{
				MarketID:   "market",
				BaseAsset:  "xrp",
				QuoteAsset: "bnb",
				Oracles:    []sdk.AccAddress{nil},
			},
			false,
		},
		{
			"empty oracle address ",
			Market{
				MarketID:   "market",
				BaseAsset:  "xrp",
				QuoteAsset: "bnb",
				Oracles:    []sdk.AccAddress{addr, addr},
			},
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.msg, func(t *testing.T) {
			err := tc.market.Validate()
			if tc.expPass {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestPostedPriceValidate(t *testing.T) {
	now := time.Now()
	mockPrivKey := tmtypes.NewMockPV()
	pubkey, err := mockPrivKey.GetPubKey()
	require.NoError(t, err)
	addr := sdk.AccAddress(pubkey.Address())

	testCases := []struct {
		msg         string
		postedPrice PostedPrice
		expPass     bool
	}{
		{
			"valid posted price",
			PostedPrice{
				MarketID:      "market",
				OracleAddress: addr,
				Price:         sdk.OneDec(),
				Expiry:        now,
			},
			true,
		},
		{
			"invalid id",
			PostedPrice{
				MarketID: " ",
			},
			false,
		},
		{
			"invalid oracle",
			PostedPrice{
				MarketID:      "market",
				OracleAddress: sdk.AccAddress{},
			},
			false,
		},
		{
			"invalid price",
			PostedPrice{
				MarketID:      "market",
				OracleAddress: addr,
				Price:         sdk.NewDec(-1),
			},
			false,
		},
		{
			"zero expiry time ",
			PostedPrice{
				MarketID:      "market",
				OracleAddress: addr,
				Price:         sdk.OneDec(),
				Expiry:        time.Time{},
			},
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.msg, func(t *testing.T) {
			err := tc.postedPrice.Validate()
			if tc.expPass {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
