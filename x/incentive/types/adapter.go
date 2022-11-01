package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// SourceAdapter provides source shares from an external module.
type SourceAdapter interface {
	// GetShares returns source shares owned by one address.
	//
	// For example, the shares a user owns in the kava:usdx and bnb:usdx swap pools.
	// It returns the shares for several sources at once, in the same order as the sourceIDs. Specifying no sourceIDS will return no shares.
	GetShares(ctx sdk.Context, owner sdk.AccAddress, sourceIDs []string) []sdk.Dec

	// GetTotalShares returns the sum of all shares for a source (across all users).
	//
	// For example, the total number of shares in the kava:usdx swap pool for all users.
	GetTotalShares(ctx sdk.Context, sourceID string) sdk.Dec
}
