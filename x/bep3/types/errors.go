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
	CodeInvalidTimestamp        CodeType          = 1
	CodeInvalidHeightSpan       CodeType          = 2
	CodeAssetNotSupported       CodeType          = 3
	CodeAssetNotActive          CodeType          = 4
	CodeAssetSupplyNotFound     CodeType          = 5
	CodeExceedsSupplyLimit      CodeType          = 6
	CodeExceedsAvailableSupply  CodeType          = 7
	CodeInvalidCurrentSupply    CodeType          = 8
	CodeInvalidIncomingSupply   CodeType          = 9
	CodeInvalidOutgoingSupply   CodeType          = 10
	CodeInvalidClaimSecret      CodeType          = 11
	CodeAtomicSwapAlreadyExists CodeType          = 12
	CodeAtomicSwapNotFound      CodeType          = 13
	CodeSwapNotRefundable       CodeType          = 14
	CodeSwapNotClaimable        CodeType          = 15
)

// ErrInvalidTimestamp error for when an timestamp is outside of bounds. Assumes block time of 10 seconds.
func ErrInvalidTimestamp(codespace sdk.CodespaceType) error {
	return sdk.NewError(codespace, CodeInvalidTimestamp, fmt.Sprintf("Timestamp can neither be 15 minutes ahead of the current time, nor 30 minutes later"))
}

// ErrInvalidHeightSpan error a proposed height span is outside of lock time range
func ErrInvalidHeightSpan(codespace sdk.CodespaceType, heightspan int64, minLockTime int64, maxLockTime int64) error {
	return sdk.NewError(codespace, CodeInvalidHeightSpan, fmt.Sprintf("height span %d is outside acceptable range %d - %d", heightspan, minLockTime, maxLockTime))
}

// ErrAssetNotSupported error for when an asset is not supported
func ErrAssetNotSupported(codespace sdk.CodespaceType, denom string) error {
	return sdk.NewError(codespace, CodeAssetNotSupported, fmt.Sprintf("asset %s is not on the list of supported assets", denom))
}

// ErrAssetNotActive error for when an asset is currently inactive
func ErrAssetNotActive(codespace sdk.CodespaceType, denom string) error {
	return sdk.NewError(codespace, CodeAssetNotActive, fmt.Sprintf("asset %s is currently inactive", denom))
}

// ErrAssetSupplyNotFound error for when an asset's supply is not found in the store
func ErrAssetSupplyNotFound(codespace sdk.CodespaceType, denom string) error {
	return sdk.NewError(codespace, CodeAssetSupplyNotFound, fmt.Sprintf("%s asset supply not found in store", denom))
}

// ErrExceedsSupplyLimit error for when the proposed supply increase would put the supply above limit
func ErrExceedsSupplyLimit(codespace sdk.CodespaceType, increase, current, limit sdk.Coin) error {
	return sdk.NewError(codespace, CodeExceedsSupplyLimit,
		fmt.Sprintf("a supply increase of %s puts current asset supply %s over supply limit %s", increase, current, limit))
}

// ErrExceedsAvailableSupply error for when the proposed outgoing amount exceeds the total available supply
func ErrExceedsAvailableSupply(codespace sdk.CodespaceType, increase sdk.Coin, available sdk.Int) error {
	return sdk.NewError(codespace, CodeExceedsAvailableSupply,
		fmt.Sprintf("an outgoing swap with amount %s exceeds total available supply %s",
			increase, sdk.NewCoin(increase.Denom, available)))
}

// ErrInvalidCurrentSupply error for when the proposed decrease would result in a negative current supply
func ErrInvalidCurrentSupply(codespace sdk.CodespaceType, decrease, current sdk.Coin) error {
	return sdk.NewError(codespace, CodeInvalidCurrentSupply,
		fmt.Sprintf("a supply decrease of %s puts current asset supply %s below 0", decrease, current))
}

// ErrInvalidIncomingSupply error for when the proposed decrease would result in a negative incoming supply
func ErrInvalidIncomingSupply(codespace sdk.CodespaceType, decrease, incoming sdk.Coin) error {
	return sdk.NewError(codespace, CodeInvalidIncomingSupply,
		fmt.Sprintf("a supply decrease of %s puts incoming asset supply %s below 0", decrease, incoming))
}

// ErrInvalidOutgoingSupply error for when the proposed decrease would result in a negative outgoing supply
func ErrInvalidOutgoingSupply(codespace sdk.CodespaceType, decrease, outgoing sdk.Coin) error {
	return sdk.NewError(codespace, CodeInvalidOutgoingSupply,
		fmt.Sprintf("a supply decrease of %s puts outgoing asset supply %s below 0", decrease, outgoing))
}

// ErrInvalidClaimSecret error when a submitted secret doesn't match an AtomicSwap's swapID
func ErrInvalidClaimSecret(codespace sdk.CodespaceType, submittedSecret []byte, swapID []byte) error {
	return sdk.NewError(codespace, CodeInvalidClaimSecret,
		fmt.Sprintf("hashed claim attempt %s does not match %s",
			hex.EncodeToString(submittedSecret),
			hex.EncodeToString(swapID),
		),
	)
}

// ErrAtomicSwapAlreadyExists error for when an AtomicSwap with this swapID already exists
func ErrAtomicSwapAlreadyExists(codespace sdk.CodespaceType, swapID cmn.HexBytes) error {
	return sdk.NewError(codespace, CodeAtomicSwapAlreadyExists, fmt.Sprintf("atomic swap %s already exists", swapID))
}

// ErrAtomicSwapNotFound error for when an atomic swap is not found
func ErrAtomicSwapNotFound(codespace sdk.CodespaceType, id []byte) error {
	return sdk.NewError(codespace, CodeAtomicSwapNotFound, fmt.Sprintf("AtomicSwap %s was not found", hex.EncodeToString(id)))
}

// ErrSwapNotRefundable error for when an AtomicSwap has not expired and cannot be refunded
func ErrSwapNotRefundable(codespace sdk.CodespaceType) error {
	return sdk.NewError(codespace, CodeSwapNotRefundable, fmt.Sprintf("atomic swap is still active and cannot be refunded"))
}

// ErrSwapNotClaimable error for when an atomic swap is not open and cannot be claimed
func ErrSwapNotClaimable(codespace sdk.CodespaceType) error {
	return sdk.NewError(codespace, CodeSwapNotClaimable, fmt.Sprintf("atomic swap is not claimable"))
}
