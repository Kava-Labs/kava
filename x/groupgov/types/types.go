package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
)

// A Group is a collection of addresses that are allowed to vote and enact any governance proposal that passes their permissions.
type Group struct {
	Members     []sdk.AccAddress
	Permissions []Permission
}

// Permission is anything with a method that validates whether a proposal is allowed by it or not.
// Collectively, if one permission allows a proposal then the proposal is allowed through.
type Permission interface {
	Allows(gov.Proposal) bool // maybe don't reuse gov's type here
}

// STANDARD GOV STUFF --------------------------
// Should be much the same as in gov module, except Proposals are linked to a group ID.

type Router struct {
	// TODO
}

type Proposal struct {
	ID      uint64
	groupID uint64
	// TODO
}

type Vote struct {
	proposalID uint64
	option     uint64
	// TODO
}
