package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// errors
var (
	ErrABIPack                 = sdkerrors.Register(ModuleName, 2, "contract ABI pack failed")
	ErrEVMCall                 = sdkerrors.Register(ModuleName, 3, "EVM call unexpected error")
	ErrConversionNotEnabled    = sdkerrors.Register(ModuleName, 4, "ERC20 token not enabled to convert to sdk.Coin")
	ErrBalanceInvariance       = sdkerrors.Register(ModuleName, 5, "post EVM transfer balance invariant failed")
	ErrUnexpectedContractEvent = sdkerrors.Register(ModuleName, 6, "unexpected contract event")
)
