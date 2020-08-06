package types

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestAssetSupplyValidate(t *testing.T) {
	coin := sdk.NewCoin("kava", sdk.OneInt())
	invalidCoin := sdk.Coin{Denom: "Invalid Denom", Amount: sdk.NewInt(-1)}
	testCases := []struct {
		msg     string
		asset   AssetSupply
		expPass bool
	}{
		{
			msg:     "valid asset",
			asset:   NewAssetSupply(coin, coin, coin, coin, time.Duration(0)),
			expPass: true,
		},
		{
			"invalid incoming supply",
			AssetSupply{IncomingSupply: invalidCoin},
			false,
		},
		{
			"invalid outgoing supply",
			AssetSupply{
				IncomingSupply: coin,
				OutgoingSupply: invalidCoin,
			},
			false,
		},
		{
			"invalid current supply",
			AssetSupply{
				IncomingSupply: coin,
				OutgoingSupply: coin,
				CurrentSupply:  invalidCoin,
			},
			false,
		},
		{
			"invalid time limitedcurrent supply",
			AssetSupply{
				IncomingSupply:           coin,
				OutgoingSupply:           coin,
				CurrentSupply:            coin,
				TimeLimitedCurrentSupply: invalidCoin,
			},
			false,
		},
		{
			"non matching denoms",
			AssetSupply{
				IncomingSupply:           coin,
				OutgoingSupply:           coin,
				CurrentSupply:            coin,
				TimeLimitedCurrentSupply: sdk.NewCoin("lol", sdk.ZeroInt()),
				TimeElapsed:              time.Hour,
			},
			false,
		},
	}

	for _, tc := range testCases {
		err := tc.asset.Validate()
		if tc.expPass {
			require.NoError(t, err, tc.msg)
		} else {
			require.Error(t, err, tc.msg)
		}
	}
}
