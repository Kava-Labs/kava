package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/shutdown/types"
)

// Keeper stores routes that have been "broken"
type Keeper struct {
}

func (k Keeper) GetMsgRoutes(ctx sdk.Context) []types.MsgRoute {
	// TODO
	return []types.MsgRoute{}
}

func (k Keeper) SetMsgRoutes(ctx sdk.Context, routes []types.MsgRoute) {
	// TODO
}
