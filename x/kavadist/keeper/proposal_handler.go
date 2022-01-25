package keeper

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/kavadist/types"
)

// HandleCommunityPoolMultiSpendProposal is a handler for executing a passed community multi-spend proposal
func HandleCommunityPoolMultiSpendProposal(ctx sdk.Context, k Keeper, p *types.CommunityPoolMultiSpendProposal, upgradeTime time.Time) error {
	for _, receiverInfo := range p.RecipientList {
		if k.blacklistedAddrs[receiverInfo.Address] {
			return sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "%s is blacklisted from receiving external funds", receiverInfo.Address)
		}

		if ctx.BlockTime().Before(upgradeTime) {
			panic(fmt.Sprintf("cannot submit multi-spend proposal before %s", upgradeTime.UTC().String()))
		} else {
			if err := k.distKeeper.DistributeFromFeePool(ctx, receiverInfo.Amount, receiverInfo.GetAddress()); err != nil {
				return err
			}
		}
	}

	return nil
}
