package types

import errorsmod "cosmossdk.io/errors"

// errors
var (
	ErrABIPack                 = errorsmod.Register(ModuleName, 2, "contract ABI pack failed")
	ErrEVMCall                 = errorsmod.Register(ModuleName, 3, "EVM call unexpected error")
	ErrConversionNotEnabled    = errorsmod.Register(ModuleName, 4, "ERC20 token not enabled to convert to sdk.Coin")
	ErrBalanceInvariance       = errorsmod.Register(ModuleName, 5, "post EVM transfer balance invariant failed")
	ErrUnexpectedContractEvent = errorsmod.Register(ModuleName, 6, "unexpected contract event")
	ErrInvalidCosmosDenom      = errorsmod.Register(ModuleName, 7, "invalid cosmos denom")
	ErrSDKConversionNotEnabled = errorsmod.Register(ModuleName, 8, "sdk.Coin not enabled to convert to ERC20 token")
)
