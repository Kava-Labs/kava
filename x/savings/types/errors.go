package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// DONTCOVER

var (
	// ErrEmptyInput error for empty input
	ErrEmptyInput = sdkerrors.Register(ModuleName, 2, "input must not be empty")
	// ErrNoDepositFound error when no deposit is found for an address
	ErrNoDepositFound = sdkerrors.Register(ModuleName, 3, "no deposit found")
	// ErrInvalidDepositDenom error for invalid deposit denom
	ErrInvalidDepositDenom = sdkerrors.Register(ModuleName, 4, "invalid deposit denom")
	// ErrInvalidWithdrawDenom error for invalid withdraw denoms
	ErrInvalidWithdrawDenom = sdkerrors.Register(ModuleName, 5, "invalid withdraw denom")
)
