package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
)

// A Committee is a collection of addresses that are allowed to vote and enact any governance proposal that passes their permissions.
type Committee struct {
	ID          uint64 // TODO or a name?
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
	CommitteeID uint64
	// TODO
	// could store votes on the proposal object
}

type Vote struct {
	ProposalID uint64
	Voter      sdk.AccAddress
	Option     byte
}

// Genesis -------------------
// Ok just to dump everything to json and reload - if time involved then begin blocker will take care of closing expired proposals. And it won't enact proposals because they would've been immediately enacted before the halt if they passed.
// committee, proposals, votes
