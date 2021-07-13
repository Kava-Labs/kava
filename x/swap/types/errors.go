package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// swap module errors
var (
	ErrNotAllowed            = sdkerrors.Register(ModuleName, 2, "not allowed")
	ErrInvalidDeadline       = sdkerrors.Register(ModuleName, 3, "invalid deadline")
	ErrDeadlineExceeded      = sdkerrors.Register(ModuleName, 4, "deadline exceeded")
	ErrSlippageExceeded      = sdkerrors.Register(ModuleName, 5, "slippage exceeded")
	ErrInvalidPool           = sdkerrors.Register(ModuleName, 6, "invalid pool")
	ErrInvalidSlippage       = sdkerrors.Register(ModuleName, 7, "invalid slippage")
	ErrInsufficientLiquidity = sdkerrors.Register(ModuleName, 8, "insufficient liquidity")
	ErrInvalidShares         = sdkerrors.Register(ModuleName, 9, "invalid shares")
	ErrDepositNotFound       = sdkerrors.Register(ModuleName, 10, "deposit not found")
	ErrInvalidCoin           = sdkerrors.Register(ModuleName, 11, "invalid coin")
	ErrNotImplemented        = sdkerrors.Register(ModuleName, 12, "not implemented")
)
