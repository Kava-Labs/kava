package types

import errorsmod "cosmossdk.io/errors"

// earn module errors
var (
	ErrInvalidVaultDenom        = errorsmod.Register(ModuleName, 2, "invalid vault denom")
	ErrInvalidVaultStrategy     = errorsmod.Register(ModuleName, 3, "vault does not support this strategy")
	ErrInsufficientAmount       = errorsmod.Register(ModuleName, 4, "insufficient amount")
	ErrInsufficientValue        = errorsmod.Register(ModuleName, 5, "insufficient vault account value")
	ErrVaultRecordNotFound      = errorsmod.Register(ModuleName, 6, "vault record not found")
	ErrVaultShareRecordNotFound = errorsmod.Register(ModuleName, 7, "vault share record not found")
	ErrAccountDepositNotAllowed = errorsmod.Register(ModuleName, 8, "account is not allowed to deposit to this vault")
)
