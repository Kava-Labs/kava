package bep3

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/bep3/keeper"
	"github.com/kava-labs/kava/x/bep3/types"
)

type msgServer struct {
	keeper keeper.Keeper
}

// NewMsgServerImpl returns an implementation of the swap MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper keeper.Keeper) types.MsgServer {
	return &msgServer{keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (k msgServer) CreateAtomicSwap(goCtx context.Context, msg *types.MsgCreateAtomicSwap) (*types.MsgCreateAtomicSwapResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	from, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		return nil, err
	}
	to, err := sdk.AccAddressFromBech32(msg.To)
	if err != nil {
		return nil, err
	}

	if err = k.keeper.CreateAtomicSwap(ctx, msg.RandomNumberHash, msg.Timestamp, msg.HeightSpan,
		from, to, msg.SenderOtherChain, msg.RecipientOtherChain, msg.Amount, true); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.From),
		),
	)

	return &types.MsgCreateAtomicSwapResponse{}, nil
}

func (k msgServer) ClaimAtomicSwap(goCtx context.Context, msg *types.MsgClaimAtomicSwap) (*types.MsgClaimAtomicSwapResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	from, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		return nil, err
	}

	err = k.keeper.ClaimAtomicSwap(ctx, from, msg.SwapID, msg.RandomNumber)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.From),
		),
	)

	return &types.MsgClaimAtomicSwapResponse{}, nil
}

func (k msgServer) RefundAtomicSwap(goCtx context.Context, msg *types.MsgRefundAtomicSwap) (*types.MsgRefundAtomicSwapResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	from, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		return nil, err
	}

	err = k.keeper.RefundAtomicSwap(ctx, from, msg.SwapID)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.From),
		),
	)

	return &types.MsgRefundAtomicSwapResponse{}, nil
}
