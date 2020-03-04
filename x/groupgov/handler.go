package cdp

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/groupgov/types"
)

// NewHandler creates an sdk.Handler for cdp messages
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case types.MsgSubmitProposal:
			handleMsgSubmitProposal(ctx, k, msg)
		case types.MsgVote:
			handleMsgVote(ctx, k, msg)
		default:
			errMsg := fmt.Sprintf("unrecognized %s msg type: %T", , types.ModuleName, msg)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleMsgSubmitProposal(ctx sdk.Context, k Keeper, msg types.MsgSubmitProposal) sdk.Result {
	// TODO limit proposals to only be submitted by group members
	
	// get group
	group, _ := k.GetGroup(ctx, msg.GroupID)
	// Check group has permissions to enact proposal. As long as one permission allows the proposal then it goes through. Its the OR of all permissions.
	var hasPermissions := false
	for p, _ := range group.Permissions {
		if p.Allows(msg.Proposal) {
			hasPermissions = true
			break
		}
	}
	if !hasPermissions {
		return sdk.ErrInternal("group does not have permissions to enact proposal").Result()
	}
	// TODO validate proposal by running it with cached context like how gov does it
	// TODO store the proposal, probably put it in a queue

}

func handleMsgVote(ctx sdk.Context, k Keeper, msg types.MsgVote) sdk.Result {
	/* TODO
	- validate vote
	- store vote
	*/
}
