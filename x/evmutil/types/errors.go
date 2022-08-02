package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// errors
var (
	ErrABIPack              = sdkerrors.Register(ModuleName, 2, "contract ABI pack failed")
	ErrConversionNotEnabled = sdkerrors.Register(ModuleName, 4, "ERC20 token not enabled to convert to sdk.Coin")
)
