package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/kava-labs/kava/x/committee/types"
)

type Keeper struct {
	// TODO other stuff as needed

	// Proposal router
	router govtypes.Router
}

/* TODO keeper methods - very similar to gov

- SubmitProposal validate and store a proposal, additionally setting things like timeout
- GetProposal
- SetProposal

- AddVote - add a vote to a particular proposal from a member
- GetVote
- SetVote

- GetCommittee
- SetCommittee

*/

func (k Keeper) SubmitProposal(ctx sdk.Context, msg types.MsgSubmitProposal) sdk.Error {
	// TODO Limit proposals to only be submitted by group members

	// Check group has permissions to enact proposal. As long as one permission allows the proposal then it goes through. Its the OR of all permissions.
	committee, _ := k.GetCommittee(ctx, msg.CommitteeID)
	hasPermissions := false
	for _, p := range committee.Permissions {
		if p.Allows(msg.Proposal) {
			hasPermissions = true
			break
		}
	}
	if !hasPermissions {
		return sdk.ErrInternal("committee does not have permissions to enact proposal").Result()
	}

	// TODO validate proposal by running it with cached context like how gov does it

	// TODO store the proposal, probably put it in a queue

	return nil
}

func (k Keeper) AddVote(ctx sdk.Context, msg types.MsgVote) sdk.Error {
	/* TODO
	- validate vote
	- store vote
	*/
	return nil
}
