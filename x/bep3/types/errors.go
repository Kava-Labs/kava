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
	CodeInvalidClaimSecret       CodeType          = 8
	CodeOnlySameChain            CodeType          = 9
	CodeOnlyOriginalCreator      CodeType          = 10
	CodeAssetNotSupported        CodeType          = 11
	CodeAssetNotActive           CodeType          = 12
	CodeInvalidHeightSpan        CodeType          = 13
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

// ErrInvalidClaimSecret error when a submitted secret doesn't match an HTLT's swapID
func ErrInvalidClaimSecret(codespace sdk.CodespaceType, submittedSecret string, swapID string) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidClaimSecret, fmt.Sprintf("hashed claim attempt %s does not match %s", submittedSecret, swapID))
}

// ErrOnlySameChain error for when an operation is not allowed for cross-chain swaps
func ErrOnlySameChain(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeOnlySameChain, fmt.Sprintf("this operation is only allowed for same-chain swaps"))
}

// ErrOnlyOriginalCreator error for when an operation restricted to the original htlt creator
func ErrOnlyOriginalCreator(codespace sdk.CodespaceType, sender sdk.AccAddress, creator sdk.AccAddress) sdk.Error {
	return sdk.NewError(codespace, CodeOnlyOriginalCreator, fmt.Sprintf("%s does not match original HTLT creator %s", sender, creator))
}

// ErrAssetNotSupported error for when an asset is not supported
func ErrAssetNotSupported(codespace sdk.CodespaceType, denom string) sdk.Error {
	return sdk.NewError(codespace, CodeAssetNotSupported, fmt.Sprintf("asset %s is not on the list of supported assets %s", denom))
}

// ErrAssetNotActive error for when an asset is currently inactive
func ErrAssetNotActive(codespace sdk.CodespaceType, denom string) sdk.Error {
	return sdk.NewError(codespace, CodeAssetNotActive, fmt.Sprintf("asset %s is current inactive %s", denom))
}

// ErrInvalidHeightSpan error a proposed height span is outside of lock time range
func ErrInvalidHeightSpan(codespace sdk.CodespaceType, heightspan int64, minLockTime int64, maxLockTime int64) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidHeightSpan, fmt.Sprintf("height span %d is outside acceptable range %d - %d", heightspan, minLockTime, maxLockTime))
}
