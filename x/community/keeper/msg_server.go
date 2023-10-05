package keeper

import (
	"context"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/kava-labs/kava/x/community/types"
)

type msgServer struct {
	keeper Keeper
}

// NewMsgServerImpl returns an implementation of the community MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// FundCommunityPool handles FundCommunityPool msgs.
func (s msgServer) FundCommunityPool(goCtx context.Context, msg *types.MsgFundCommunityPool) (*types.MsgFundCommunityPoolResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// above validation will fail if depositor is invalid
	depositor := sdk.MustAccAddressFromBech32(msg.Depositor)

	if err := s.keeper.FundCommunityPool(ctx, depositor, msg.Amount); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueFundCommunityPool),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Depositor),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.String()),
		),
	)

	return &types.MsgFundCommunityPoolResponse{}, nil
}

// UpdateParams handles UpdateParams msgs.
func (s msgServer) UpdateParams(
	goCtx context.Context,
	msg *types.MsgUpdateParams,
) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if s.keeper.GetAuthority().String() != msg.Authority {
		return nil, errors.Wrapf(
			govtypes.ErrInvalidSigner,
			"invalid authority; expected %s, got %s",
			s.keeper.GetAuthority(),
			msg.Authority,
		)
	}

	if err := msg.Params.Validate(); err != nil {
		return nil, errors.Wrap(types.ErrInvalidParams, err.Error())
	}

	s.keeper.SetParams(ctx, msg.Params)

	return &types.MsgUpdateParamsResponse{}, nil
}
