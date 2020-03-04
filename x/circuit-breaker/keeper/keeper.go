package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/circuit-breaker/types"
)

// Keeper stores routes that have been "broken"
type Keeper struct {
}

func (k Keeper) GetMsgRoutes(ctx sdk.Context) []types.MsgRoute {
	// TODO
}

func (k Keeper) SetMsgRoutes(ctx sdk.Context, routes []types.MsgRoute) {
	// TODO
}

func HandleCircuitBreakerProposal(ctx sdk.Context, k Keeper, c types.CircuitBreakProposal) {
	k.SetMsgRoutes(ctx, c.MsgRoutes)
}
