package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// DONTCOVER

var (
	ErrCustom = sdkerrors.Register(ModuleName, 2, "")
)
