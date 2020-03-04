package ante

import (
	"fmt"

	"github.com/kava-labs/kava/x/shutdown/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DisableMsgDecorator errors if a tx contains a disallowed msg type and calls the next AnteHandler if all msgs are allowed
type DisableMsgDecorator struct {
	shutdownKeeper keeper.Keeper
}

func NewDisableMsgDecorator(shutdownKeeper keeper.Keeper) DisableMsgDecorator {
	return DisableMsgDecorator{
		shutdownKeeper: shutdownKeeper,
	}
}

func (dmd DisableMsgDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {

	// get msg route, error if not allowed
	disallowedRoutes := dmd.shutdownKeeper.GetMsgRoutes(ctx)
	for _, m := range tx.GetMsgs() {
		for _, r := range disallowedRoutes {
			if r.Route == m.Route() && r.Msg == m.Type() {
				return ctx, fmt.Errorf("route %s has been disabled, tx rejected", r)
			}
		}
	}
	// otherwise continue to next antehandler decorator
	return next(ctx, tx, simulate)
}
