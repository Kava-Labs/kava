package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// earn module errors
var (
	ErrInvalidVaultDenom        = sdkerrors.Register(ModuleName, 2, "invalid vault denom")
	ErrInvalidVaultStrategy     = sdkerrors.Register(ModuleName, 3, "invalid vault strategy")
	ErrInsufficientAmount       = sdkerrors.Register(ModuleName, 4, "insufficient amount")
	ErrInsufficientValue        = sdkerrors.Register(ModuleName, 5, "insufficient vault account value")
	ErrVaultRecordNotFound      = sdkerrors.Register(ModuleName, 6, "vault record not found")
	ErrVaultShareRecordNotFound = sdkerrors.Register(ModuleName, 7, "vault share record not found")
	ErrInvalidVaultTotalSupply  = sdkerrors.Register(ModuleName, 8, "invalid vault total supply")
)
