package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// DONTCOVER

var (
	// ErrEmptyInput error for empty input
	ErrEmptyInput = sdkerrors.Register(ModuleName, 2, "input must not be empty")
)
