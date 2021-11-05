package pricefeed

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/pricefeed/keeper"
	"github.com/kava-labs/kava/x/pricefeed/types"
)

type msgServer struct {
	keeper keeper.Keeper
}

// NewMsgServerImpl returns an implementation of the pricefeed MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper keeper.Keeper) types.MsgServer {
	return &msgServer{keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (k msgServer) PostPrice(goCtx context.Context, msg *types.MsgPostPrice) (*types.MsgPostPriceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	from, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		return nil, err
	}

	_, err = k.keeper.GetOracle(ctx, msg.MarketID, from)
	if err != nil {
		return nil, err
	}

	_, err = k.keeper.SetPrice(ctx, from, msg.MarketID, msg.Price, msg.Expiry)
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

	return &types.MsgPostPriceResponse{}, nil
}
