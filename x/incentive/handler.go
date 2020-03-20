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
	// TODO
	return sdk.Result{
		Events: ctx.EventManager().Events(),
	}
}
