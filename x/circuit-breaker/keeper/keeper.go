package circuit-breaker

// Keeper stores routes that have been "broken"
type Keeper struct {
}


func (k Keeper) GetRoutes(ctx sdk.Context) []string {
	// TODO
}

func (k Keeper) SetRoutes(ctx sdk.Context, routes []string) {
	// TODO
}

func (k Keeper) HandleCircuitBreakerProposal(ctx sdk.Context, c Content) {
	k.SetRoutes(ctx, c.Routes)
}