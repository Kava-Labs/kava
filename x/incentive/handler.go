package incentive

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/incentive/keeper"
	"github.com/kava-labs/kava/x/incentive/types"
)

// NewHandler creates an sdk.Handler for incentive module messages
func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case types.MsgClaimUSDXMintingReward:
			return handleMsgClaimUSDXMintingReward(ctx, k, msg)
		case types.MsgClaimUSDXMintingRewardVVesting:
			return handleMsgClaimUSDXMintingRewardVVesting(ctx, k, msg)
		case types.MsgClaimHardReward:
			return handleMsgClaimHardReward(ctx, k, msg)
		case types.MsgClaimHardRewardVVesting:
			return handleMsgClaimHardRewardVVesting(ctx, k, msg)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized %s message type: %T", ModuleName, msg)
		}
	}
}

func handleMsgClaimUSDXMintingReward(ctx sdk.Context, k keeper.Keeper, msg types.MsgClaimUSDXMintingReward) (*sdk.Result, error) {
	err := k.ClaimUSDXMintingReward(ctx, msg.Sender, types.MultiplierName(msg.MultiplierName))
	if err != nil {
		return nil, err
	}
	return &sdk.Result{
		Events: ctx.EventManager().Events(),
	}, nil
}

func handleMsgClaimUSDXMintingRewardVVesting(ctx sdk.Context, k keeper.Keeper, msg types.MsgClaimUSDXMintingRewardVVesting) (*sdk.Result, error) {

	err := k.ClaimUSDXMintingRewardVVesting(ctx, msg.Sender, msg.Receiver, types.MultiplierName(msg.MultiplierName))
	if err != nil {
		return nil, err
	}
	return &sdk.Result{
		Events: ctx.EventManager().Events(),
	}, nil
}

func handleMsgClaimHardReward(ctx sdk.Context, k keeper.Keeper, msg types.MsgClaimHardReward) (*sdk.Result, error) {

	err := k.ClaimHardReward(ctx, msg.Sender, types.MultiplierName(msg.MultiplierName))
	if err != nil {
		return nil, err
	}
	return &sdk.Result{
		Events: ctx.EventManager().Events(),
	}, nil
}

func handleMsgClaimHardRewardVVesting(ctx sdk.Context, k keeper.Keeper, msg types.MsgClaimHardRewardVVesting) (*sdk.Result, error) {

	err := k.ClaimHardRewardVVesting(ctx, msg.Sender, msg.Receiver, types.MultiplierName(msg.MultiplierName))
	if err != nil {
		return nil, err
	}
	return &sdk.Result{
		Events: ctx.EventManager().Events(),
	}, nil
}
