package simulation

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/kv"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/kava-labs/kava/x/committee/types"
)

func makeTestCodec() (cdc *codec.Codec) {
	cdc = codec.New()
	sdk.RegisterCodec(cdc)
	govtypes.RegisterCodec(cdc)
	types.RegisterCodec(cdc)
	return cdc
}

func TestDecodeStore(t *testing.T) {
	cdc := makeTestCodec()

	committee := types.NewMemberCommittee(
		12,
		"This committee is for testing.",
		nil,
		[]types.Permission{types.TextPermission{}},
		sdk.MustNewDecFromStr("0.667"),
		time.Hour*24*7,
		types.FirstPastThePost,
	)
	proposal := types.Proposal{
		ID:          34,
		CommitteeID: 12,
		Deadline:    time.Date(1998, time.January, 1, 1, 0, 0, 0, time.UTC),
		PubProposal: govtypes.NewTextProposal("A Title", "A description of this proposal."),
	}
	vote := types.Vote{
		ProposalID: 9,
		Voter:      nil,
	}

	kvPairs := kv.Pairs{
		kv.Pair{Key: types.CommitteeKeyPrefix, Value: cdc.MustMarshalBinaryLengthPrefixed(&committee)},
		kv.Pair{Key: types.ProposalKeyPrefix, Value: cdc.MustMarshalBinaryLengthPrefixed(&proposal)},
		kv.Pair{Key: types.VoteKeyPrefix, Value: cdc.MustMarshalBinaryLengthPrefixed(&vote)},
		kv.Pair{Key: types.NextProposalIDKey, Value: sdk.Uint64ToBigEndian(10)},
		kv.Pair{Key: []byte{0x99}, Value: []byte{0x99}},
	}

	tests := []struct {
		name        string
		expectedLog string
	}{
		{"Committee", fmt.Sprintf("%v\n%v", committee, committee)},
		{"Proposal", fmt.Sprintf("%v\n%v", proposal, proposal)},
		{"Vote", fmt.Sprintf("%v\n%v", vote, vote)},
		{"NextProposalID", "10\n10"},
		{"other", ""},
	}
	for i, tt := range tests {
		i, tt := i, tt
		t.Run(tt.name, func(t *testing.T) {
			switch i {
			case len(tests) - 1:
				require.Panics(t, func() { DecodeStore(cdc, kvPairs[i], kvPairs[i]) }, tt.name)
			default:
				require.Equal(t, tt.expectedLog, DecodeStore(cdc, kvPairs[i], kvPairs[i]), tt.name)
			}
		})
	}
}
