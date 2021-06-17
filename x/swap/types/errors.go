package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrNotAllowed       = sdkerrors.Register(ModuleName, 2, "not allowed")
	ErrInvalidDeadline  = sdkerrors.Register(ModuleName, 3, "invalid deadline")
	ErrDeadlineExceeded = sdkerrors.Register(ModuleName, 4, "deadline exceeded")
	ErrNotImplemented   = sdkerrors.Register(ModuleName, 5, "not implemented")
)
