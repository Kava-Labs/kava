package committee

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/committee/keeper"
	"github.com/kava-labs/kava/x/committee/types"
)

// InitGenesis initializes the store state from a genesis state.
func InitGenesis(ctx sdk.Context, keeper keeper.Keeper, gs *types.GenesisState) {
	if err := gs.Validate(); err != nil {
		panic(fmt.Sprintf("failed to validate %s genesis state: %s", types.ModuleName, err))
	}

	keeper.SetNextProposalID(ctx, gs.NextProposalID)

	for _, com := range gs.GetCommittees() {
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
func ExportGenesis(ctx sdk.Context, keeper keeper.Keeper) *types.GenesisState {

	nextID, err := keeper.GetNextProposalID(ctx)
	if err != nil {
		panic(err)
	}
	committees := keeper.GetCommittees(ctx)
	proposals := keeper.GetProposals(ctx)
	votes := keeper.GetVotes(ctx)

	return types.NewGenesisState(
		nextID,
		committees,
		proposals,
		votes,
	)
}
