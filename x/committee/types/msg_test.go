package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

func TestMsgSubmitProposal_ValidateBasic(t *testing.T) {
	addr := sdk.AccAddress([]byte("someName"))
	tests := []struct {
		name       string
		msg        MsgSubmitProposal
		expectPass bool
	}{
		{
			name:       "normal",
			msg:        MsgSubmitProposal{govtypes.NewTextProposal("A Title", "A proposal description."), addr, 3},
			expectPass: true,
		},
		{
			name:       "empty address",
			msg:        MsgSubmitProposal{govtypes.NewTextProposal("A Title", "A proposal description."), nil, 3},
			expectPass: false,
		},
		{
			name:       "invalid proposal",
			msg:        MsgSubmitProposal{govtypes.TextProposal{}, addr, 3},
			expectPass: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			err := tc.msg.ValidateBasic()

			if tc.expectPass {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestMsgVote_ValidateBasic(t *testing.T) {
	addr := sdk.AccAddress([]byte("someName"))
	tests := []struct {
		name       string
		msg        MsgVote
		expectPass bool
	}{
		{
			name:       "normal",
			msg:        MsgVote{5, addr, Yes},
			expectPass: true,
		},
		{
			name:       "No",
			msg:        MsgVote{5, addr, No},
			expectPass: true,
		},
		{
			name:       "Abstain",
			msg:        MsgVote{5, addr, Abstain},
			expectPass: true,
		},
		{
			name:       "Null vote",
			msg:        MsgVote{5, addr, NullVoteType},
			expectPass: false,
		},
		{
			name:       "empty address",
			msg:        MsgVote{5, nil, Yes},
			expectPass: false,
		},
		{
			name:       "invalid vote (greater)",
			msg:        MsgVote{5, addr, 4},
			expectPass: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			err := tc.msg.ValidateBasic()

			if tc.expectPass {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
