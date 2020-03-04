package circuit-breaker

import (
	"github.com/cosmos/cosmos-sdk/x/gov"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/circuit-breaker/types"
	"github.com/kava-labs/kava/x/circuit-breaker/keeper"
)

func NewCircuitBreakerProposalHandler(k Keeper) gov.Handler {
	return func(ctx sdk.Context, content gov.Content) sdk.Error {
		switch c := content.(type) {
		case types.CircuitBreakerProposal:
			return keeper.HandleCircuitBreakerProposal(ctx, k, c)

		default:
			errMsg := fmt.Sprintf("unrecognized circuit-breaker proposal content type: %T", c)
			return sdk.ErrUnknownRequest(errMsg)
		}
	}
}