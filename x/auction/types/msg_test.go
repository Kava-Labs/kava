package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestMsgPlaceBid_ValidateBasic(t *testing.T) {
	tests := []struct {
		name       string
		msg        MsgPlaceBid
		expectPass bool
	}{
		{
			"normal",
			NewMsgPlaceBid(1, testAccAddress1, c("token", 10)),
			true,
		},
		{
			"zero id",
			NewMsgPlaceBid(0, testAccAddress1, c("token", 10)),
			false,
		},
		{
			"empty address ",
			NewMsgPlaceBid(1, "", c("token", 10)),
			false,
		},
		{
			"negative amount",
			NewMsgPlaceBid(1, testAccAddress1, sdk.Coin{Denom: "token", Amount: sdkmath.NewInt(-10)}),
			false,
		},
		{
			"zero amount",
			NewMsgPlaceBid(1, testAccAddress1, c("token", 0)),
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
