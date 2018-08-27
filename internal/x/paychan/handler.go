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
		case MsgSubmitUpdate:
			return handleMsgSubmitUpdate(ctx, k, msg)
		default:
			errMsg := "Unrecognized paychan Msg type: " + reflect.TypeOf(msg).Name()
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// Handle MsgCreate
// Leaves validation to the keeper methods.
func handleMsgCreate(ctx sdk.Context, k Keeper, msg MsgCreate) sdk.Result {
	tags, err := k.CreateChannel(ctx, msg.Participants[0], msg.Participants[len(msg.Participants)-1], msg.Coins)
	if err != nil {
		return err.Result()
	}
	// TODO any other information that should be returned in Result?
	return sdk.Result{
		Tags: tags,
	}
}

// Handle MsgSubmitUpdate
// Leaves validation to the keeper methods.
func handleMsgSubmitUpdate(ctx sdk.Context, k Keeper, msg MsgSubmitUpdate) sdk.Result {

	participants := k.getChannel(ctx, msg.Update.ChannelID).Participants

	// if only sender signed
	if msg.submitter == participants[0] {
		tags, err := k.InitCloseChannelBySender()
		// else if receiver signed
	} else if msg.submitter == participants[len(participants)-1] {
		tags, err := k.CloseChannelByReceiver()
	}

	if err != nil {
		return err.Result()
	}
	// These tags can be used by clients to subscribe to channel close attempts
	return sdk.Result{
		Tags: tags,
	}
}
