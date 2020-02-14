package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CodeType is the local code type
type CodeType = sdk.CodeType

const (
	// DefaultCodespace default bep3 codespace
	DefaultCodespace             sdk.CodespaceType = ModuleName
	CodeInvalidLockTime          CodeType          = 1
	CodeInvalidModulePermissions CodeType          = 2
	CodeHTLTNotFound             CodeType          = 3
	CodeInvalidCoinDenom         CodeType          = 4
	CodeAmountTooLarge           CodeType          = 5
	CodeAmountTooSmall           CodeType          = 6
	CodeHTLTAlreadyExists        CodeType          = 7
)

// ErrInvalidLockTime Error constructor
func ErrInvalidLockTime(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidLockTime, fmt.Sprintf("invalid lock time: must be greater than minimum lock time"))
}

// ErrInvalidModulePermissions error for when module doesn't have valid permissions
func ErrInvalidModulePermissions(codespace sdk.CodespaceType, permission string) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidModulePermissions, fmt.Sprintf("module does not have required permission '%s'", permission))
}

// ErrHTLTNotFound error for when an htlt is not found
func ErrHTLTNotFound(codespace sdk.CodespaceType, id string) sdk.Error {
	return sdk.NewError(codespace, CodeHTLTNotFound, fmt.Sprintf("HTLT %s was not found", id))
}

// ErrInvalidCoinDenom error for when coin denom doesn't match HTLT coin denom
func ErrInvalidCoinDenom(codespace sdk.CodespaceType, coinDenom string, htltCoinDenom string) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidCoinDenom, fmt.Sprintf("coin denom %s doesn't match HTLT coin denom %s", coinDenom, htltCoinDenom))
}

// ErrAmountTooLarge error for when a coin amount will put the asset supply over the asset limit
func ErrAmountTooLarge(codespace sdk.CodespaceType, coin sdk.Coin) sdk.Error {
	return sdk.NewError(codespace, CodeAmountTooLarge, fmt.Sprintf("deposit of asset %s not allowed due to the asset's global supply limit", coin.String()))
}

// ErrAmountTooSmall error for when a coin amount is 0
func ErrAmountTooSmall(codespace sdk.CodespaceType, coin sdk.Coin) sdk.Error {
	return sdk.NewError(codespace, CodeAmountTooSmall, fmt.Sprintf("coin %s amount is below the limit for this operation", coin.String()))
}

// ErrHTLTAlreadyExists error for when an HTLT with this swapID already exists
func ErrHTLTAlreadyExists(codespace sdk.CodespaceType, swapID string) sdk.Error {
	return sdk.NewError(codespace, CodeHTLTAlreadyExists, fmt.Sprintf("coin %s amount is below the limit for this operation", swapID))
}
