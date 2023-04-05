package community

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/kava-labs/kava/x/community/keeper"
	"github.com/kava-labs/kava/x/community/types"
)

// NewCommunityPoolProposalHandler handles x/community proposals.
func NewCommunityPoolProposalHandler(k keeper.Keeper) govv1beta1.Handler {
	return func(ctx sdk.Context, content govv1beta1.Content) error {
		switch c := content.(type) {
		case *types.CommunityPoolLendDepositProposal:
			return keeper.HandleCommunityPoolLendDepositProposal(ctx, k, c)
		case *types.CommunityPoolLendWithdrawProposal:
			return keeper.HandleCommunityPoolLendWithdrawProposal(ctx, k, c)
		default:
			return errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized community proposal content type: %T", c)
		}
	}
}
