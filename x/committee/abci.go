package committee

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/kava/x/committee/types"
)

// BeginBlocker runs at the start of every block.
func BeginBlocker(ctx sdk.Context, _ abci.RequestBeginBlock, k Keeper) {

	// Close all expired proposals
	// TODO optimize by using an index to avoid iterating over non expired proposals
	k.IterateProposals(ctx, func(proposal types.Proposal) bool {
		if proposal.HasExpiredBy(ctx.BlockTime()) {
			if err := k.CloseOutProposal(ctx, proposal.ID); err != nil {
				panic(err) // if an expired proposal does not close then something has gone very wrong
			}
		}
		return false
	})
}
