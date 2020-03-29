package committee

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/committee/keeper"
	"github.com/kava-labs/kava/x/committee/types"
)

// NewHandler creates an sdk.Handler for committee messages
func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case types.MsgSubmitProposal:
			return handleMsgSubmitProposal(ctx, k, msg)
		case types.MsgVote:
			return handleMsgVote(ctx, k, msg)
		default:
			errMsg := fmt.Sprintf("unrecognized %s msg type: %T", types.ModuleName, msg)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleMsgSubmitProposal(ctx sdk.Context, k keeper.Keeper, msg types.MsgSubmitProposal) sdk.Result {
	proposalID, err := k.SubmitProposal(ctx, msg.Proposer, msg.CommitteeID, msg.PubProposal)
	if err != nil {
		return err.Result()
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Proposer.String()),
		),
	)

	return sdk.Result{
		Data:   GetKeyFromID(proposalID),
		Events: ctx.EventManager().Events(),
	}
}

func handleMsgVote(ctx sdk.Context, k keeper.Keeper, msg types.MsgVote) sdk.Result {
	// get the proposal just to add fields to the event
	proposal, found := k.GetProposal(ctx, msg.ProposalID)
	if !found {
		return ErrUnknownProposal(DefaultCodespace, msg.ProposalID).Result()
	}

	err := k.AddVote(ctx, msg.ProposalID, msg.Voter)
	if err != nil {
		return err.Result()
	}

	// Enact a proposal if it has enough votes
	passes, err := k.GetProposalResult(ctx, msg.ProposalID)
	if err != nil {
		return err.Result()
	}
	if passes {
		err = k.EnactProposal(ctx, msg.ProposalID)
		outcome := types.AttributeValueProposalPassed
		if err != nil {
			outcome = types.AttributeValueProposalFailed
		}
		k.DeleteProposalAndVotes(ctx, msg.ProposalID)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeProposalClose,
				sdk.NewAttribute(types.AttributeKeyCommitteeID, fmt.Sprintf("%d", proposal.CommitteeID)),
				sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", proposal.ID)),
				sdk.NewAttribute(types.AttributeKeyProposalCloseStatus, outcome),
			),
		)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Voter.String()),
		),
	)

	return sdk.Result{Events: ctx.EventManager().Events()}
}
