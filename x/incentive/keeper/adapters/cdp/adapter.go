package cdp

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	cdptypes "github.com/kava-labs/kava/x/cdp/types"

	"github.com/kava-labs/kava/x/incentive/types"
)

var _ types.SourceAdapter = SourceAdapter{}

type SourceAdapter struct {
	keeper types.CdpKeeper
}

func NewSourceAdapter(keeper types.CdpKeeper) SourceAdapter {
	return SourceAdapter{
		keeper: keeper,
	}
}

func (f SourceAdapter) TotalSharesBySource(ctx sdk.Context, sourceID string) sdk.Dec {
	totalPrincipal := f.keeper.GetTotalPrincipal(ctx, sourceID, cdptypes.DefaultStableDenom)
	return totalPrincipal.ToDec()
}

func (f SourceAdapter) OwnerSharesBySource(
	ctx sdk.Context,
	owner sdk.AccAddress,
	sourceIDs []string,
) map[string]sdk.Dec {
	shares := make(map[string]sdk.Dec)
	for _, id := range sourceIDs {
		cdp, found := f.keeper.GetCdpByOwnerAndCollateralType(ctx, owner, id)
		if !found {
			shares[id] = sdk.ZeroDec()
			continue
		}

		normalizedPrincipal, err := cdp.GetNormalizedPrincipal()
		if err != nil {
			panic(fmt.Sprintf("could not get cdp normalized principal for %s: %s", cdp.Owner, err.Error()))
		}

		// Sets shares to zero if not found
		shares[id] = normalizedPrincipal
	}

	return shares
}
