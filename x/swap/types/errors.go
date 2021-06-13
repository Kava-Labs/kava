package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrNotAllowed     = sdkerrors.Register(ModuleName, 2, "not allowed")
	ErrNotImplemented = sdkerrors.Register(ModuleName, 3, "not implemented")
)
