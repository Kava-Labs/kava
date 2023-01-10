package distribution

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CommunityKeeper defines the contract needed to be fulfilled for community module dependencies.
type CommunityKeeper interface {
	GetModuleAccountBalance(sdk.Context) sdk.Coins
}
