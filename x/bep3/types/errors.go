package types

import errorsmod "cosmossdk.io/errors"

// DONTCOVER

var (
	// ErrInvalidTimestamp error for when an timestamp is outside of bounds. Assumes block time of 10 seconds.
	ErrInvalidTimestamp = errorsmod.Register(ModuleName, 2, "timestamp can neither be 15 minutes ahead of the current time, nor 30 minutes later")
	// ErrInvalidHeightSpan error for when a proposed height span is outside of lock time range
	ErrInvalidHeightSpan = errorsmod.Register(ModuleName, 3, "height span is outside acceptable range")
	// ErrInsufficientAmount error for when a swap's amount cannot cover the deputy's fixed fee
	ErrInsufficientAmount = errorsmod.Register(ModuleName, 4, "amount cannot cover the deputy fixed fee")
	// ErrAssetNotSupported error for when an asset is not supported
	ErrAssetNotSupported = errorsmod.Register(ModuleName, 5, "asset not found")
	// ErrAssetNotActive error for when an asset is currently inactive
	ErrAssetNotActive = errorsmod.Register(ModuleName, 6, "asset is currently inactive")
	// ErrAssetSupplyNotFound error for when an asset's supply is not found in the store
	ErrAssetSupplyNotFound = errorsmod.Register(ModuleName, 7, "asset supply not found in store")
	// ErrExceedsSupplyLimit error for when the proposed supply increase would put the supply above limit
	ErrExceedsSupplyLimit = errorsmod.Register(ModuleName, 8, "asset supply over limit")
	// ErrExceedsAvailableSupply error for when the proposed outgoing amount exceeds the total available supply
	ErrExceedsAvailableSupply = errorsmod.Register(ModuleName, 9, "outgoing swap exceeds total available supply")
	// ErrInvalidCurrentSupply error for when the proposed decrease would result in a negative current supplyx
	ErrInvalidCurrentSupply = errorsmod.Register(ModuleName, 10, "supply decrease puts current asset supply below 0")
	// ErrInvalidIncomingSupply error for when the proposed decrease would result in a negative incoming supply
	ErrInvalidIncomingSupply = errorsmod.Register(ModuleName, 11, "supply decrease puts incoming asset supply below 0")
	// ErrInvalidOutgoingSupply error for when the proposed decrease would result in a negative outgoing supply
	ErrInvalidOutgoingSupply = errorsmod.Register(ModuleName, 12, "supply decrease puts outgoing asset supply below 0")
	// ErrInvalidClaimSecret error when a submitted secret doesn't match an AtomicSwap's swapID
	ErrInvalidClaimSecret = errorsmod.Register(ModuleName, 13, "hashed claim attempt does not match")
	// ErrAtomicSwapAlreadyExists error for when an AtomicSwap with this swapID already exists
	ErrAtomicSwapAlreadyExists = errorsmod.Register(ModuleName, 14, "atomic swap already exists")
	// ErrAtomicSwapNotFound error for when an atomic swap is not found
	ErrAtomicSwapNotFound = errorsmod.Register(ModuleName, 15, "atomic swap not found")
	// ErrSwapNotRefundable error for when an AtomicSwap has not expired and cannot be refunded
	ErrSwapNotRefundable = errorsmod.Register(ModuleName, 16, "atomic swap is still active and cannot be refunded")
	// ErrSwapNotClaimable error for when an atomic swap is not open and cannot be claimed
	ErrSwapNotClaimable = errorsmod.Register(ModuleName, 17, "atomic swap is not claimable")
	// ErrInvalidAmount error for when a swap's amount is outside acceptable range
	ErrInvalidAmount = errorsmod.Register(ModuleName, 18, "amount is outside acceptable range")
	// ErrInvalidSwapAccount error for when a swap involves an invalid account
	ErrInvalidSwapAccount = errorsmod.Register(ModuleName, 19, "atomic swap has invalid account")
	// ErrExceedsTimeBasedSupplyLimit error for when the proposed supply increase would put the supply above limit for the current time period
	ErrExceedsTimeBasedSupplyLimit = errorsmod.Register(ModuleName, 20, "asset supply over limit for current time period")
)
