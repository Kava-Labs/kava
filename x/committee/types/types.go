package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
)

// A Committee is a collection of addresses that are allowed to vote and enact any governance proposal that passes their permissions.
type Committee struct {
	Members     []sdk.AccAddress
	Permissions []Permission
}

// Permission is anything with a method that validates whether a proposal is allowed by it or not.
type Permission interface {
	Allows(gov.Content) bool
}

// GOV STUFF --------------------------
// Should be much the same as in gov module, except Proposals are linked to a committee ID.

type Proposal struct {
	gov.Content
	ID          uint64
	committeeID uint64
	// TODO
}

type Vote struct {
	ProposalID uint64
	Voter      sdk.AccAddress
	Option     byte
}
