package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/kavadist/types"
)

// HandleCommunityPoolLendDepositProposal handles a gov proposal for depositing community pool funds into x/hard
func HandleCommunityPoolLendDepositProposal(ctx sdk.Context, k Keeper, p *types.CommunityPoolLendDepositProposal) error {
	panic("TODO: implement me")
}

// HandleCommunityPoolLendWithdrawProposal handles a gov proposal for withdrawing community pool positions in x/hard
func HandleCommunityPoolLendWithdrawProposal(ctx sdk.Context, k Keeper, p *types.CommunityPoolLendWithdrawProposal) error {
	panic("TODO: implement me")
}

// HandleCommunityPoolMultiSpendProposal is a handler for executing a passed community multi-spend proposal
func HandleCommunityPoolMultiSpendProposal(ctx sdk.Context, k Keeper, p *types.CommunityPoolMultiSpendProposal) error {
	for _, receiverInfo := range p.RecipientList {
		if k.blacklistedAddrs[receiverInfo.Address] {
			return sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "%s is blacklisted from receiving external funds", receiverInfo.Address)
		}
		err := k.distKeeper.DistributeFromFeePool(ctx, receiverInfo.Amount, receiverInfo.GetAddress())
		if err != nil {
			return err
		}
	}
	return nil
}
