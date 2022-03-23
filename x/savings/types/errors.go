package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// DONTCOVER

var (
	// ErrEmptyInput error for empty input
	ErrEmptyInput = sdkerrors.Register(ModuleName, 2, "input must not be empty")
	// ErrInvalidDepositDenom error for invalid deposit denoms
	ErrInvalidDepositDenom = sdkerrors.Register(ModuleName, 3, "invalid deposit denom")
)
