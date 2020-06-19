package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// DONTCOVER

// Errors used by the issuance module
var (
	ErrAssetNotFound = sdkerrors.Register(ModuleName, 1, "no asset with input denom found")
	ErrNotAuthorized = sdkerrors.Register(ModuleName, 2, "account not authorized ")
)
