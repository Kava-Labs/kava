package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// DONTCOVER

var (
	// ErrFailedUndelegation error for delegations that fail to unbond
	ErrFailedUndelegation = sdkerrors.Register(ModuleName, 2, "undelegation failed")
)
