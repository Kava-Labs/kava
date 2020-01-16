package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// DefaultCodespace codespace for the module
	DefaultCodespace sdk.CodespaceType = ModuleName

	// CodeEmptyInput error code for empty input errors
	CodeEmptyInput sdk.CodeType = 1
	// CodeExpired error code for expired prices
	CodeExpired sdk.CodeType = 2
	// CodeInvalidPrice error code for all input prices expired
	CodeInvalidPrice sdk.CodeType = 3
	// CodeInvalidAsset error code for invalid asset
	CodeInvalidAsset sdk.CodeType = 4
	// CodeInvalidOracle error code for invalid oracle
	CodeInvalidOracle sdk.CodeType = 5
)

// ErrEmptyInput Error constructor
func ErrEmptyInput(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeEmptyInput, fmt.Sprintf("Input must not be empty."))
}

// ErrExpired Error constructor for posted price messages with expired price
func ErrExpired(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeExpired, fmt.Sprintf("Price is expired."))
}

// ErrNoValidPrice Error constructor for posted price messages with expired price
func ErrNoValidPrice(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidPrice, fmt.Sprintf("All input prices are expired."))
}

// ErrInvalidAsset Error constructor for posted price messages for invalid markets
func ErrInvalidMarket(codespace sdk.CodespaceType, marketId string) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidAsset, fmt.Sprintf("market %s does not exist", marketId))
}

// ErrInvalidOracle Error constructor for posted price messages for invalid oracles
func ErrInvalidOracle(codespace sdk.CodespaceType, addr sdk.AccAddress) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidOracle, fmt.Sprintf("oracle %s does not exist or not authorized", addr))
}
