package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrInvalidProposal          = sdkerrors.Register(ModuleName, 1, "invalid community pool proposal")
	ErrProposalExecutionErr     = sdkerrors.Register(ModuleName, 2, "community proposal message execution error")
	ErrProposalMsgNotEnabledErr = sdkerrors.Register(ModuleName, 3, "community proposal message not whitelisted")
	ErrProposalSigningErr       = sdkerrors.Register(ModuleName, 4, "community proposal message signer validation error")
)
