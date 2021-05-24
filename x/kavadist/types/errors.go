package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/kavadist errors
var (
	ErrInvalidProposalAmount  = sdkerrors.Register(ModuleName, 2, "invalid community pool multi-spend proposal amount")
	ErrEmptyProposalRecipient = sdkerrors.Register(ModuleName, 3, "invalid community pool multi-spend proposal recipient")
)
