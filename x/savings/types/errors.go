package types

import errorsmod "cosmossdk.io/errors"

// DONTCOVER

var (
	// ErrEmptyInput error for empty input
	ErrEmptyInput = errorsmod.Register(ModuleName, 2, "input must not be empty")
	// ErrNoDepositFound error when no deposit is found for an address
	ErrNoDepositFound = errorsmod.Register(ModuleName, 3, "no deposit found")
	// ErrInvalidDepositDenom error for invalid deposit denom
	ErrInvalidDepositDenom = errorsmod.Register(ModuleName, 4, "invalid deposit denom")
	// ErrInvalidWithdrawDenom error for invalid withdraw denoms
	ErrInvalidWithdrawDenom = errorsmod.Register(ModuleName, 5, "invalid withdraw denom")
)
