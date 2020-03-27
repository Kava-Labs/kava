package committee

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/kava/x/committee/types"
)

// BeginBlocker runs at the start of every block.
func BeginBlocker(ctx sdk.Context, _ abci.RequestBeginBlock, k Keeper) {

	// Close all expired proposals
	k.IterateProposals(ctx, func(proposal types.Proposal) bool {
		if proposal.HasExpiredBy(ctx.BlockTime()) {

			k.DeleteProposalAndVotes(ctx, proposal.ID)

			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeProposalClose,
					sdk.NewAttribute(types.AttributeKeyCommitteeID, fmt.Sprintf("%d", proposal.CommitteeID)),
					sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", proposal.ID)),
					sdk.NewAttribute(types.AttributeKeyProposalCloseStatus, types.AttributeValueProposalTimeout),
				),
			)
		}
		return false
	})
}
