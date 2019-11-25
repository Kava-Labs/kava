package cdp

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Handle all cdp messages.
func NewHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgCreateOrModifyCDP:
			return handleMsgCreateOrModifyCDP(ctx, keeper, msg)
		default:
			errMsg := fmt.Sprintf("Unrecognized cdp msg type: %T", msg)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleMsgCreateOrModifyCDP(ctx sdk.Context, keeper Keeper, msg MsgCreateOrModifyCDP) sdk.Result {

	err := keeper.ModifyCDP(ctx, msg.Sender, msg.CollateralDenom, msg.CollateralChange, msg.DebtChange)
	if err != nil {
		return err.Result()
	}

	return sdk.Result{}
}
