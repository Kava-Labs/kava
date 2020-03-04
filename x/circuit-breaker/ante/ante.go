package ante

import (
	"fmt"

	"github.com/kava-labs/kava/x/circuit-breaker/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CircuitBreakerDecorator needs to be combined with other standard decorators (from auth) to create the app's AnteHandler.

// CircuitBreakerDecorator errors if a tx contains a disallowed msg type
// Call next AnteHandler if all msgs are allowed
type CircuitBreakerDecorator struct {
	cbk keeper.Keeper
}

func NewCircuitBreakerDecorator(cbk keeper.Keeper) CircuitBreakerDecorator {
	return CircuitBreakerDecorator{
		cbk: cbk,
	}
}

func (cbd CircuitBreakerDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {

	// get msg route, error if not allowed
	disallowedRoutes := cbd.cbk.GetMsgRoutes(ctx)
	for _, m := range tx.GetMsgs() {
		for _, r := range disallowedRoutes {
			if r.Route == m.Route() && r.Msg == m.Type() {
				return ctx, fmt.Errorf("route %s has been circuit broken, tx rejected", r)
			}
		}
	}
	return next(ctx, tx, simulate)
}
