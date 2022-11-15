package swap

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/incentive/types"
)

var _ types.SourceAdapter = SourceAdapter{}

type SourceAdapter struct {
	keeper types.SwapKeeper
}

func NewSourceAdapter(keeper types.SwapKeeper) SourceAdapter {
	return SourceAdapter{
		keeper: keeper,
	}
}

func (f SourceAdapter) TotalSharesBySource(ctx sdk.Context, sourceID string) sdk.Dec {
	shares, found := f.keeper.GetPoolShares(ctx, sourceID)
	if !found {
		shares = sdk.ZeroInt()
	}

	return shares.ToDec()
}

func (f SourceAdapter) OwnerSharesBySource(
	ctx sdk.Context,
	owner sdk.AccAddress,
	sourceIDs []string,
) map[string]sdk.Dec {
	shares := make(map[string]sdk.Dec)
	for _, id := range sourceIDs {
		s, found := f.keeper.GetDepositorSharesAmount(ctx, owner, id)
		if !found {
			s = sdk.ZeroInt()
		}

		shares[id] = s.ToDec()
	}

	return shares
}
