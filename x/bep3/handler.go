package bep3

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewHandler creates an sdk.Handler for all the bep3 type messages
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case MsgCreateAtomicSwap:
			return handleMsgCreateAtomicSwap(ctx, k, msg)
		case MsgClaimAtomicSwap:
			return handleMsgClaimAtomicSwap(ctx, k, msg)
		case MsgRefundAtomicSwap:
			return handleMsgRefundAtomicSwap(ctx, k, msg)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized %s message type: %T", ModuleName, msg)
		}
	}
}

// handleMsgCreateAtomicSwap handles requests to create a new AtomicSwap
func handleMsgCreateAtomicSwap(ctx sdk.Context, k Keeper, msg MsgCreateAtomicSwap) (*sdk.Result, error) {
	err := k.CreateAtomicSwap(ctx, msg.RandomNumberHash, msg.Timestamp, msg.HeightSpan, msg.From, msg.To,
		msg.SenderOtherChain, msg.RecipientOtherChain, msg.Amount, true)

	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.From.String()),
		),
	)

	return &sdk.Result{
		Events: ctx.EventManager().Events(),
	}, nil
}

// handleMsgClaimAtomicSwap handles requests to claim funds in an active AtomicSwap
func handleMsgClaimAtomicSwap(ctx sdk.Context, k Keeper, msg MsgClaimAtomicSwap) (*sdk.Result, error) {

	err := k.ClaimAtomicSwap(ctx, msg.From, msg.SwapID, msg.RandomNumber)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.From.String()),
		),
	)

	return &sdk.Result{
		Events: ctx.EventManager().Events(),
	}, nil
}

// handleMsgRefundAtomicSwap handles requests to refund an active AtomicSwap
func handleMsgRefundAtomicSwap(ctx sdk.Context, k Keeper, msg MsgRefundAtomicSwap) (*sdk.Result, error) {

	err := k.RefundAtomicSwap(ctx, msg.From, msg.SwapID)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.From.String()),
		),
	)

	return &sdk.Result{
		Events: ctx.EventManager().Events(),
	}, nil
}
