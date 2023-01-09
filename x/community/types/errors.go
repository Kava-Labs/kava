package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrInvalidProposal      = sdkerrors.Register(ModuleName, 1, "invalid community pool proposal")
	ErrProposalExecutionErr = sdkerrors.Register(ModuleName, 2, "community pool proposal message execution error")
)
