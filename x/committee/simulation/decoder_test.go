package simulation

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	cmn "github.com/tendermint/tendermint/libs/common"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"

	"github.com/kava-labs/kava/x/committee/types"
)

func makeTestCodec() (cdc *codec.Codec) {
	cdc = codec.New()
	sdk.RegisterCodec(cdc)
	gov.RegisterCodec(cdc)
	types.RegisterAppCodec(cdc)
	return cdc
}

func TestDecodeStore(t *testing.T) {
	cdc := makeTestCodec()

	committee := types.NewCommittee(
		12,
		"This committee is for testing.",
		nil,
		[]types.Permission{types.GodPermission{}},
		sdk.MustNewDecFromStr("0.667"),
		time.Hour*24*7,
	)
	proposal := types.Proposal{
		ID:          34,
		CommitteeID: 12,
		Deadline:    time.Date(1998, time.January, 1, 1, 0, 0, 0, time.UTC),
		PubProposal: gov.NewTextProposal("A Title", "A description of this proposal."),
	}
	vote := types.Vote{
		ProposalID: 9,
		Voter:      nil,
	}

	kvPairs := cmn.KVPairs{
		cmn.KVPair{Key: types.CommitteeKeyPrefix, Value: cdc.MustMarshalBinaryLengthPrefixed(&committee)},
		cmn.KVPair{Key: types.ProposalKeyPrefix, Value: cdc.MustMarshalBinaryLengthPrefixed(&proposal)},
		cmn.KVPair{Key: types.VoteKeyPrefix, Value: cdc.MustMarshalBinaryLengthPrefixed(&vote)},
		cmn.KVPair{Key: types.NextProposalIDKey, Value: sdk.Uint64ToBigEndian(10)},
		cmn.KVPair{Key: []byte{0x99}, Value: []byte{0x99}},
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
