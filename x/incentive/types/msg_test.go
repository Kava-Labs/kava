package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"

	"github.com/kava-labs/kava/x/incentive/types"
)

func TestMsgClaimUSDXMintingReward_Validate(t *testing.T) {
	tests := []struct {
		from           sdk.AccAddress
		multiplierName string
		expectPass     bool
	}{
		{
			from:           sdk.AccAddress(crypto.AddressHash([]byte("KavaTest1"))),
			multiplierName: "large",
			expectPass:     true,
		},
		{
			from:           sdk.AccAddress(crypto.AddressHash([]byte("KavaTest1"))),
			multiplierName: "medium",
			expectPass:     true,
		},
		{
			from:           sdk.AccAddress(crypto.AddressHash([]byte("KavaTest1"))),
			multiplierName: "small",
			expectPass:     true,
		},
		{
			from:           sdk.AccAddress{},
			multiplierName: "medium",
			expectPass:     false,
		},
		{
			from:           sdk.AccAddress(crypto.AddressHash([]byte("KavaTest1"))),
			multiplierName: "huge",
			expectPass:     false,
		},
	}
	for _, tc := range tests {
		msg := types.NewMsgClaimUSDXMintingReward(tc.from, tc.multiplierName)
		err := msg.ValidateBasic()
		if tc.expectPass {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
		}
	}
}
