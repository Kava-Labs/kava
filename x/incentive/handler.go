package incentive

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/incentive/keeper"
	"github.com/kava-labs/kava/x/incentive/types"
)

// NewHandler creates an sdk.Handler for incentive module messages
func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case types.MsgClaimReward:
			return handleMsgClaimReward(ctx, k, msg)
		default:
			errMsg := fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg)
			return sdk.ErrUnknownRequest(errMsg).Result()

		}
	}
}

func handleMsgClaimReward(ctx sdk.Context, k keeper.Keeper, msg types.MsgClaimReward) sdk.Result {
	foundClaim := false
	var txError sdk.Error
	k.IterateClaimPeriods(ctx, func(cp types.ClaimPeriod) (stop bool) {
		if cp.Denom == msg.Denom {
			_, found := k.GetClaim(ctx, msg.Sender, cp.Denom, cp.ID)
			if found {
				foundClaim = true
				err := k.PayoutClaim(ctx, msg.Sender, cp.Denom, cp.ID)
				if err != nil {
					txError = err
					return true
				}
				ctx.EventManager().EmitEvent(
					sdk.NewEvent(
						sdk.EventTypeMessage,
						sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
						sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
					),
				)
			}
		}
		return false
	})
	if txError != nil {
		return txError.Result()
	}
	if !foundClaim {
		return types.ErrNoClaimsFound(k.Codespace(), msg.Sender, msg.Denom).Result()
	}
	return sdk.Result{
		Events: ctx.EventManager().Events(),
	}
}
