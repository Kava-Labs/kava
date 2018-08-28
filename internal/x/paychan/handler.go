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
	var err sdk.Error
	tags := sdk.EmptyTags()

	// TODO refactor signer detection - move to keeper or find nicer setup
	channel, _ := k.getChannel(ctx, msg.Update.ChannelID)
	participants := channel.Participants

	// if only sender signed
	if reflect.DeepEqual(msg.submitter, participants[0]) {
		tags, err = k.InitCloseChannelBySender(ctx, msg.Update)
		// else if receiver signed
	} else if reflect.DeepEqual(msg.submitter, participants[len(participants)-1]) {
		tags, err = k.CloseChannelByReceiver(ctx, msg.Update)
	}

	if err != nil {
		return err.Result()
	}
	// These tags can be used by clients to subscribe to channel close attempts
	return sdk.Result{
		Tags: tags,
	}
}
