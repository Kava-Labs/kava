package cdp

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewHandler creates an sdk.Handler for cdp messages
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case MsgCreateCDP:
			return handleMsgCreateCDP(ctx, k, msg)
		case MsgDeposit:
			return handleMsgDeposit(ctx, k, msg)
		case MsgWithdraw:
			return handleMsgWithdraw(ctx, k, msg)
		case MsgDrawDebt:
			return handleMsgDrawDebt(ctx, k, msg)
		case MsgRepayDebt:
			return handleMsgRepayDebt(ctx, k, msg)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized %s message type: %T", ModuleName, msg)
		}
	}
}

func handleMsgCreateCDP(ctx sdk.Context, k Keeper, msg MsgCreateCDP) (*sdk.Result, error) {
	err := k.AddCdp(ctx, msg.Sender, msg.Collateral, msg.Principal, msg.CollateralType)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
		),
	)
	id, _ := k.GetCdpID(ctx, msg.Sender, msg.CollateralType)

	return &sdk.Result{
		Data:   GetCdpIDBytes(id),
		Events: ctx.EventManager().Events(),
	}, nil
}

func handleMsgDeposit(ctx sdk.Context, k Keeper, msg MsgDeposit) (*sdk.Result, error) {
	err := k.DepositCollateral(ctx, msg.Owner, msg.Depositor, msg.Collateral, msg.CollateralType)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Depositor.String()),
		),
	)
	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

func handleMsgWithdraw(ctx sdk.Context, k Keeper, msg MsgWithdraw) (*sdk.Result, error) {
	err := k.WithdrawCollateral(ctx, msg.Owner, msg.Depositor, msg.Collateral, msg.CollateralType)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Depositor.String()),
		),
	)
	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

func handleMsgDrawDebt(ctx sdk.Context, k Keeper, msg MsgDrawDebt) (*sdk.Result, error) {
	err := k.AddPrincipal(ctx, msg.Sender, msg.CollateralType, msg.Principal)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
		),
	)
	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

func handleMsgRepayDebt(ctx sdk.Context, k Keeper, msg MsgRepayDebt) (*sdk.Result, error) {
	err := k.RepayPrincipal(ctx, msg.Sender, msg.CollateralType, msg.Payment)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
		),
	)
	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}
