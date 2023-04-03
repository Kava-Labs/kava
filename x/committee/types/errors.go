package types

import errorsmod "cosmossdk.io/errors"

var (
	ErrUnknownCommittee        = errorsmod.Register(ModuleName, 2, "committee not found")
	ErrInvalidCommittee        = errorsmod.Register(ModuleName, 3, "invalid committee")
	ErrUnknownProposal         = errorsmod.Register(ModuleName, 4, "proposal not found")
	ErrProposalExpired         = errorsmod.Register(ModuleName, 5, "proposal expired")
	ErrInvalidPubProposal      = errorsmod.Register(ModuleName, 6, "invalid pubproposal")
	ErrUnknownVote             = errorsmod.Register(ModuleName, 7, "vote not found")
	ErrInvalidGenesis          = errorsmod.Register(ModuleName, 8, "invalid genesis")
	ErrNoProposalHandlerExists = errorsmod.Register(ModuleName, 9, "pubproposal has no corresponding handler")
	ErrUnknownSubspace         = errorsmod.Register(ModuleName, 10, "subspace not found")
	ErrInvalidVoteType         = errorsmod.Register(ModuleName, 11, "invalid vote type")
	ErrNotFoundProposalTally   = errorsmod.Register(ModuleName, 12, "proposal tally not found")
)
