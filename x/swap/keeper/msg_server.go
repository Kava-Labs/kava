package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/swap/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the swap MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (k msgServer) Deposit(goCtx context.Context, msg *types.MsgDeposit) (*types.MsgDepositResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := checkDeadline(ctx, msg); err != nil {
		return nil, err
	}

	depositor, err := sdk.AccAddressFromBech32(msg.Depositor)
	if err != nil {
		return nil, err
	}

	if err := k.DepositLiquidity(ctx, depositor, msg.TokenA, msg.TokenB, msg.Slippage); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, depositor.String()),
		),
	)

	return &types.MsgDepositResponse{}, nil
}

func (k msgServer) Withdraw(goCtx context.Context, msg *types.MsgWithdraw) (*types.MsgWithdrawResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := checkDeadline(ctx, msg); err != nil {
		return nil, err
	}

	from, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		return nil, err
	}

	if err := k.WithdrawLiquidity(ctx, from, msg.Shares, msg.MinTokenA, msg.MinTokenB); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, from.String()),
		),
	)

	return &types.MsgWithdrawResponse{}, nil
}

func (k msgServer) SwapExactForTokens(goCtx context.Context, msg *types.MsgSwapExactForTokens) (*types.MsgSwapExactForTokensResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := checkDeadline(ctx, msg); err != nil {
		return nil, err
	}

	requester, err := sdk.AccAddressFromBech32(msg.Requester)
	if err != nil {
		return nil, err
	}

	if err := k.ExecuteSwapExactForTokens(ctx, requester, msg.ExactTokenA, msg.TokenB, msg.Slippage); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, requester.String()),
		),
	)

	return &types.MsgSwapExactForTokensResponse{}, nil
}

func (k msgServer) SwapForExactTokens(goCtx context.Context, msg *types.MsgSwapForExactTokens) (*types.MsgSwapForExactTokensResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := checkDeadline(ctx, msg); err != nil {
		return nil, err
	}

	requester, err := sdk.AccAddressFromBech32(msg.Requester)
	if err != nil {
		return nil, err
	}

	if err := k.ExecuteSwapForExactTokens(ctx, requester, msg.TokenA, msg.ExactTokenB, msg.Slippage); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, requester.String()),
		),
	)

	return &types.MsgSwapForExactTokensResponse{}, nil
}

// checkDeadline returns an error if block time exceeds an included deadline
func checkDeadline(ctx sdk.Context, msg sdk.Msg) error {
	deadlineMsg, ok := msg.(types.MsgWithDeadline)
	if !ok {
		return nil
	}

	if deadlineMsg.DeadlineExceeded(ctx.BlockTime()) {
		return sdkerrors.Wrapf(
			types.ErrDeadlineExceeded,
			"block time %d >= deadline %d",
			ctx.BlockTime().Unix(),
			deadlineMsg.GetDeadline().Unix(),
		)
	}

	return nil
}
