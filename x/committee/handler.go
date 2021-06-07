package committee

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/committee/keeper"
	"github.com/kava-labs/kava/x/committee/types"
)

// NewHandler creates an sdk.Handler for committee messages
func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case types.MsgSubmitProposal:
			return handleMsgSubmitProposal(ctx, k, msg)
		case types.MsgVote:
			return handleMsgVote(ctx, k, msg)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized %s message type: %T", ModuleName, msg)
		}
	}
}

func handleMsgSubmitProposal(ctx sdk.Context, k keeper.Keeper, msg types.MsgSubmitProposal) (*sdk.Result, error) {
	proposalID, err := k.SubmitProposal(ctx, msg.Proposer, msg.CommitteeID, msg.PubProposal)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Proposer.String()),
		),
	)

	return &sdk.Result{
		Data:   GetKeyFromID(proposalID),
		Events: ctx.EventManager().Events(),
	}, nil
}

func handleMsgVote(ctx sdk.Context, k keeper.Keeper, msg types.MsgVote) (*sdk.Result, error) {
	err := k.AddVote(ctx, msg.ProposalID, msg.Voter, msg.VoteType)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Voter.String()),
		),
	)

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}
