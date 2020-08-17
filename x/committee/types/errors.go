package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrUnknownCommittee        = sdkerrors.Register(ModuleName, 2, "committee not found")
	ErrInvalidCommittee        = sdkerrors.Register(ModuleName, 3, "invalid committee")
	ErrUnknownProposal         = sdkerrors.Register(ModuleName, 4, "proposal not found")
	ErrProposalExpired         = sdkerrors.Register(ModuleName, 5, "proposal expired")
	ErrInvalidPubProposal      = sdkerrors.Register(ModuleName, 6, "invalid pubproposal")
	ErrUnknownVote             = sdkerrors.Register(ModuleName, 7, "vote not found")
	ErrInvalidGenesis          = sdkerrors.Register(ModuleName, 8, "invalid genesis")
	ErrNoProposalHandlerExists = sdkerrors.Register(ModuleName, 9, "pubproposal has no corresponding handler")
	ErrUnknownSubspace         = sdkerrors.Register(ModuleName, 10, "subspace not found")
)
