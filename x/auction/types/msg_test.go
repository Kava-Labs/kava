package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestMsgPlaceBid_ValidateBasic(t *testing.T) {
	addr, err := sdk.AccAddressFromBech32(testAccAddress1)
	require.NoError(t, err)

	tests := []struct {
		name       string
		msg        MsgPlaceBid
		expectPass bool
	}{
		{
			"normal",
			NewMsgPlaceBid(1, addr, c("token", 10)),
			true,
		},
		{
			"zero id",
			NewMsgPlaceBid(0, addr, c("token", 10)),
			false,
		},
		{
			"empty address ",
			NewMsgPlaceBid(1, nil, c("token", 10)),
			false,
		},
		{
			"invalid address",
			NewMsgPlaceBid(1, addr[:10], c("token", 10)),
			false,
		},
		{
			"negative amount",
			NewMsgPlaceBid(1, addr, sdk.Coin{Denom: "token", Amount: sdk.NewInt(-10)}),
			false,
		},
		{
			"zero amount",
			NewMsgPlaceBid(1, addr, c("token", 0)),
			true,
		},
	}

	for _, tc := range tests {
		if tc.expectPass {
			require.NoError(t, tc.msg.ValidateBasic(), tc.name)
		} else {
			require.Error(t, tc.msg.ValidateBasic(), tc.name)
		}
	}
}

func c(denom string, amount int64) sdk.Coin { return sdk.NewInt64Coin(denom, amount) }
