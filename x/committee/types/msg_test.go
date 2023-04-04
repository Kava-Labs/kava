package types

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

func MustNewMsgSubmitProposal(pubProposal PubProposal, proposer sdk.AccAddress, committeeId uint64) *MsgSubmitProposal {
	proposal, err := NewMsgSubmitProposal(pubProposal, proposer, committeeId)
	if err != nil {
		panic(err)
	}
	return proposal
}

func TestMsgSubmitProposal_ValidateBasic(t *testing.T) {
	addr := sdk.AccAddress(crypto.AddressHash([]byte("KavaTest1")))
	tests := []struct {
		name       string
		msg        *MsgSubmitProposal
		expectPass bool
	}{
		{
			name:       "normal",
			msg:        MustNewMsgSubmitProposal(govv1beta1.NewTextProposal("A Title", "A proposal description."), addr, 3),
			expectPass: true,
		},
		{
			name:       "empty address",
			msg:        MustNewMsgSubmitProposal(govv1beta1.NewTextProposal("A Title", "A proposal description."), nil, 3),
			expectPass: false,
		},
		{
			name:       "invalid proposal",
			msg:        &MsgSubmitProposal{PubProposal: &types.Any{}, Proposer: addr.String(), CommitteeID: 3},
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
	addr := sdk.AccAddress(crypto.AddressHash([]byte("KavaTest1")))
	tests := []struct {
		name       string
		msg        MsgVote
		expectPass bool
	}{
		{
			name:       "normal",
			msg:        MsgVote{5, addr.String(), VOTE_TYPE_YES},
			expectPass: true,
		},
		{
			name:       "No",
			msg:        MsgVote{5, addr.String(), VOTE_TYPE_NO},
			expectPass: true,
		},
		{
			name:       "Abstain",
			msg:        MsgVote{5, addr.String(), VOTE_TYPE_ABSTAIN},
			expectPass: true,
		},
		{
			name:       "Null vote",
			msg:        MsgVote{5, addr.String(), VOTE_TYPE_UNSPECIFIED},
			expectPass: false,
		},
		{
			name:       "empty address",
			msg:        MsgVote{5, "", VOTE_TYPE_YES},
			expectPass: false,
		},
		{
			name:       "invalid vote (greater)",
			msg:        MsgVote{5, addr.String(), 4},
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
