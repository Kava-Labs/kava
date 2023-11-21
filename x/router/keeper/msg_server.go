package keeper

import (
	"context"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

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

// MintDeposit converts a delegation into staking derivatives and deposits it all into an earn vault
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

// DelegateMintDeposit delegates tokens to a validator, then converts them into staking derivatives,
// then deposits to an earn vault.
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
		return nil, stakingtypes.ErrNoValidatorFound
	}
	bondDenom := m.keeper.stakingKeeper.BondDenom(ctx)
	if msg.Amount.Denom != bondDenom {
		return nil, errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest, "invalid coin denomination: got %s, expected %s", msg.Amount.Denom, bondDenom,
		)
	}
	newShares, err := m.keeper.stakingKeeper.Delegate(ctx, depositor, msg.Amount.Amount, stakingtypes.Unbonded, validator, true)
	if err != nil {
		return nil, err
	}

	derivativeMinted, err := m.keeper.liquidKeeper.MintDerivative(ctx, depositor, valAddr, msg.Amount)
	if err != nil {
		return nil, err
	}

	err = m.keeper.earnKeeper.Deposit(ctx, depositor, derivativeMinted, earntypes.STRATEGY_TYPE_SAVINGS)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			stakingtypes.EventTypeDelegate,
			sdk.NewAttribute(stakingtypes.AttributeKeyValidator, valAddr.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.String()),
			sdk.NewAttribute(stakingtypes.AttributeKeyNewShares, newShares.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeySender, depositor.String()),
		),
	})

	return &types.MsgDelegateMintDepositResponse{}, nil
}

// WithdrawBurn removes staking derivatives from an earn vault and converts them back to a staking delegation.
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

	withdrawnAmount, err := m.keeper.earnKeeper.Withdraw(ctx, depositor, tokenAmount, earntypes.STRATEGY_TYPE_SAVINGS)
	if err != nil {
		return nil, err
	}

	_, err = m.keeper.liquidKeeper.BurnDerivative(ctx, depositor, val, withdrawnAmount)
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

// WithdrawBurnUndelegate removes staking derivatives from an earn vault, converts them to a staking delegation,
// then undelegates them from their validator.
func (m msgServer) WithdrawBurnUndelegate(goCtx context.Context, msg *types.MsgWithdrawBurnUndelegate) (*types.MsgWithdrawBurnUndelegateResponse, error) {
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

	withdrawnAmount, err := m.keeper.earnKeeper.Withdraw(ctx, depositor, tokenAmount, earntypes.STRATEGY_TYPE_SAVINGS)
	if err != nil {
		return nil, err
	}

	sharesReturned, err := m.keeper.liquidKeeper.BurnDerivative(ctx, depositor, val, withdrawnAmount)
	if err != nil {
		return nil, err
	}

	completionTime, err := m.keeper.stakingKeeper.Undelegate(ctx, depositor, val, sharesReturned)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			stakingtypes.EventTypeUnbond,
			sdk.NewAttribute(stakingtypes.AttributeKeyValidator, val.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.String()),
			sdk.NewAttribute(stakingtypes.AttributeKeyCompletionTime, completionTime.Format(time.RFC3339)),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeySender, depositor.String()),
		),
	})
	return &types.MsgWithdrawBurnUndelegateResponse{}, nil
}
