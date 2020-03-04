package circuit-breaker

// CircuiteBreakerDecorator needs to be combined with other standard decorators (from auth) to create the app's AnteHandler.

// CircuitBreakerDecorator errors if a tx contains a disallowed msg type
// Call next AnteHandler if all msgs are allowed
type CircuitBreakerDecorator struct {
	cbk           Keeper
}

func NewCircuitBreakerDecorator(cbk Keeper) CircuitBreakerDecorator {
	return CircuitBreakerDecorator{
		cbk:           cbk,
	}
}

func (cbd CircuitBreakerDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
// TODO need to tidy up the types used to store broken routes

	// get msg route, error if not allowed
	disallowedRoutes := cbd.cbk.GetRoutes(ctx)
	requestedRoutes := tx.Msgs
	for m, _ := range tx.Msgs {
		for r, _ := range disallowedRoutes {
			if r == m.Route() {
				return ctx, fmt.Errorf("route %s has been circuit broken, tx rejected", r)
			}
		}
	}
	return next(ctx, tx, simulate)
}
