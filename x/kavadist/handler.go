package kavadist

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewHandler creates an sdk.Handler for kavadist messages
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		default:
			errMsg := fmt.Sprintf("unrecognized cdp msg type: %T", msg)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}