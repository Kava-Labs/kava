package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/auction/types"
)

type msgServer struct {
	keeper Keeper
}

// NewMsgServerImpl returns an implementation of the swap MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (k msgServer) PlaceBid(goCtx context.Context, msg *types.MsgPlaceBid) (*types.MsgPlaceBidResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	bidder, err := sdk.AccAddressFromBech32(msg.Bidder)
	if err != nil {
		return nil, err
	}

	err = k.keeper.PlaceBid(ctx, msg.AuctionId, bidder, msg.Amount)
	if err != nil {
		return nil, err
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Bidder),
		),
	)
	return &types.MsgPlaceBidResponse{}, nil
}
