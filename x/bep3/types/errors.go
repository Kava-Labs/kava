package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// DONTCOVER

var (
	// ErrInvalidTimestamp error for when an timestamp is outside of bounds. Assumes block time of 10 seconds.
	ErrInvalidTimestamp = sdkerrors.Register(ModuleName, 2, "timestamp can neither be 15 minutes ahead of the current time, nor 30 minutes later")
	// ErrInvalidHeightSpan error for when a proposed height span is outside of lock time range
	ErrInvalidHeightSpan = sdkerrors.Register(ModuleName, 3, "height span is outside acceptable range")
	// ErrInsufficientAmount error for when a swap's amount cannot cover the deputy's fixed fee
	ErrInsufficientAmount = sdkerrors.Register(ModuleName, 4, "amount cannot cover the deputy fixed fee")
	// ErrAssetNotSupported error for when an asset is not supported
	ErrAssetNotSupported = sdkerrors.Register(ModuleName, 5, "asset not found")
	// ErrAssetNotActive error for when an asset is currently inactive
	ErrAssetNotActive = sdkerrors.Register(ModuleName, 6, "asset is currently inactive")
	// ErrAssetSupplyNotFound error for when an asset's supply is not found in the store
	ErrAssetSupplyNotFound = sdkerrors.Register(ModuleName, 7, "asset supply not found in store")
	// ErrExceedsSupplyLimit error for when the proposed supply increase would put the supply above limit
	ErrExceedsSupplyLimit = sdkerrors.Register(ModuleName, 8, "asset supply over limit")
	// ErrExceedsAvailableSupply error for when the proposed outgoing amount exceeds the total available supply
	ErrExceedsAvailableSupply = sdkerrors.Register(ModuleName, 9, "outgoing swap exceeds total available supply")
	// ErrInvalidCurrentSupply error for when the proposed decrease would result in a negative current supplyx
	ErrInvalidCurrentSupply = sdkerrors.Register(ModuleName, 10, "supply decrease puts current asset supply below 0")
	// ErrInvalidIncomingSupply error for when the proposed decrease would result in a negative incoming supply
	ErrInvalidIncomingSupply = sdkerrors.Register(ModuleName, 11, "supply decrease puts incoming asset supply below 0")
	// ErrInvalidOutgoingSupply error for when the proposed decrease would result in a negative outgoing supply
	ErrInvalidOutgoingSupply = sdkerrors.Register(ModuleName, 12, "supply decrease puts outgoing asset supply below 0")
	// ErrInvalidClaimSecret error when a submitted secret doesn't match an AtomicSwap's swapID
	ErrInvalidClaimSecret = sdkerrors.Register(ModuleName, 13, "hashed claim attempt does not match")
	// ErrAtomicSwapAlreadyExists error for when an AtomicSwap with this swapID already exists
	ErrAtomicSwapAlreadyExists = sdkerrors.Register(ModuleName, 14, "atomic swap already exists")
	// ErrAtomicSwapNotFound error for when an atomic swap is not found
	ErrAtomicSwapNotFound = sdkerrors.Register(ModuleName, 15, "atomic swap not found")
	// ErrSwapNotRefundable error for when an AtomicSwap has not expired and cannot be refunded
	ErrSwapNotRefundable = sdkerrors.Register(ModuleName, 16, "atomic swap is still active and cannot be refunded")
	// ErrSwapNotClaimable error for when an atomic swap is not open and cannot be claimed
	ErrSwapNotClaimable = sdkerrors.Register(ModuleName, 17, "atomic swap is not claimable")
	// ErrInvalidAmount error for when a swap's amount is outside acceptable range
	ErrInvalidAmount = sdkerrors.Register(ModuleName, 18, "amount is outside acceptable range")
	// ErrInvalidSwapAccount error for when a swap involves an invalid account
	ErrInvalidSwapAccount = sdkerrors.Register(ModuleName, 19, "atomic swap has invalid account")
)
