package kavadist

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/kava-labs/kava/x/kavadist/keeper"
	"github.com/kava-labs/kava/x/kavadist/types"
)

// NewKavaDistProposalsHandler handles all proposals for the x/kavadist module.
func NewKavaDistProposalsHandler(k keeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.CommunityPoolLendDepositProposal:
			return keeper.HandleCommunityPoolLendDepositProposal(ctx, k, c)
		case *types.CommunityPoolLendWithdrawProposal:
			return keeper.HandleCommunityPoolLendWithdrawProposal(ctx, k, c)
		case *types.CommunityPoolMultiSpendProposal:
			return keeper.HandleCommunityPoolMultiSpendProposal(ctx, k, c)
		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized kavadist proposal content type: %T", c)
		}
	}
}
