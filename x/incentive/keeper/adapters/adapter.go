package adapters

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/incentive/keeper/adapters/earn"
	"github.com/kava-labs/kava/x/incentive/keeper/adapters/swap"
	"github.com/kava-labs/kava/x/incentive/types"
)

// SourceAdapters is a collection of source adapters.
type SourceAdapters struct {
	adapters map[types.ClaimType]types.SourceAdapter
}

// SourceShare is a single share from a source with it's corresponding ID.
type SourceShare struct {
	ID     string
	Shares sdk.Dec
}

// NewSourceAdapters returns a new SourceAdapters instance with all available
// source adapters.
func NewSourceAdapters(
	swapKeeper types.SwapKeeper,
	earnKeeper types.EarnKeeper,
) SourceAdapters {
	return SourceAdapters{
		adapters: map[types.ClaimType]types.SourceAdapter{
			types.CLAIM_TYPE_SWAP: swap.NewSourceAdapter(swapKeeper),
			types.CLAIM_TYPE_EARN: earn.NewSourceAdapter(earnKeeper),
		},
	}
}

// OwnerSharesBySource returns a slice of SourceShares for each sourceID from a
// specified owner. The slice is sorted by sourceID.
func (a SourceAdapters) OwnerSharesBySource(
	ctx sdk.Context,
	claimType types.ClaimType,
	owner sdk.AccAddress,
	sourceIDs []string,
) []SourceShare {
	adapter, found := a.adapters[claimType]
	if !found {
		panic(fmt.Sprintf("no source share fetcher for claim type %s", claimType))
	}

	ownerShares := adapter.OwnerSharesBySource(ctx, owner, sourceIDs)

	var shares []SourceShare
	for _, sourceID := range sourceIDs {
		singleShares, found := ownerShares[sourceID]
		if !found {
			panic(fmt.Sprintf("no source shares for claimType %s and source %s", claimType, sourceID))
		}

		shares = append(shares, SourceShare{
			ID:     sourceID,
			Shares: singleShares,
		})
	}

	return shares
}

// TotalSharesBySource returns the total shares of a given claimType and sourceID.
func (a SourceAdapters) TotalSharesBySource(
	ctx sdk.Context,
	claimType types.ClaimType,
	sourceID string,
) sdk.Dec {
	adapter, found := a.adapters[claimType]
	if !found {
		panic(fmt.Sprintf("no source share fetcher for claim type %s", claimType))
	}

	return adapter.TotalSharesBySource(ctx, sourceID)
}
