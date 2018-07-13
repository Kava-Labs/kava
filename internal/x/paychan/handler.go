package paychan

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"reflect"
)

// NewHandler returns a handler for "paychan" type messages.
// Called when adding routes to a newly created app.
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgCreate:
			return handleMsgCreate(ctx, k, msg)
		case MsgClose:
			return handleMsgClose(ctx, k, msg)
		default:
			errMsg := "Unrecognized paychan Msg type: " + reflect.TypeOf(msg).Name()
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// Handle CreateMsg.
// Leaves validation to the keeper methods.
func handleMsgCreate(ctx sdk.Context, k Keeper, msg MsgCreate) sdk.Result {
	// TODO maybe remove tags for first version
	tags, err := k.CreatePaychan(ctx, msg.sender, msg.receiver, msg.amount)
	if err != nil {
		return err.Result()
	}
	// TODO any other information that should be returned in Result?
	return sdk.Result{
		Tags: tags,
	}
}

// Handle CloseMsg.
// Leaves validation to the keeper methods.
func handleMsgClose(ctx sdk.Context, k Keeper, msg MsgClose) sdk.Result {
	// TODO maybe remove tags for first version
	tags, err := k.ClosePaychan(ctx, msg.sender, msg.receiver, msg.id, msg.receiverAmount)
	if err != nil {
		return err.Result()
	}
	// These tags can be used to subscribe to channel closures
	return sdk.Result{
		Tags: tags,
	}
}
