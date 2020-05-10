package types

import (
	"testing"

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
			asset:   NewAssetSupply("kava", coin, coin, coin, coin),
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
			"invalid limit",
			AssetSupply{
				IncomingSupply: coin,
				OutgoingSupply: coin,
				CurrentSupply:  coin,
				Limit:          invalidCoin,
			},
			false,
		},
		{
			msg:     "invalid denom",
			asset:   NewAssetSupply("Invalid Denom", coin, coin, coin, coin),
			expPass: false,
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
