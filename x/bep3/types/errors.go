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
	DefaultCodespace                sdk.CodespaceType = ModuleName
	CodeInvalidTimestamp            CodeType          = 1
	CodeInvalidHeightSpan           CodeType          = 2
	CodeAssetNotSupported           CodeType          = 3
	CodeAssetNotActive              CodeType          = 4
	CodeAboveTotalAssetSupplyLimit  CodeType          = 5
	CodeAboveAssetActiveSupplyLimit CodeType          = 6
	CodeInvalidClaimSecret          CodeType          = 7
	CodeAtomicSwapAlreadyExists     CodeType          = 8
	CodeAtomicSwapNotFound          CodeType          = 9
	CodeSwapNotRefundable           CodeType          = 10
	CodeSwapNotClaimable            CodeType          = 11
)

// ErrInvalidTimestamp error for when an timestamp is outside of bounds. Assumes block time of 10 seconds.
func ErrInvalidTimestamp(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidTimestamp, fmt.Sprintf("Timestamp can neither be 15 minutes ahead of the current time, nor 30 minutes later"))
}

// ErrInvalidHeightSpan error a proposed height span is outside of lock time range
func ErrInvalidHeightSpan(codespace sdk.CodespaceType, heightspan int64, minLockTime int64, maxLockTime int64) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidHeightSpan, fmt.Sprintf("height span %d is outside acceptable range %d - %d", heightspan, minLockTime, maxLockTime))
}

// ErrAssetNotSupported error for when an asset is not supported
func ErrAssetNotSupported(codespace sdk.CodespaceType, denom string) sdk.Error {
	return sdk.NewError(codespace, CodeAssetNotSupported, fmt.Sprintf("asset %s is not on the list of supported assets", denom))
}

// ErrAssetNotActive error for when an asset is currently inactive
func ErrAssetNotActive(codespace sdk.CodespaceType, denom string) sdk.Error {
	return sdk.NewError(codespace, CodeAssetNotActive, fmt.Sprintf("asset %s is currently inactive", denom))
}

// ErrAboveTotalAssetSupplyLimit error for when a proposed swap's amount is greater than the total supply limit (amount in swaps + amount active)
func ErrAboveTotalAssetSupplyLimit(codespace sdk.CodespaceType, denom string, supplyLimit, currAssetSupply, currInSwapSupply sdk.Int) sdk.Error {
	return sdk.NewError(codespace, CodeAboveTotalAssetSupplyLimit,
		fmt.Sprintf("%s proposed supply increase is over the supply limit of %d - current supply is %d, amount in active swaps is %d",
			denom, supplyLimit.Int64(), currAssetSupply.Int64(), currInSwapSupply.Int64()))
}

// ErrAboveAssetActiveSupplyLimit error for when the swap amount of an attempted claim is greater than active supply limit
func ErrAboveAssetActiveSupplyLimit(codespace sdk.CodespaceType, denom string, supplyLimit, currAssetSupply sdk.Int) sdk.Error {
	return sdk.NewError(codespace, CodeAboveAssetActiveSupplyLimit,
		fmt.Sprintf("%s proposed supply increase is over the supply limit of %d - current supply is %d",
			denom, supplyLimit.Int64(), currAssetSupply.Int64()))
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
