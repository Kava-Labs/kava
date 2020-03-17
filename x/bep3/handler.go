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
		case MsgCreateAtomicSwap:
			return handleMsgCreateAtomicSwap(ctx, k, msg)
		case MsgClaimAtomicSwap:
			return handleMsgClaimAtomicSwap(ctx, k, msg)
		case MsgRefundAtomicSwap:
			return handleMsgRefundAtomicSwap(ctx, k, msg)
		default:
			errMsg := fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// handleMsgCreateAtomicSwap handles requests to create a new AtomicSwap
func handleMsgCreateAtomicSwap(ctx sdk.Context, k Keeper, msg types.MsgCreateAtomicSwap) sdk.Result {
	err := k.CreateAtomicSwap(ctx, msg.RandomNumberHash, msg.Timestamp, msg.HeightSpan, msg.From, msg.To,
		msg.SenderOtherChain, msg.RecipientOtherChain, msg.Amount, msg.ExpectedIncome, msg.CrossChain)

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

// handleMsgClaimAtomicSwap handles requests to claim funds in an active AtomicSwap
func handleMsgClaimAtomicSwap(ctx sdk.Context, k Keeper, msg types.MsgClaimAtomicSwap) sdk.Result {

	err := k.ClaimAtomicSwap(ctx, msg.From, msg.SwapID, msg.RandomNumber)
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

// handleMsgRefundAtomicSwap handles requests to refund an active AtomicSwap
func handleMsgRefundAtomicSwap(ctx sdk.Context, k Keeper, msg types.MsgRefundAtomicSwap) sdk.Result {

	err := k.RefundAtomicSwap(ctx, msg.From, msg.SwapID)
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
