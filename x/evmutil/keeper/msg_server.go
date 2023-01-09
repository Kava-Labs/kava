package keeper

import (
	"context"
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/kava-labs/kava/x/evmutil/types"
)

type msgServer struct {
	keeper Keeper
}

// NewMsgServerImpl returns an implementation of the evmutil MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// ConvertCoinToERC20 handles a MsgConvertCoinToERC20 message to convert
// sdk.Coin to Kava EVM tokens.
func (s msgServer) ConvertCoinToERC20(
	goCtx context.Context,
	msg *types.MsgConvertCoinToERC20,
) (*types.MsgConvertCoinToERC20Response, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	initiator, err := sdk.AccAddressFromBech32(msg.Initiator)
	if err != nil {
		return nil, fmt.Errorf("invalid Initiator address: %w", err)
	}

	receiver, err := types.NewInternalEVMAddressFromString(msg.Receiver)
	if err != nil {
		return nil, fmt.Errorf("invalid Receiver address: %w", err)
	}

	if err := s.keeper.ConvertCoinToERC20(
		ctx,
		initiator,
		receiver,
		*msg.Amount,
	); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Initiator),
		),
	)

	return &types.MsgConvertCoinToERC20Response{}, nil
}

// ConvertERC20ToCoin handles a MsgConvertERC20ToCoin message to convert
// sdk.Coin to Kava EVM tokens.
func (s msgServer) ConvertERC20ToCoin(
	goCtx context.Context,
	msg *types.MsgConvertERC20ToCoin,
) (*types.MsgConvertERC20ToCoinResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	initiator, err := types.NewInternalEVMAddressFromString(msg.Initiator)
	if err != nil {
		return nil, fmt.Errorf("invalid initiator address: %w", err)
	}

	receiver, err := sdk.AccAddressFromBech32(msg.Receiver)
	if err != nil {
		return nil, fmt.Errorf("invalid receiver address: %w", err)
	}

	contractAddr, err := types.NewInternalEVMAddressFromString(msg.KavaERC20Address)
	if err != nil {
		return nil, fmt.Errorf("invalid contract address: %w", err)
	}

	if err := s.keeper.ConvertERC20ToCoin(
		ctx,
		initiator,
		receiver,
		contractAddr,
		msg.Amount,
	); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Initiator),
		),
	)

	return &types.MsgConvertERC20ToCoinResponse{}, nil
}

// MsgEVMCall executes the msg data against the go-ethereum EVM.
func (k msgServer) EVMCall(goCtx context.Context, msg *types.MsgEVMCall) (*types.MsgEVMCallResponse, error) {
	expMsgAuthority := k.keeper.GetAuthority()
	if expMsgAuthority != msg.Authority {
		return nil, sdkerrors.Wrapf(types.ErrInvalidSigner, "invalid authority; expected %s, got %s", expMsgAuthority, msg.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	toAddr, err := types.NewInternalEVMAddressFromString(msg.To)
	if err != nil {
		return nil, fmt.Errorf("invalid to evm address: %w", err)
	}

	// convert empty data to valid hex string
	msgData := msg.Data
	if len(msg.Data) == 0 {
		msgData = "0x"
	}

	data, err := hexutil.Decode(msgData)
	if err != nil {
		return nil, fmt.Errorf("unable to decode msg data: %w", err)
	}
	authorityAddr, err := sdk.AccAddressFromBech32(expMsgAuthority)
	if err != nil {
		return nil, fmt.Errorf("invalid authority address: %w", err)
	}
	authorityEvmAddr := common.BytesToAddress(authorityAddr.Bytes())

	amt := msg.Amount.BigInt()
	if amt == nil {
		amt = big.NewInt(0)
	}
	_, err = k.keeper.CallEVMWithData(
		ctx,
		authorityEvmAddr,
		&toAddr,
		data,
		amt,
	)
	if err != nil {
		ctx.Logger().Debug("EVMCall failed with params: (%s, %s, %s, %d) - %w", authorityAddr, toAddr, msgData, amt, err)
		return nil, fmt.Errorf("evm call failed: %w", err)
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeEVMCall,
		sdk.NewAttribute(sdk.AttributeKeySender, authorityAddr.String()),
		sdk.NewAttribute(types.AttributeKeyEVMToAddress, toAddr.String()),
		sdk.NewAttribute(types.AttributeKeyAmount, amt.String()),
	))

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Authority),
		),
	)

	return &types.MsgEVMCallResponse{}, nil
}
