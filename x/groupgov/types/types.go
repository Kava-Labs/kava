package types

import (
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// A Group is a collection of addresses that are allowed to vote and enact any governance proposal that passes their permissions.
type Group struct {
	Members     []sdk.AccAddress
	Permissions []Permission
}

// handler for MsgSubmitProposal needs to loop apply all group permission Allows methods to the proposal and do a bit OR to see if it should be accepted

// Permission is anything with a method that validates whether a proposal is allowed by it or not.
// Collectively, if one permission allows a proposal then the proposal is allowed through.
type Permission interface {
	Allows(gov.Proposal) bool // maybe don't reuse gov's type here
}

// A gov.Proposal to used to add/remove members from a group, or to add/remove permissions.
// Normally registered with standard gov. But could also be registed with groupgov to allow groups to be controlled by other groups.
type GroupChangeProposal struct {
	Members     []sdk.AccAddress
	Permissions []Permission
}

// STANDARD GOV STUFF --------------------------
// Should be much the same as in gov module. Either import gov types directly or do some copy n pasting.

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
