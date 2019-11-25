package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestMsgPlaceBid_ValidateBasic(t *testing.T) {
	addr := sdk.AccAddress([]byte("someName"))
	tests := []struct {
		name       string
		msg        MsgPlaceBid
		expectPass bool
	}{
		{"normal", MsgPlaceBid{0, addr, sdk.NewInt64Coin("usdx", 10), sdk.NewInt64Coin("kava", 20)}, true},
		{"emptyAddr", MsgPlaceBid{0, sdk.AccAddress{}, sdk.NewInt64Coin("usdx", 10), sdk.NewInt64Coin("kava", 20)}, false},
		{"negativeBid", MsgPlaceBid{0, addr, sdk.Coin{"usdx", sdk.NewInt(-10)}, sdk.NewInt64Coin("kava", 20)}, false},
		{"negativeLot", MsgPlaceBid{0, addr, sdk.NewInt64Coin("usdx", 10), sdk.Coin{"kava", sdk.NewInt(-20)}}, false},
		{"zerocoins", MsgPlaceBid{0, addr, sdk.NewInt64Coin("usdx", 0), sdk.NewInt64Coin("kava", 0)}, true},
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
