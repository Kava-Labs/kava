package types

import (
	"encoding/hex"
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
	CodeAssetSupplyNotSet       CodeType          = 7
	CodeAboveAssetSupplyLimit   CodeType          = 8
	CodeInvalidClaimSecret      CodeType          = 9
	CodeAtomicSwapAlreadyExists CodeType          = 10
	CodeAtomicSwapNotFound      CodeType          = 11
	CodeSwapNotRefundable       CodeType          = 12
	CodeSwapNotClaimable        CodeType          = 13
)

// ErrConversionToBytesFailed error for when a hex encoded string can't be converted to bytes
func ErrConversionToBytesFailed(codespace sdk.CodespaceType, hexEncodedValue string) sdk.Error {
	return sdk.NewError(codespace, CodeConversionToBytesFailed, fmt.Sprintf("couldn't convert hex encoded %s to bytes", hexEncodedValue))
}

// ErrInvalidTimestamp error for when an timestamp is outside of bounds. Assumes block time of 10 seconds.
func ErrInvalidTimestamp(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidTimestamp, fmt.Sprintf("Timestamp can neither be 15 minutes ahead of the current time, nor 30 minutes later"))
}

// ErrInvalidHeightSpan error a proposed height span is outside of lock time range
func ErrInvalidHeightSpan(codespace sdk.CodespaceType, heightspan int64, minLockTime int64, maxLockTime int64) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidHeightSpan, fmt.Sprintf("height span %d is outside acceptable range %d - %d", heightspan, minLockTime, maxLockTime))
}

// ErrAmountTooSmall error for when a coin amount is 0
func ErrAmountTooSmall(codespace sdk.CodespaceType, coin sdk.Coin) sdk.Error {
	return sdk.NewError(codespace, CodeAmountTooSmall, fmt.Sprintf("coin %s amount is below the limit for this operation", coin))
}

// ErrAssetNotSupported error for when an asset is not supported
func ErrAssetNotSupported(codespace sdk.CodespaceType, denom string) sdk.Error {
	return sdk.NewError(codespace, CodeAssetNotSupported, fmt.Sprintf("asset %s is not on the list of supported assets", denom))
}

// ErrAssetNotActive error for when an asset is currently inactive
func ErrAssetNotActive(codespace sdk.CodespaceType, denom string) sdk.Error {
	return sdk.NewError(codespace, CodeAssetNotActive, fmt.Sprintf("asset %s is current inactive", denom))
}

// ErrAssetSupplyNotSet error for when an asset's supply has not been initialized
func ErrAssetSupplyNotSet(codespace sdk.CodespaceType, denom string) sdk.Error {
	return sdk.NewError(codespace, CodeAssetSupplyNotSet, fmt.Sprintf("%s asset supply hasn't been set", denom))
}

// ErrAboveAssetSupplyLimit error for when a proposed asset supply increase would put the supply over the limit
func ErrAboveAssetSupplyLimit(codespace sdk.CodespaceType, denom string, currentSupply, supplyLimit sdk.Int) sdk.Error {
	return sdk.NewError(codespace, CodeAboveAssetSupplyLimit,
		fmt.Sprintf("invalid supply increase - %s has a supply of %d and a supply limit of %d",
			denom, currentSupply.Int64(), supplyLimit.Int64()))
}

// ErrInvalidClaimSecret error when a submitted secret doesn't match an AtomicSwap's swapID
func ErrInvalidClaimSecret(codespace sdk.CodespaceType, submittedSecret []byte, swapID []byte) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidClaimSecret,
		fmt.Sprintf("hashed claim attempt %s does not match %s",
			hex.EncodeToString(submittedSecret),
			hex.EncodeToString(swapID),
		),
	)
}

// ErrAtomicSwapAlreadyExists error for when an AtomicSwap with this swapID already exists
func ErrAtomicSwapAlreadyExists(codespace sdk.CodespaceType, swapID cmn.HexBytes) sdk.Error {
	return sdk.NewError(codespace, CodeAtomicSwapAlreadyExists, fmt.Sprintf("atomic swap %s already exists", swapID))
}

// ErrAtomicSwapNotFound error for when an atomic swap is not found
func ErrAtomicSwapNotFound(codespace sdk.CodespaceType, id []byte) sdk.Error {
	return sdk.NewError(codespace, CodeAtomicSwapNotFound, fmt.Sprintf("AtomicSwap %s was not found", hex.EncodeToString(id)))
}

// ErrSwapNotRefundable error for when an AtomicSwap has not expired and cannot be refunded
func ErrSwapNotRefundable(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeSwapNotRefundable, fmt.Sprintf("atomic swap is still active and cannot be refunded"))
}

// ErrSwapNotClaimable error for when an atomic swap is not open and cannot be claimed
func ErrSwapNotClaimable(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeSwapNotClaimable, fmt.Sprintf("atomic swap is not claimable"))
}
