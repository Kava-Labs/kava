package hard_borrow

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/incentive/types"
)

var _ types.SourceAdapter = SourceAdapter{}

type SourceAdapter struct {
	keeper types.HardKeeper
}

func NewSourceAdapter(keeper types.HardKeeper) SourceAdapter {
	return SourceAdapter{
		keeper: keeper,
	}
}

func (f SourceAdapter) TotalSharesBySource(ctx sdk.Context, sourceID string) sdk.Dec {
	coins, found := f.keeper.GetBorrowedCoins(ctx)
	if !found {
		return sdk.ZeroDec()
	}

	return coins.AmountOf(sourceID).ToDec()
}

func (f SourceAdapter) OwnerSharesBySource(
	ctx sdk.Context,
	owner sdk.AccAddress,
	sourceIDs []string,
) map[string]sdk.Dec {
	borrowCoins := sdk.NewCoins()

	accBorrow, found := f.keeper.GetBorrow(ctx, owner)
	if found {
		borrowCoins = accBorrow.Amount
	}

	shares := make(map[string]sdk.Dec)
	for _, id := range sourceIDs {
		shares[id] = borrowCoins.AmountOf(id).ToDec()
	}

	return shares
}
