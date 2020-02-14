package bep3

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/bep3/types"
)

// NewHandler creates an sdk.Handler for all the bep3 type messages
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case MsgCreateHTLT:
			return handleMsgCreateHTLT(ctx, k, msg)
		case MsgDepositHTLT:
			return handleMsgDepositHTLT(ctx, k, msg)
		case MsgClaimHTLT:
			return handleMsgClaimHTLT(ctx, k, msg)
		case MsgRefundHTLT:
			return handleMsgRefundHTLT(ctx, k, msg)
		default:
			errMsg := fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// handleMsgCreateHTLT handles requests to create a new HTLT
func handleMsgCreateHTLT(ctx sdk.Context, k Keeper, msg types.MsgCreateHTLT) sdk.Result {

	id, err := k.CreateHTLT(ctx, msg.From, msg.To, msg.RecipientOtherChain,
		msg.SenderOtherChain, msg.RandomNumberHash, msg.Timestamp, msg.Amount,
		msg.ExpectedIncome, msg.HeightSpan, msg.CrossChain)
	if err != nil {
		return err.Result()
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.From.String()),
		),
	)

	swapID, err2 := types.HexEncodedStringToBytes(id)
	if err2 != nil {
		return sdk.ErrInternal(fmt.Sprintf("could not decode swap id %x. Error: %s", id, err2)).Result()
	}

	return sdk.Result{
		Data:   swapID,
		Events: ctx.EventManager().Events(),
	}
}

// handleMsgDepositHTLT handles requests to deposit into an active HTLT
func handleMsgDepositHTLT(ctx sdk.Context, k Keeper, msg types.MsgDepositHTLT) sdk.Result {

	err := k.DepositHTLT(ctx, msg.From, msg.SwapID, msg.Amount)
	if err != nil {
		return err.Result()
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.From.String()),
		),
	)

	return sdk.Result{
		Events: ctx.EventManager().Events(),
	}
}

// handleMsgClaimHTLT handles requests to claim funds in an active HTLT
func handleMsgClaimHTLT(ctx sdk.Context, k Keeper, msg types.MsgClaimHTLT) sdk.Result {

	err := k.ClaimHTLT(ctx, msg.From, msg.SwapID, msg.RandomNumber)
	if err != nil {
		return err.Result()
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.From.String()),
		),
	)

	return sdk.Result{
		Events: ctx.EventManager().Events(),
	}
}

// handleMsgRefundHTLT handles requests to refund an active HTLT
func handleMsgRefundHTLT(ctx sdk.Context, k Keeper, msg types.MsgRefundHTLT) sdk.Result {

	err := k.RefundHTLT(ctx, msg.From, msg.SwapID)
	if err != nil {
		return err.Result()
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.From.String()),
		),
	)

	return sdk.Result{
		Events: ctx.EventManager().Events(),
	}
}
