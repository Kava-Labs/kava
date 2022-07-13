package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// swap module errors
var (
	ErrInvalidVaultDenom    = sdkerrors.Register(ModuleName, 2, "invalid vault denom")
	ErrInvalidVaultStrategy = sdkerrors.Register(ModuleName, 3, "invalid vault strategy")
)
