package shutdown

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"

	"github.com/kava-labs/kava/x/shutdown/keeper"
	"github.com/kava-labs/kava/x/shutdown/types"
)

func NewShutdownProposalHandler(k keeper.Keeper) gov.Handler {
	return func(ctx sdk.Context, content gov.Content) sdk.Error {
		switch c := content.(type) {
		case types.ShutdownProposal:
			return handleShutdownProposal(ctx, k, c)

		default:
			errMsg := fmt.Sprintf("unrecognized %s proposal content type: %T", types.ModuleName, c)
			return sdk.ErrUnknownRequest(errMsg)
		}
	}
}

func handleShutdownProposal(ctx sdk.Context, k keeper.Keeper, c types.ShutdownProposal) sdk.Error {
	// TODO validate proposal
	k.SetMsgRoutes(ctx, c.MsgRoutes)
	return nil
}
