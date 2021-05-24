package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/kavadist errors
var (
	ErrInvalidProposalAmount  = sdkerrors.Register(ModuleName, 1, "invalid community pool multi-spend proposal amount")
	ErrEmptyProposalRecipient = sdkerrors.Register(ModuleName, 10, "invalid community pool multi-spend proposal recipient")
)
