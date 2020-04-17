package simulation

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	cmn "github.com/tendermint/tendermint/libs/common"

	"github.com/kava-labs/kava/x/committee/types"
)

// DecodeStore unmarshals the KVPair's Value to the corresponding module type
func DecodeStore(cdc *codec.Codec, kvA, kvB cmn.KVPair) string {
	switch {
	case bytes.Equal(kvA.Key[:1], types.CommitteeKeyPrefix):
		var committeeA, committeeB types.Committee
		cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &committeeA)
		cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &committeeB)
		return fmt.Sprintf("%v\n%v", committeeA, committeeB)

	case bytes.Equal(kvA.Key[:1], types.ProposalKeyPrefix):
		var proposalA, proposalB types.Proposal
		cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &proposalA)
		cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &proposalB)
		return fmt.Sprintf("%v\n%v", proposalA, proposalB)

	case bytes.Equal(kvA.Key[:1], types.VoteKeyPrefix):
		var voteA, voteB types.Vote
		cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &voteA)
		cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &voteB)
		return fmt.Sprintf("%v\n%v", voteA, voteB)

	case bytes.Equal(kvA.Key[:1], types.NextProposalIDKey):
		proposalIDA := types.Uint64FromBytes(kvA.Value)
		proposalIDB := types.Uint64FromBytes(kvB.Value)
		return fmt.Sprintf("%d\n%d", proposalIDA, proposalIDB)

	default:
		panic(fmt.Sprintf("invalid %s key prefix %X", types.ModuleName, kvA.Key[:1]))
	}
}
