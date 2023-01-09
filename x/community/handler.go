package community

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/kava-labs/kava/x/community/keeper"
	"github.com/kava-labs/kava/x/community/types"
)

// NewCommunityPoolProposalHandler handles x/community proposals.
func NewCommunityPoolProposalHandler(k keeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.CommunityPoolLendDepositProposal:
			return keeper.HandleCommunityPoolLendDepositProposal(ctx, k, c)
		case *types.CommunityPoolLendWithdrawProposal:
			return keeper.HandleCommunityPoolLendWithdrawProposal(ctx, k, c)
		case *types.CommunityPoolProposal:
			return keeper.HandleCommunityPoolProposal(ctx, k, c)
		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized community proposal content type: %T", c)
		}
	}
}
