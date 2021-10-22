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
				MarketId:   "market",
				BaseAsset:  "xrp",
				QuoteAsset: "bnb",
				Oracles:    []string{addr.String()},
				Active:     true,
			},
			true,
		},
		{
			"invalid id",
			Market{
				MarketId: " ",
			},
			false,
		},
		{
			"invalid base asset",
			Market{
				MarketId:  "market",
				BaseAsset: "XRP",
			},
			false,
		},
		{
			"invalid market",
			Market{
				MarketId:  "market",
				BaseAsset: "xrp",
				// Denoms can be uppercase in v0.44
				QuoteAsset: "BNB.",
			},
			false,
		},
		{
			"empty oracle address ",
			Market{
				MarketId:   "market",
				BaseAsset:  "xrp",
				QuoteAsset: "bnb",
				Oracles:    []string{""},
			},
			false,
		},
		{
			"empty oracle address ",
			Market{
				MarketId:   "market",
				BaseAsset:  "xrp",
				QuoteAsset: "bnb",
				Oracles:    []string{addr.String(), addr.String()},
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
				MarketId:      "market",
				OracleAddress: addr.String(),
				Price:         sdk.OneDec(),
				Expiry:        now,
			},
			true,
		},
		{
			"invalid id",
			PostedPrice{
				MarketId: " ",
			},
			false,
		},
		{
			"invalid oracle",
			PostedPrice{
				MarketId:      "market",
				OracleAddress: "",
			},
			false,
		},
		{
			"invalid price",
			PostedPrice{
				MarketId:      "market",
				OracleAddress: addr.String(),
				Price:         sdk.NewDec(-1),
			},
			false,
		},
		{
			"zero expiry time ",
			PostedPrice{
				MarketId:      "market",
				OracleAddress: addr.String(),
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
