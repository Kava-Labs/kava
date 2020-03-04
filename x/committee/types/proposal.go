package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// A gov.Proposal to used to add/remove members from a group, or to add/remove permissions.
// Normally registered with standard gov. But could also be registed with committee to allow groups to be controlled by other groups.
type GroupChangeProposal struct {
	Members     []sdk.AccAddress
	Permissions []Permission
}
