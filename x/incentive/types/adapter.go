package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// SourceAdapter queries source shares from an external module.
type SourceAdapter interface {
	// OwnerSharesBySource returns source shares owned by one address.
	//
	// For example, the shares a user owns in the kava:usdx and bnb:usdx swap pools.
	// It returns the shares for several sources at once, in a map of sourceIDs to shares. Specifying no sourceIDS will return no shares.
	// Note the returned map does not have a deterministic order.
	OwnerSharesBySource(ctx sdk.Context, owner sdk.AccAddress, sourceIDs []string) map[string]sdk.Dec

	// TotalSharesBySource returns the sum of all shares for a source (across all users).
	//
	// For example, the total number of shares in the kava:usdx swap pool for all users.
	TotalSharesBySource(ctx sdk.Context, sourceID string) sdk.Dec
}
