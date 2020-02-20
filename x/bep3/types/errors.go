package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	cmn "github.com/tendermint/tendermint/libs/common"
)

// CodeType is the local code type
type CodeType = sdk.CodeType

const (
	// DefaultCodespace default bep3 codespace
	DefaultCodespace             sdk.CodespaceType = ModuleName
	CodeInvalidLockTime          CodeType          = 1
	CodeInvalidModulePermissions CodeType          = 2
	CodeAtomicSwapNotFound       CodeType          = 3
	CodeInvalidCoinDenom         CodeType          = 4
	CodeAmountTooLarge           CodeType          = 5
	CodeAmountTooSmall           CodeType          = 6
	CodeAtomicSwapAlreadyExists  CodeType          = 7
	CodeInvalidClaimSecret       CodeType          = 8
	CodeOnlySameChain            CodeType          = 9
	CodeOnlyOriginalCreator      CodeType          = 10
	CodeAssetNotSupported        CodeType          = 11
	CodeAssetNotActive           CodeType          = 12
	CodeInvalidHeightSpan        CodeType          = 13
	CodeAtomicSwapHasExpired     CodeType          = 14
	CodeOnlyDeputy               CodeType          = 15
)

// ErrInvalidLockTime Error constructor
func ErrInvalidLockTime(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidLockTime, fmt.Sprintf("invalid lock time: must be greater than minimum lock time"))
}

// ErrInvalidModulePermissions error for when module doesn't have valid permissions
func ErrInvalidModulePermissions(codespace sdk.CodespaceType, permission string) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidModulePermissions, fmt.Sprintf("module does not have required permission '%s'", permission))
}

// ErrAtomicSwapNotFound error for when an atomic swap is not found
func ErrAtomicSwapNotFound(codespace sdk.CodespaceType, id []byte) sdk.Error {
	return sdk.NewError(codespace, CodeAtomicSwapNotFound, fmt.Sprintf("AtomicSwap %s was not found", BytesToHexEncodedString(id)))
}

// ErrInvalidCoinDenom error for when coin denom doesn't match AtomicSwap coin denom
func ErrInvalidCoinDenom(codespace sdk.CodespaceType, coinDenom string, swapCoinDenom string) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidCoinDenom, fmt.Sprintf("coin denom %s doesn't match AtomicSwap coin denom %s", coinDenom, swapCoinDenom))
}

// ErrAmountTooLarge error for when a coin amount will put the asset supply over the asset limit
func ErrAmountTooLarge(codespace sdk.CodespaceType, coin sdk.Coin) sdk.Error {
	return sdk.NewError(codespace, CodeAmountTooLarge, fmt.Sprintf("deposit of asset %s not allowed due to the asset's global supply limit", coin))
}

// ErrAmountTooSmall error for when a coin amount is 0
func ErrAmountTooSmall(codespace sdk.CodespaceType, coin sdk.Coin) sdk.Error {
	return sdk.NewError(codespace, CodeAmountTooSmall, fmt.Sprintf("coin %s amount is below the limit for this operation", coin))
}

// ErrAtomicSwapAlreadyExists error for when an AtomicSwap with this swapID already exists
func ErrAtomicSwapAlreadyExists(codespace sdk.CodespaceType, swapID cmn.HexBytes) sdk.Error {
	return sdk.NewError(codespace, CodeAtomicSwapAlreadyExists, fmt.Sprintf("coin %s amount is below the limit for this operation", swapID))
}

// ErrInvalidClaimSecret error when a submitted secret doesn't match an AtomicSwap's swapID
func ErrInvalidClaimSecret(codespace sdk.CodespaceType, submittedSecret []byte, swapID []byte) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidClaimSecret,
		fmt.Sprintf("hashed claim attempt %s does not match %s",
			BytesToHexEncodedString(submittedSecret),
			BytesToHexEncodedString(swapID),
		),
	)
}

// ErrOnlySameChain error for when an operation is not allowed for cross-chain swaps
func ErrOnlySameChain(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeOnlySameChain, fmt.Sprintf("this operation is only allowed for same-chain swaps"))
}

// ErrOnlyOriginalCreator error for when an operation restricted to the original atomic swap creator
func ErrOnlyOriginalCreator(codespace sdk.CodespaceType, sender sdk.AccAddress, creator sdk.AccAddress) sdk.Error {
	return sdk.NewError(codespace, CodeOnlyOriginalCreator, fmt.Sprintf("%s does not match original AtomicSwap creator %s", sender, creator))
}

// ErrAssetNotSupported error for when an asset is not supported
func ErrAssetNotSupported(codespace sdk.CodespaceType, denom string) sdk.Error {
	return sdk.NewError(codespace, CodeAssetNotSupported, fmt.Sprintf("asset %s is not on the list of supported assets", denom))
}

// ErrAssetNotActive error for when an asset is currently inactive
func ErrAssetNotActive(codespace sdk.CodespaceType, denom string) sdk.Error {
	return sdk.NewError(codespace, CodeAssetNotActive, fmt.Sprintf("asset %s is current inactive", denom))
}

// ErrInvalidHeightSpan error a proposed height span is outside of lock time range
func ErrInvalidHeightSpan(codespace sdk.CodespaceType, heightspan int64, minLockTime int64, maxLockTime int64) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidHeightSpan, fmt.Sprintf("height span %d is outside acceptable range %d - %d", heightspan, minLockTime, maxLockTime))
}

// ErrAtomicSwapHasExpired error for when a AtomicSwap has expired and cannot be claimed
func ErrAtomicSwapHasExpired(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeAtomicSwapHasExpired, fmt.Sprintf("atomic swap is expired"))
}

// ErrOnlyDeputy error for when an operation restricted to the authorized deputy
func ErrOnlyDeputy(codespace sdk.CodespaceType, sender sdk.AccAddress, deputy sdk.AccAddress) sdk.Error {
	return sdk.NewError(codespace, CodeOnlyDeputy, fmt.Sprintf("%s does not match authorized deputy %s", sender, deputy))
}
