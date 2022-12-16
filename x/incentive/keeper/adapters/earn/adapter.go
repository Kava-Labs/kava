package earn

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	earntypes "github.com/kava-labs/kava/x/earn/types"
	"github.com/kava-labs/kava/x/incentive/types"
)

var _ types.SourceAdapter = SourceAdapter{}

type SourceAdapter struct {
	keeper types.EarnKeeper
}

func NewSourceAdapter(keeper types.EarnKeeper) SourceAdapter {
	return SourceAdapter{
		keeper: keeper,
	}
}

func (f SourceAdapter) TotalSharesBySource(ctx sdk.Context, sourceID string) sdk.Dec {
	vaultShares, found := f.keeper.GetVaultTotalShares(ctx, sourceID)
	if !found {
		return sdk.ZeroDec()
	}

	return vaultShares.Amount
}

func (f SourceAdapter) OwnerSharesBySource(
	ctx sdk.Context,
	owner sdk.AccAddress,
	sourceIDs []string,
) map[string]sdk.Dec {
	accountShares, found := f.keeper.GetVaultAccountShares(ctx, owner)
	if !found {
		accountShares = earntypes.VaultShares{}
	}

	shares := make(map[string]sdk.Dec)
	for _, id := range sourceIDs {
		// Sets shares to zero if not found
		shares[id] = accountShares.AmountOf(id)
	}

	return shares
}
