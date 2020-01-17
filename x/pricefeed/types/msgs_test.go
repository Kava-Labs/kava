package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	tmtime "github.com/tendermint/tendermint/types/time"
)

func TestMsgPlaceBid_ValidateBasic(t *testing.T) {
	addr := sdk.AccAddress([]byte("someName"))
	price, _ := sdk.NewDecFromStr("0.3005")
	expiry := tmtime.Now()
	negativePrice, _ := sdk.NewDecFromStr("-3.05")

	tests := []struct {
		name       string
		msg        MsgPostPrice
		expectPass bool
	}{
		{"normal", MsgPostPrice{addr, "xrp", price, expiry}, true},
		{"emptyAddr", MsgPostPrice{sdk.AccAddress{}, "xrp", price, expiry}, false},
		{"emptyAsset", MsgPostPrice{addr, "", price, expiry}, false},
		{"negativePrice", MsgPostPrice{addr, "xrp", negativePrice, expiry}, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expectPass {
				require.Nil(t, tc.msg.ValidateBasic())
			} else {
				require.NotNil(t, tc.msg.ValidateBasic())
			}
		})
	}
}
