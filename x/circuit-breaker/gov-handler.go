package circuit-breaker

func NewCircuitBreakerProposalHandler(k Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) sdk.Error {
		switch c := content.(type) {
		case types.CircuitBreakerProposal:
			return keeper.HandleCircuitBreakerProposal(ctx, k, c)

		default:
			errMsg := fmt.Sprintf("unrecognized circuit-breaker proposal content type: %T", c)
			return sdk.ErrUnknownRequest(errMsg)
		}
	}
}