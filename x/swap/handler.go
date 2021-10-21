package swap

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/swap/keeper"
	"github.com/kava-labs/kava/x/swap/types"
)

// NewHandler creates an sdk.Handler for swap messages
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		if deadlineMsg, ok := msg.(types.MsgWithDeadline); ok {
			if deadlineExceeded := deadlineMsg.DeadlineExceeded(ctx.BlockTime()); deadlineExceeded {
				return nil, sdkerrors.Wrapf(types.ErrDeadlineExceeded, "block time %d >= deadline %d", ctx.BlockTime().Unix(), deadlineMsg.GetDeadline().Unix())
			}
		}

		switch msg := msg.(type) {
		case types.MsgDeposit:
			return handleMsgDeposit(ctx, k, msg)
		case types.MsgWithdraw:
			return handleMsgWithdraw(ctx, k, msg)
		case types.MsgSwapExactForTokens:
			return handleMsgSwapExactForTokens(ctx, k, msg)
		case types.MsgSwapForExactTokens:
			return handleMsgSwapForExactTokens(ctx, k, msg)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized %s message type: %T", ModuleName, msg)
		}
	}
}

func handleMsgDeposit(ctx sdk.Context, k keeper.Keeper, msg types.MsgDeposit) (*sdk.Result, error) {
	if err := k.Deposit(ctx, msg.Depositor, msg.TokenA, msg.TokenB, msg.Slippage); err != nil {
		return nil, err
	}

	return resultWithMsgSender(ctx, msg.Depositor), nil
}

func handleMsgWithdraw(ctx sdk.Context, k keeper.Keeper, msg types.MsgWithdraw) (*sdk.Result, error) {
	if err := k.Withdraw(ctx, msg.From, msg.Shares, msg.MinTokenA, msg.MinTokenB); err != nil {
		return nil, err
	}

	return resultWithMsgSender(ctx, msg.From), nil
}

func handleMsgSwapExactForTokens(ctx sdk.Context, k keeper.Keeper, msg types.MsgSwapExactForTokens) (*sdk.Result, error) {
	if err := k.SwapExactForTokens(ctx, msg.Requester, msg.ExactTokenA, msg.TokenB, msg.Slippage); err != nil {
		return nil, err
	}

	return resultWithMsgSender(ctx, msg.Requester), nil
}

func handleMsgSwapForExactTokens(ctx sdk.Context, k keeper.Keeper, msg types.MsgSwapForExactTokens) (*sdk.Result, error) {
	if err := k.SwapForExactTokens(ctx, msg.Requester, msg.TokenA, msg.ExactTokenB, msg.Slippage); err != nil {
		return nil, err
	}

	return resultWithMsgSender(ctx, msg.Requester), nil
}

func resultWithMsgSender(ctx sdk.Context, sender sdk.AccAddress) *sdk.Result {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, sender.String()),
		),
	)

	return &sdk.Result{
		Events: ctx.EventManager().Events(),
	}
}
