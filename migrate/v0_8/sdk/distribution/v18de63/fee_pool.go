package v18de63

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// global fee pool for distribution
type FeePool struct {
	CommunityPool sdk.DecCoins `json:"community_pool" yaml:"community_pool"` // pool for community funds yet to be spent
}
