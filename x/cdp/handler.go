package cdp

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewHandler creates an sdk.Handler for cdp messages
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
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
			errMsg := fmt.Sprintf("unrecognized cdp msg type: %T", msg)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleMsgCreateCDP(ctx sdk.Context, k Keeper, msg MsgCreateCDP) sdk.Result {
	err := k.AddCdp(ctx, msg.Sender, msg.Collateral, msg.Principal)
	if err != nil {
		return err.Result()
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
		),
	)
	id, _ := k.GetCdpID(ctx, msg.Sender, msg.Collateral[0].Denom)

	return sdk.Result{
		Data:   GetCdpIDBytes(id),
		Events: ctx.EventManager().Events(),
	}
}

func handleMsgDeposit(ctx sdk.Context, k Keeper, msg MsgDeposit) sdk.Result {
	err := k.DepositCollateral(ctx, msg.Owner, msg.Depositor, msg.Collateral)
	if err != nil {
		return err.Result()
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Depositor.String()),
		),
	)
	return sdk.Result{Events: ctx.EventManager().Events()}
}

func handleMsgWithdraw(ctx sdk.Context, k Keeper, msg MsgWithdraw) sdk.Result {
	err := k.WithdrawCollateral(ctx, msg.Owner, msg.Depositor, msg.Collateral)
	if err != nil {
		return err.Result()
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Depositor.String()),
		),
	)
	return sdk.Result{Events: ctx.EventManager().Events()}
}

func handleMsgDrawDebt(ctx sdk.Context, k Keeper, msg MsgDrawDebt) sdk.Result {
	err := k.AddPrincipal(ctx, msg.Sender, msg.CdpDenom, msg.Principal)
	if err != nil {
		return err.Result()
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
		),
	)
	return sdk.Result{Events: ctx.EventManager().Events()}
}

func handleMsgRepayDebt(ctx sdk.Context, k Keeper, msg MsgRepayDebt) sdk.Result {
	err := k.RepayPrincipal(ctx, msg.Sender, msg.CdpDenom, msg.Payment)
	if err != nil {
		return err.Result()
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
		),
	)
	return sdk.Result{Events: ctx.EventManager().Events()}
}
