package pricefeed

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewHandler handles all pricefeed type messages
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgPostPrice:
			return HandleMsgPostPrice(ctx, k, msg)
		default:
			errMsg := fmt.Sprintf("unrecognized pricefeed message type: %T", msg)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// price feed questions:
// do proposers need to post the round in the message? If not, how do we determine the round?

// HandleMsgPostPrice handles prices posted by oracles
func HandleMsgPostPrice(
	ctx sdk.Context,
	k Keeper,
	msg MsgPostPrice) sdk.Result {

	_, err := k.GetOracle(ctx, msg.MarketID, msg.From)
	if err != nil {
		return err.Result()
	}
	_, err = k.SetPrice(ctx, msg.From, msg.MarketID, msg.Price, msg.Expiry)
	if err != nil {
		err.Result()
	}
	return sdk.Result{}
}
