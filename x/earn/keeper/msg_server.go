package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/kava-labs/kava/x/earn/types"
)

type msgServer struct {
	keeper Keeper
}

// NewMsgServerImpl returns an implementation of the earn MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// Deposit handles MsgDeposit messages
func (m msgServer) Deposit(goCtx context.Context, msg *types.MsgDeposit) (*types.MsgDepositResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	depositor, err := sdk.AccAddressFromBech32(msg.Depositor)
	if err != nil {
		return nil, err
	}

	if err := m.keeper.Deposit(ctx, depositor, msg.Amount, msg.Strategy); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeySender, depositor.String()),
		),
	)

	return &types.MsgDepositResponse{}, nil
}

// Withdraw handles MsgWithdraw messages
func (m msgServer) Withdraw(goCtx context.Context, msg *types.MsgWithdraw) (*types.MsgWithdrawResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	from, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		return nil, err
	}

	if err := m.keeper.Withdraw(ctx, from, msg.Amount, msg.Strategy); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeySender, from.String()),
		),
	)

	return &types.MsgWithdrawResponse{}, nil
}

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
	// should deposit all of derivative balance
	// without this msg webapp needs to calculate bkava = floor(validator_total_shares * msg.Amount / validator_total_kava) accurately
	err = m.keeper.Deposit(ctx, depositor, derivative, msg.Strategy)
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
	ctx := sdk.UnwrapSDKContext(goCtx)

	depositor, err := sdk.AccAddressFromBech32(msg.Depositor)
	if err != nil {
		return nil, err
	}
	valAddr, err := sdk.ValAddressFromBech32(msg.Validator)
	if err != nil {
		return nil, err
	}
	validator, found := m.keeper.stakingKeeper.GetValidator(ctx, valAddr)
	if !found {
		panic("TODO custom errors") // TODO
	}
	bondDenom := m.keeper.stakingKeeper.BondDenom(ctx)
	if msg.Amount.Denom != bondDenom {
		return nil, sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest, "invalid coin denomination: got %s, expected %s", msg.Amount.Denom, bondDenom,
		)
	}
	_, err = m.keeper.stakingKeeper.Delegate(ctx, depositor, msg.Amount.Amount, stakingtypes.Unbonded, validator, true)
	if err != nil {
		return nil, err
	}
	// This can leave a dust amount of shares in user's delegation.
	// MintDerivative could be modified to accept shares returned by Delegate to avoid this.
	// Could fail if we don't return accurate vested delegation balances in liquid api.
	derivativeMinted, err := m.keeper.liquidKeeper.MintDerivative(ctx, depositor, valAddr, msg.Amount)
	if err != nil {
		return nil, err
	}
	// deposit is exact
	err = m.keeper.Deposit(ctx, depositor, derivativeMinted, msg.Strategy)
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

	return &types.MsgDelegateMintDepositResponse{}, nil
}

func (m msgServer) WithdrawBurn(goCtx context.Context, msg *types.MsgWithdrawBurn) (*types.MsgWithdrawBurnResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	depositor, err := sdk.AccAddressFromBech32(msg.Depositor)
	if err != nil {
		return nil, err
	}
	val, err := sdk.ValAddressFromBech32(msg.Validator)
	if err != nil {
		return nil, err
	}

	// User specifies withdraw as kava. Needs to be converted to bkava.
	// If user is withdrawing full balance, we could round up here to avoid dust.
	tokenAmount, err := m.keeper.liquidKeeper.TokenToDerivative(ctx, val, msg.Amount.Amount)
	if err != nil {
		return nil, err
	}

	err = m.keeper.Withdraw(ctx, depositor, tokenAmount, msg.Strategy)
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
	ctx := sdk.UnwrapSDKContext(goCtx)

	depositor, err := sdk.AccAddressFromBech32(msg.Depositor)
	if err != nil {
		return nil, err
	}
	val, err := sdk.ValAddressFromBech32(msg.Validator)
	if err != nil {
		return nil, err
	}
	tokenAmount, err := m.keeper.liquidKeeper.TokenToDerivative(ctx, val, msg.Amount.Amount)
	if err != nil {
		return nil, err
	}

	err = m.keeper.Withdraw(ctx, depositor, tokenAmount, msg.Strategy)
	if err != nil {
		return nil, err
	}

	sharesReturned, err := m.keeper.liquidKeeper.BurnDerivative(ctx, depositor, val, tokenAmount)
	if err != nil {
		return nil, err
	}

	// TODO use msgServer interface? it has extra validations and events
	// exact shares undelegated
	// without this msg, the webapp needs to calculate expected shares returned from burnDerivative, and then convert to kava as MsgUndelegate uses kava.
	_, err = m.keeper.stakingKeeper.Undelegate(ctx, depositor, val, sharesReturned)
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

	return &types.MsgWithdrawBurnUndelegateResponse{}, nil
}
