package adapters

import (
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/incentive/keeper/adapters/swap"
	"github.com/kava-labs/kava/x/incentive/types"
)

type SourceAdapters struct {
	adapters map[types.ClaimType]types.SourceAdapter
}

type SourceShare struct {
	ID     string
	Shares sdk.Dec
}

func NewSourceAdapters(
	swapKeeper types.SwapKeeper,
) SourceAdapters {
	return SourceAdapters{
		adapters: map[types.ClaimType]types.SourceAdapter{
			types.CLAIM_TYPE_SWAP: swap.NewSourceAdapter(swapKeeper),
		},
	}
}

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

	var sortedSourceIDs []string
	for sourceID := range ownerShares {
		sortedSourceIDs = append(sortedSourceIDs, sourceID)
	}

	// Sort source IDs to ensure deterministic order of claim syncs
	sort.Strings(sortedSourceIDs)

	var shares []SourceShare
	for _, sourceID := range sortedSourceIDs {
		shares = append(shares, SourceShare{
			ID:     sourceID,
			Shares: ownerShares[sourceID],
		})
	}

	return shares
}
