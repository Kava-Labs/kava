package savings

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/incentive/types"
	savingstypes "github.com/kava-labs/kava/x/savings/types"
)

var _ types.SourceAdapter = SourceAdapter{}

type SourceAdapter struct {
	keeper types.SavingsKeeper
}

func NewSourceAdapter(keeper types.SavingsKeeper) SourceAdapter {
	return SourceAdapter{
		keeper: keeper,
	}
}

func (f SourceAdapter) TotalSharesBySource(ctx sdk.Context, sourceID string) sdk.Dec {
	bal := f.keeper.GetSavingsModuleAccountBalances(ctx)
	return bal.AmountOf(sourceID).ToDec()
}

func (f SourceAdapter) OwnerSharesBySource(
	ctx sdk.Context,
	owner sdk.AccAddress,
	sourceIDs []string,
) map[string]sdk.Dec {
	deposit, found := f.keeper.GetDeposit(ctx, owner)
	if !found {
		deposit = savingstypes.Deposit{}
	}

	shares := make(map[string]sdk.Dec)
	for _, id := range sourceIDs {
		shares[id] = deposit.Amount.AmountOf(id).ToDec()
	}

	return shares
}
