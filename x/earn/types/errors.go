package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// earn module errors
var (
	ErrInvalidVaultDenom         = sdkerrors.Register(ModuleName, 2, "invalid vault denom")
	ErrInvalidVaultStrategy      = sdkerrors.Register(ModuleName, 3, "invalid vault strategy")
	ErrInsufficientAmount        = sdkerrors.Register(ModuleName, 4, "insufficient amount")
	ErrInvalidShares             = sdkerrors.Register(ModuleName, 5, "invalid shares")
	ErrVaultRecordNotFound       = sdkerrors.Register(ModuleName, 6, "vault record not found")
	ErrVaultShareRecordNotFound  = sdkerrors.Register(ModuleName, 7, "vault share record not found")
	ErrStrategyDenomNotSupported = sdkerrors.Register(ModuleName, 8, "denom not supported for strategy")
)
