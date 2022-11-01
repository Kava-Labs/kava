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

func (f SourceAdapter) GetTotalShares(ctx sdk.Context, sourceID string) sdk.Dec {
	vaultShare, found := f.keeper.GetVaultTotalShares(ctx, sourceID)
	if !found {
		return sdk.ZeroDec()
	}
	return vaultShare.Amount
}

func (f SourceAdapter) GetShares(ctx sdk.Context, owner sdk.AccAddress, sourceIDs []string) []sdk.Dec {
	vaultShares, found := f.keeper.GetVaultAccountShares(ctx, owner)
	if !found {
		vaultShares = earntypes.NewVaultShares()
	}

	var shares []sdk.Dec
	for _, id := range sourceIDs {
		shares = append(shares, vaultShares.AmountOf(id))
	}
	return shares
}
