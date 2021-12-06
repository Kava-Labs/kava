package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/incentive/types"
)

type msgServer struct {
	keeper Keeper
}

// NewMsgServerImpl returns an implementation of the incentive MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (k msgServer) ClaimUSDXMintingReward(goCtx context.Context, msg *types.MsgClaimUSDXMintingReward) (*types.MsgClaimUSDXMintingRewardResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	err = k.keeper.ClaimUSDXMintingReward(ctx, sender, sender, msg.MultiplierName)
	if err != nil {
		return nil, err
	}

	return &types.MsgClaimUSDXMintingRewardResponse{}, nil
}

func (k msgServer) ClaimUSDXMintingRewardVVesting(goCtx context.Context, msg *types.MsgClaimUSDXMintingRewardVVesting) (*types.MsgClaimUSDXMintingRewardVVestingResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// if err := k.keeper.ValidateIsValidatorVestingAccount(ctx, msg.Sender); err != nil {
	//	return nil, err
	// }

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	receiver, err := sdk.AccAddressFromBech32(msg.Receiver)
	if err != nil {
		return nil, err
	}

	err = k.keeper.ClaimUSDXMintingReward(ctx, sender, receiver, msg.MultiplierName)
	if err != nil {
		return nil, err
	}

	return &types.MsgClaimUSDXMintingRewardVVestingResponse{}, nil
}

func (k msgServer) ClaimHardReward(goCtx context.Context, msg *types.MsgClaimHardReward) (*types.MsgClaimHardRewardResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	for _, selection := range msg.DenomsToClaim {
		err := k.keeper.ClaimHardReward(ctx, sender, sender, selection.Denom, selection.MultiplierName)
		if err != nil {
			return nil, err
		}

	}

	return &types.MsgClaimHardRewardResponse{}, nil
}

func (k msgServer) ClaimHardRewardVVesting(goCtx context.Context, msg *types.MsgClaimHardRewardVVesting) (*types.MsgClaimHardRewardVVestingResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// if err := k.keeper.ValidateIsValidatorVestingAccount(ctx, msg.Sender); err != nil {
	//	return nil, err
	// }

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	receiver, err := sdk.AccAddressFromBech32(msg.Receiver)
	if err != nil {
		return nil, err
	}

	for _, selection := range msg.DenomsToClaim {
		err := k.keeper.ClaimHardReward(ctx, sender, receiver, selection.Denom, selection.MultiplierName)
		if err != nil {
			return nil, err
		}

	}

	return &types.MsgClaimHardRewardVVestingResponse{}, nil
}

func (k msgServer) ClaimDelegatorReward(goCtx context.Context, msg *types.MsgClaimDelegatorReward) (*types.MsgClaimDelegatorRewardResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	for _, selection := range msg.DenomsToClaim {
		err := k.keeper.ClaimDelegatorReward(ctx, sender, sender, selection.Denom, selection.MultiplierName)
		if err != nil {
			return nil, err
		}
	}

	return &types.MsgClaimDelegatorRewardResponse{}, nil
}

func (k msgServer) ClaimDelegatorRewardVVesting(goCtx context.Context, msg *types.MsgClaimDelegatorRewardVVesting) (*types.MsgClaimDelegatorRewardVVestingResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// if err := k.keeper.ValidateIsValidatorVestingAccount(ctx, msg.Sender); err != nil {
	//	return nil, err
	// }

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	receiver, err := sdk.AccAddressFromBech32(msg.Receiver)
	if err != nil {
		return nil, err
	}

	for _, selection := range msg.DenomsToClaim {
		err := k.keeper.ClaimDelegatorReward(ctx, sender, receiver, selection.Denom, selection.MultiplierName)
		if err != nil {
			return nil, err
		}
	}

	return &types.MsgClaimDelegatorRewardVVestingResponse{}, nil
}

func (k msgServer) ClaimSwapReward(goCtx context.Context, msg *types.MsgClaimSwapReward) (*types.MsgClaimSwapRewardResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	for _, selection := range msg.DenomsToClaim {
		err := k.keeper.ClaimSwapReward(ctx, sender, sender, selection.Denom, selection.MultiplierName)
		if err != nil {
			return nil, err
		}
	}

	return &types.MsgClaimSwapRewardResponse{}, nil
}

func (k msgServer) ClaimSwapRewardVVesting(goCtx context.Context, msg *types.MsgClaimSwapRewardVVesting) (*types.MsgClaimSwapRewardVVestingResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// if err := k.keeper.ValidateIsValidatorVestingAccount(ctx, msg.Sender); err != nil {
	//	return nil, err
	// }

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	receiver, err := sdk.AccAddressFromBech32(msg.Receiver)
	if err != nil {
		return nil, err
	}

	for _, selection := range msg.DenomsToClaim {
		err := k.keeper.ClaimSwapReward(ctx, sender, receiver, selection.Denom, selection.MultiplierName)
		if err != nil {
			return nil, err
		}
	}

	return &types.MsgClaimSwapRewardVVestingResponse{}, nil
}
