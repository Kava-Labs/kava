package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// DONTCOVER

var (
	// ErrEmptyInput error for empty input
	ErrEmptyInput = sdkerrors.Register(ModuleName, 2, "input must not be empty")
	// ErrNoDerivativeFound error for no derivative found in store
	ErrNoDerivativeFound = sdkerrors.Register(ModuleName, 3, "no derivative found")
	// ErrInvalidBurnAmount error for an invalid burn amount
	ErrInvalidBurnAmount = sdkerrors.Register(ModuleName, 4, "invalid burn amount")
)
