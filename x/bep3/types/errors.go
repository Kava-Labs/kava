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
	DefaultCodespace            sdk.CodespaceType = ModuleName
	CodeConversionToBytesFailed CodeType          = 1
	CodeInvalidTimestamp        CodeType          = 2
	CodeInvalidHeightSpan       CodeType          = 3
	CodeAmountTooSmall          CodeType          = 4
	CodeAssetNotSupported       CodeType          = 5
	CodeAssetNotActive          CodeType          = 6
	CodeAboveAssetSupplyLimit   CodeType          = 7
	CodeInvalidClaimSecret      CodeType          = 8
	CodeAtomicSwapAlreadyExists CodeType          = 9
	CodeAtomicSwapNotFound      CodeType          = 10
	CodeSwapNotRefundable       CodeType          = 11
	CodeSwapNotOpen             CodeType          = 12
	CodeAtomicSwapHasExpired    CodeType          = 13
)

// ErrAtomicSwapNotFound error for when an atomic swap is not found
func ErrAtomicSwapNotFound(codespace sdk.CodespaceType, id []byte) sdk.Error {
	return sdk.NewError(codespace, CodeAtomicSwapNotFound, fmt.Sprintf("AtomicSwap %s was not found", BytesToHex(id)))
}

// ErrAmountTooSmall error for when a coin amount is 0
func ErrAmountTooSmall(codespace sdk.CodespaceType, coin sdk.Coin) sdk.Error {
	return sdk.NewError(codespace, CodeAmountTooSmall, fmt.Sprintf("coin %s amount is below the limit for this operation", coin))
}

// ErrAtomicSwapAlreadyExists error for when an AtomicSwap with this swapID already exists
func ErrAtomicSwapAlreadyExists(codespace sdk.CodespaceType, swapID cmn.HexBytes) sdk.Error {
	return sdk.NewError(codespace, CodeAtomicSwapAlreadyExists, fmt.Sprintf("atomic swap %s already exists", swapID))
}

// ErrInvalidClaimSecret error when a submitted secret doesn't match an AtomicSwap's swapID
func ErrInvalidClaimSecret(codespace sdk.CodespaceType, submittedSecret []byte, swapID []byte) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidClaimSecret,
		fmt.Sprintf("hashed claim attempt %s does not match %s",
			BytesToHex(submittedSecret),
			BytesToHex(swapID),
		),
	)
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

// ErrAtomicSwapHasExpired error for when an AtomicSwap has expired and cannot be claimed
func ErrAtomicSwapHasExpired(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeAtomicSwapHasExpired, fmt.Sprintf("atomic swap is expired"))
}

// ErrConversionToBytesFailed error for when a hex encoded string can't be converted to bytes
func ErrConversionToBytesFailed(codespace sdk.CodespaceType, hexEncodedValue string) sdk.Error {
	return sdk.NewError(codespace, CodeConversionToBytesFailed, fmt.Sprintf("couldn't convert hex encoded %s to bytes", hexEncodedValue))
}

// ErrAboveAssetSupplyLimit error for when a proposed asset supply increase would put the supply over the limit
func ErrAboveAssetSupplyLimit(codespace sdk.CodespaceType, denom string, currentSupply, proposedIncrease, supplyLimit int64) sdk.Error {
	return sdk.NewError(codespace, CodeAboveAssetSupplyLimit,
		fmt.Sprintf("%s has a current supply of %d. An increase of %d would put it above supply limit %d",
			denom, currentSupply, proposedIncrease, supplyLimit))
}

// ErrSwapNotOpen error for when an atomic swap is not open
func ErrSwapNotOpen(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeSwapNotOpen, fmt.Sprintf("swap is not open"))
}

// ErrSwapNotRefundable error for when an AtomicSwap has not expired and cannot be refunded
func ErrSwapNotRefundable(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeSwapNotRefundable, fmt.Sprintf("atomic swap is still active and cannot be refunded"))
}

// ErrInvalidTimestamp error for when an timestamp is outside of bounds. Assumes block time of 10 seconds.
func ErrInvalidTimestamp(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidTimestamp, fmt.Sprintf("Timestamp can neither be 15 minutes ahead of the current time, nor 30 minutes later"))
}
