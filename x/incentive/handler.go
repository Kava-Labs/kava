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
		case types.MsgClaimDelegatorReward:
			return handleMsgClaimDelegatorReward(ctx, k, msg)
		case types.MsgClaimDelegatorRewardVVesting:
			return handleMsgClaimDelegatorRewardVVesting(ctx, k, msg)
		case types.MsgClaimSwapReward:
			return handleMsgClaimSwapReward(ctx, k, msg)
		case types.MsgClaimSwapRewardVVesting:
			return handleMsgClaimSwapRewardVVesting(ctx, k, msg)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized %s message type: %T", ModuleName, msg)
		}
	}
}

func handleMsgClaimUSDXMintingReward(ctx sdk.Context, k keeper.Keeper, msg types.MsgClaimUSDXMintingReward) (*sdk.Result, error) {
	err := k.ClaimUSDXMintingReward(ctx, msg.Sender, msg.Sender, types.MultiplierName(msg.MultiplierName))
	if err != nil {
		return nil, err
	}
	return &sdk.Result{
		Events: ctx.EventManager().Events(),
	}, nil
}

func handleMsgClaimUSDXMintingRewardVVesting(ctx sdk.Context, k keeper.Keeper, msg types.MsgClaimUSDXMintingRewardVVesting) (*sdk.Result, error) {

	if err := k.ValidateIsValidatorVestingAccount(ctx, msg.Sender); err != nil {
		return nil, err
	}
	err := k.ClaimUSDXMintingReward(ctx, msg.Sender, msg.Receiver, types.MultiplierName(msg.MultiplierName))
	if err != nil {
		return nil, err
	}
	return &sdk.Result{
		Events: ctx.EventManager().Events(),
	}, nil
}

func handleMsgClaimHardReward(ctx sdk.Context, k keeper.Keeper, msg types.MsgClaimHardReward) (*sdk.Result, error) {

	err := k.ClaimHardReward(ctx, msg.Sender, msg.Sender, msg.DenomsToClaim)
	if err != nil {
		return nil, err
	}
	return &sdk.Result{
		Events: ctx.EventManager().Events(),
	}, nil
}

func handleMsgClaimHardRewardVVesting(ctx sdk.Context, k keeper.Keeper, msg types.MsgClaimHardRewardVVesting) (*sdk.Result, error) {

	if err := k.ValidateIsValidatorVestingAccount(ctx, msg.Sender); err != nil {
		return nil, err
	}
	err := k.ClaimHardReward(ctx, msg.Sender, msg.Receiver, msg.DenomsToClaim)
	if err != nil {
		return nil, err
	}
	return &sdk.Result{
		Events: ctx.EventManager().Events(),
	}, nil
}

func handleMsgClaimDelegatorReward(ctx sdk.Context, k keeper.Keeper, msg types.MsgClaimDelegatorReward) (*sdk.Result, error) {

	err := k.ClaimDelegatorReward(ctx, msg.Sender, msg.Sender, types.MultiplierName(msg.MultiplierName), msg.DenomsToClaim)
	if err != nil {
		return nil, err
	}
	return &sdk.Result{
		Events: ctx.EventManager().Events(),
	}, nil
}

func handleMsgClaimDelegatorRewardVVesting(ctx sdk.Context, k keeper.Keeper, msg types.MsgClaimDelegatorRewardVVesting) (*sdk.Result, error) {

	if err := k.ValidateIsValidatorVestingAccount(ctx, msg.Sender); err != nil {
		return nil, err
	}
	err := k.ClaimDelegatorReward(ctx, msg.Sender, msg.Receiver, types.MultiplierName(msg.MultiplierName), msg.DenomsToClaim)
	if err != nil {
		return nil, err
	}
	return &sdk.Result{
		Events: ctx.EventManager().Events(),
	}, nil
}

func handleMsgClaimSwapReward(ctx sdk.Context, k keeper.Keeper, msg types.MsgClaimSwapReward) (*sdk.Result, error) {

	err := k.ClaimSwapReward(ctx, msg.Sender, msg.Sender, types.MultiplierName(msg.MultiplierName), msg.DenomsToClaim)
	if err != nil {
		return nil, err
	}
	return &sdk.Result{
		Events: ctx.EventManager().Events(),
	}, nil
}

func handleMsgClaimSwapRewardVVesting(ctx sdk.Context, k keeper.Keeper, msg types.MsgClaimSwapRewardVVesting) (*sdk.Result, error) {

	if err := k.ValidateIsValidatorVestingAccount(ctx, msg.Sender); err != nil {
		return nil, err
	}
	err := k.ClaimSwapReward(ctx, msg.Sender, msg.Receiver, types.MultiplierName(msg.MultiplierName), msg.DenomsToClaim)
	if err != nil {
		return nil, err
	}
	return &sdk.Result{
		Events: ctx.EventManager().Events(),
	}, nil
}
