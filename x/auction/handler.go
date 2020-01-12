package auction

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewHandler returns a function to handle all "auction" type messages.
func NewHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgPlaceBid:
			return handleMsgPlaceBid(ctx, keeper, msg)
		default:
			errMsg := fmt.Sprintf("Unrecognized auction msg type: %T", msg)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleMsgPlaceBid(ctx sdk.Context, keeper Keeper, msg MsgPlaceBid) sdk.Result {

	err := keeper.PlaceBid(ctx, msg.AuctionID, msg.Bidder, msg.Amount)
	if err != nil {
		return err.Result()
	}

	return sdk.Result{}
}
