package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/liquid/types"
)

type msgServer struct {
	keeper Keeper
}

// NewMsgServerImpl returns an implementation of the liquid MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (k msgServer) MintDerivative(goCtx context.Context, msg *types.MsgMintDerivative) (*types.MsgMintDerivativeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	validator, err := sdk.ValAddressFromBech32(msg.Validator)
	if err != nil {
		return nil, err
	}

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	err = k.keeper.MintDerivative(ctx, sender, validator, msg.Amount)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Validator),
		),
	)

	return &types.MsgMintDerivativeResponse{
		// Construct coin here to avoid returning a deterministic value from MintDerivative method
		Amount: sdk.NewCoin(k.keeper.GetLiquidStakingTokenDenom(validator), msg.Amount.Amount),
	}, nil
}

func (k msgServer) BurnDerivative(goCtx context.Context, msg *types.MsgBurnDerivative) (*types.MsgBurnDerivativeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	validator, err := sdk.ValAddressFromBech32(msg.Validator)
	if err != nil {
		return nil, err
	}

	err = k.keeper.BurnDerivative(ctx, sender, validator, msg.Amount)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Validator),
		),
	)
	return &types.MsgBurnDerivativeResponse{
		Amount: msg.Amount,
	}, nil
}
