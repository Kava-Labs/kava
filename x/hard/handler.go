package hard

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/hard/keeper"
	"github.com/kava-labs/kava/x/hard/types"
)

// NewHandler creates an sdk.Handler for hard messages
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case types.MsgDeposit:
			return handleMsgDeposit(ctx, k, msg)
		case types.MsgWithdraw:
			return handleMsgWithdraw(ctx, k, msg)
		case types.MsgBorrow:
			return handleMsgBorrow(ctx, k, msg)
		case types.MsgRepay:
			return handleMsgRepay(ctx, k, msg)
		case types.MsgLiquidate:
			return handleMsgLiquidate(ctx, k, msg)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized %s message type: %T", ModuleName, msg)
		}
	}
}

func handleMsgDeposit(ctx sdk.Context, k keeper.Keeper, msg types.MsgDeposit) (*sdk.Result, error) {
	err := k.Deposit(ctx, msg.Depositor, msg.Amount)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Depositor.String()),
		),
	)
	return &sdk.Result{
		Events: ctx.EventManager().Events(),
	}, nil
}

func handleMsgWithdraw(ctx sdk.Context, k keeper.Keeper, msg types.MsgWithdraw) (*sdk.Result, error) {
	err := k.Withdraw(ctx, msg.Depositor, msg.Amount)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Depositor.String()),
		),
	)
	return &sdk.Result{
		Events: ctx.EventManager().Events(),
	}, nil
}

func handleMsgBorrow(ctx sdk.Context, k keeper.Keeper, msg types.MsgBorrow) (*sdk.Result, error) {
	err := k.Borrow(ctx, msg.Borrower, msg.Amount)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Borrower.String()),
		),
	)
	return &sdk.Result{
		Events: ctx.EventManager().Events(),
	}, nil
}

func handleMsgRepay(ctx sdk.Context, k keeper.Keeper, msg types.MsgRepay) (*sdk.Result, error) {
	err := k.Repay(ctx, msg.Sender, msg.Amount)
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

func handleMsgLiquidate(ctx sdk.Context, k keeper.Keeper, msg types.MsgLiquidate) (*sdk.Result, error) {
	_, err := k.AttemptKeeperLiquidation(ctx, msg.Keeper, msg.Borrower)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Keeper.String()),
		),
	)
	return &sdk.Result{
		Events: ctx.EventManager().Events(),
	}, nil
}
