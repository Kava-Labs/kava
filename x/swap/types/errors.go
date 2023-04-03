package types

import errorsmod "cosmossdk.io/errors"

// swap module errors
var (
	ErrNotAllowed            = errorsmod.Register(ModuleName, 2, "not allowed")
	ErrInvalidDeadline       = errorsmod.Register(ModuleName, 3, "invalid deadline")
	ErrDeadlineExceeded      = errorsmod.Register(ModuleName, 4, "deadline exceeded")
	ErrSlippageExceeded      = errorsmod.Register(ModuleName, 5, "slippage exceeded")
	ErrInvalidPool           = errorsmod.Register(ModuleName, 6, "invalid pool")
	ErrInvalidSlippage       = errorsmod.Register(ModuleName, 7, "invalid slippage")
	ErrInsufficientLiquidity = errorsmod.Register(ModuleName, 8, "insufficient liquidity")
	ErrInvalidShares         = errorsmod.Register(ModuleName, 9, "invalid shares")
	ErrDepositNotFound       = errorsmod.Register(ModuleName, 10, "deposit not found")
	ErrInvalidCoin           = errorsmod.Register(ModuleName, 11, "invalid coin")
	ErrNotImplemented        = errorsmod.Register(ModuleName, 12, "not implemented")
)
