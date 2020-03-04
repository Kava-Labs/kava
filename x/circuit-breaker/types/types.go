package types

// TODO implement a gov proposal for adding a route to the circuit breaker keeper.

type CircuitBreakProposal struct {
	MsgRoutes []MsgRoute
}

type MsgRoute struct {
	Route string
	Msg   sdk.Msg // how best to store a Msg type? as a string?
}

// TODO gov.Proposal methods...
