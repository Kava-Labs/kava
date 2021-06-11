package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrNotAllowed     = sdkerrors.Register(ModuleName, 1, "not allowed")
	ErrNotImplemented = sdkerrors.Register(ModuleName, 99, "not implemented")
)
