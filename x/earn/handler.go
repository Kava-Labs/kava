package earn

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/kava-labs/kava/x/earn/keeper"
	"github.com/kava-labs/kava/x/earn/types"
)

// NewCommunityPoolProposalHandler
func NewCommunityPoolProposalHandler(k keeper.Keeper) govv1beta1.Handler {
	return func(ctx sdk.Context, content govv1beta1.Content) error {
		switch c := content.(type) {
		case *types.CommunityPoolDepositProposal:
			return keeper.HandleCommunityPoolDepositProposal(ctx, k, c)
		case *types.CommunityPoolWithdrawProposal:
			return keeper.HandleCommunityPoolWithdrawProposal(ctx, k, c)
		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized earn proposal content type: %T", c)
		}
	}
}
