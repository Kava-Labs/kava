package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/kavadist/types"
)

// HandleCommunityPoolMultiSpendProposal is a handler for executing a passed community multi-spend proposal
func HandleCommunityPoolMultiSpendProposal(ctx sdk.Context, k Keeper, p *types.CommunityPoolMultiSpendProposal) error {
	for _, receiverInfo := range p.RecipientList {
		if k.blacklistedAddrs[receiverInfo.Address] {
			return errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "%s is blacklisted from receiving external funds", receiverInfo.Address)
		}
		err := k.distKeeper.DistributeFromFeePool(ctx, receiverInfo.Amount, receiverInfo.GetAddress())
		if err != nil {
			return err
		}
	}

	return nil
}
