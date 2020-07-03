package issuance

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/issuance/keeper"
	"github.com/kava-labs/kava/x/issuance/types"
)

// NewHandler creates an sdk.Handler for issuance messages
func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case types.MsgIssueTokens:
			return handleMsgIssueTokens(ctx, k, msg)
		case types.MsgRedeemTokens:
			return handleMsgRedeemTokens(ctx, k, msg)
		case types.MsgBlockAddress:
			return handleMsgBlockAddress(ctx, k, msg)
		case types.MsgUnblockAddress:
			return handleMsgUnblockAddress(ctx, k, msg)
		case types.MsgChangePauseStatus:
			return handleMsgChangePauseStatus(ctx, k, msg)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized %s message type: %T", types.ModuleName, msg)
		}
	}
}

func handleMsgIssueTokens(ctx sdk.Context, k keeper.Keeper, msg types.MsgIssueTokens) (*sdk.Result, error) {
	err := k.IssueTokens(ctx, msg.Tokens, msg.Sender, msg.Receiver)
	if err != nil {
		return nil, err
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
		),
	)
	return &sdk.Result{
		Events: ctx.EventManager().Events(),
	}, nil
}

func handleMsgRedeemTokens(ctx sdk.Context, k keeper.Keeper, msg types.MsgRedeemTokens) (*sdk.Result, error) {
	err := k.RedeemTokens(ctx, msg.Tokens, msg.Sender)
	if err != nil {
		return nil, err
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
		),
	)
	return &sdk.Result{
		Events: ctx.EventManager().Events(),
	}, nil
}

func handleMsgBlockAddress(ctx sdk.Context, k keeper.Keeper, msg types.MsgBlockAddress) (*sdk.Result, error) {
	err := k.BlockAddress(ctx, msg.Denom, msg.Sender, msg.BlockedAddress)
	if err != nil {
		return nil, err
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
		),
	)
	return &sdk.Result{
		Events: ctx.EventManager().Events(),
	}, nil
}

func handleMsgUnblockAddress(ctx sdk.Context, k keeper.Keeper, msg types.MsgUnblockAddress) (*sdk.Result, error) {
	err := k.UnblockAddress(ctx, msg.Denom, msg.Sender, msg.Address)
	if err != nil {
		return nil, err
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
		),
	)
	return &sdk.Result{
		Events: ctx.EventManager().Events(),
	}, nil
}

func handleMsgChangePauseStatus(ctx sdk.Context, k keeper.Keeper, msg types.MsgChangePauseStatus) (*sdk.Result, error) {
	err := k.ChangePauseStatus(ctx, msg.Sender, msg.Denom, msg.Status)
	if err != nil {
		return nil, err
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
		),
	)
	return &sdk.Result{
		Events: ctx.EventManager().Events(),
	}, nil
}
