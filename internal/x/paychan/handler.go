package paychan

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"reflect"
)

// Called when adding routes to a newly created app.
// NewHandler returns a handler for "paychan" type messages.
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgSend:
			return handleMsgSend(ctx, k, msg)
		case MsgIssue:
			return handleMsgIssue(ctx, k, msg)
		default:
			errMsg := "Unrecognized paychan Msg type: " + reflect.TypeOf(msg).Name()
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// Handle CreateMsg.
func handleMsgCreate(ctx sdk.Context, k Keeper, msg MsgSend) sdk.Result {
	// k.CreatePaychan(args...)
	// handle erros
}

// Handle CloseMsg.
func handleMsgClose(ctx sdk.Context, k Keeper, msg MsgIssue) sdk.Result {
	// k.ClosePaychan(args...)
	// handle errors
}
