package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	earntypes "github.com/kava-labs/kava/x/earn/types"
	"github.com/kava-labs/kava/x/router/types"
)

type msgServer struct {
	keeper Keeper
}

// NewMsgServerImpl returns an implementation of the module's MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (m msgServer) MintDeposit(goCtx context.Context, msg *types.MsgMintDeposit) (*types.MsgMintDepositResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	depositor, err := sdk.AccAddressFromBech32(msg.Depositor)
	if err != nil {
		return nil, err
	}
	val, err := sdk.ValAddressFromBech32(msg.Validator)
	if err != nil {
		return nil, err
	}

	derivative, err := m.keeper.liquidKeeper.MintDerivative(ctx, depositor, val, msg.Amount)
	if err != nil {
		return nil, err
	}
	err = m.keeper.earnKeeper.Deposit(ctx, depositor, derivative, earntypes.STRATEGY_TYPE_SAVINGS)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeySender, depositor.String()),
		),
	)
	return &types.MsgMintDepositResponse{}, nil
}

func (m msgServer) DelegateMintDeposit(goCtx context.Context, msg *types.MsgDelegateMintDeposit) (*types.MsgDelegateMintDepositResponse, error) {
	return &types.MsgDelegateMintDepositResponse{}, nil
}

func (m msgServer) WithdrawBurn(goCtx context.Context, msg *types.MsgWithdrawBurn) (*types.MsgWithdrawBurnResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	depositor, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		return nil, err
	}
	val, err := sdk.ValAddressFromBech32(msg.Validator)
	if err != nil {
		return nil, err
	}

	tokenAmount, err := m.keeper.liquidKeeper.DerivativeFromTokens(ctx, val, msg.Amount)
	if err != nil {
		return nil, err
	}

	err = m.keeper.earnKeeper.Withdraw(ctx, depositor, tokenAmount, earntypes.STRATEGY_TYPE_SAVINGS)
	if err != nil {
		return nil, err
	}

	// exact bkava burned, but can leave dust delegation in module account (not a big problem).
	_, err = m.keeper.liquidKeeper.BurnDerivative(ctx, depositor, val, tokenAmount)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeySender, depositor.String()),
		),
	)

	return &types.MsgWithdrawBurnResponse{}, nil
}

func (m msgServer) WithdrawBurnUndelegate(goCtx context.Context, msg *types.MsgWithdrawBurnUndelegate) (*types.MsgWithdrawBurnUndelegateResponse, error) {
	return &types.MsgWithdrawBurnUndelegateResponse{}, nil
}
