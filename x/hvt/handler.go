package hvt

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/hvt/keeper"
	"github.com/kava-labs/kava/x/hvt/types"
)

// NewHandler creates an sdk.Handler for kavadist messages
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case types.MsgClaimReward:
			return handleMsgClaimReward(ctx, k, msg)
		case types.MsgDeposit:
			return handleMsgDeposit(ctx, k, msg)
		case types.MsgWithdraw:
			return handleMsgWithdraw(ctx, k, msg)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized %s message type: %T", ModuleName, msg)
		}
	}
}

func handleMsgClaimReward(ctx sdk.Context, k keeper.Keeper, msg types.MsgClaimReward) (*sdk.Result, error) {
	err := k.ClaimReward(ctx, msg.Sender, msg.DepositDenom, types.DepositType(strings.ToLower(msg.DepositType)), types.RewardMultiplier(strings.ToLower(msg.RewardMultiplier)))
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

func handleMsgDeposit(ctx sdk.Context, k keeper.Keeper, msg types.MsgDeposit) (*sdk.Result, error) {
	err := k.Deposit(ctx, msg.Depositor, msg.Amount, types.DepositType(strings.ToLower(msg.DepositType)))
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
	err := k.Withdraw(ctx, msg.Depositor, msg.Amount, types.DepositType(strings.ToLower(msg.DepositType)))
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
