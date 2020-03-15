package committee

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/committee/types"
)

// InitGenesis initializes the store state from a genesis state.
func InitGenesis(ctx sdk.Context, keeper Keeper, gs GenesisState) {
	if err := gs.Validate(); err != nil {
		panic(fmt.Sprintf("failed to validate %s genesis state: %s", ModuleName, err))
	}

	keeper.SetNextProposalID(ctx, gs.NextProposalID)

	for _, com := range gs.Committees {
		keeper.SetCommittee(ctx, com)
	}
	for _, p := range gs.Proposals {
		keeper.SetProposal(ctx, p)
	}
	for _, v := range gs.Votes {
		keeper.SetVote(ctx, v)
	}
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper Keeper) GenesisState {

	nextID, err := keeper.GetNextProposalID(ctx)
	if err != nil {
		panic(err)
	}
	committees := []types.Committee{}
	keeper.IterateCommittees(ctx, func(com types.Committee) bool {
		committees = append(committees, com)
		return false
	})
	proposals := []types.Proposal{}
	votes := []types.Vote{}
	keeper.IterateProposals(ctx, func(p types.Proposal) bool {
		proposals = append(proposals, p)
		keeper.IterateVotes(ctx, p.ID, func(v types.Vote) bool {
			votes = append(votes, v)
			return false
		})
		return false
	})

	return types.NewGenesisState(
		nextID,
		committees,
		proposals,
		votes,
	)
}
