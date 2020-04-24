package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/committee/types"
)

// proposalVoteMap collects up votes into a map indexed by proposalID
func getProposalVoteMap(k keeper.Keeper, ctx sdk.Context) map[uint64]([]types.Vote) {

	proposalVoteMap = map[uint64]([]types.Vote){}

	keeper.IterateProposals(suite.ctx, func(p types.Proposal) bool {
		keeper.IterateVotes(suite.ctx, p.ID, func(v types.Vote) bool {
			proposalVoteMap[p.ID] = append(proposalVoteMap[p.ID], v)
			return false
		})
		return false
	})
	return proposalVoteMap
}
