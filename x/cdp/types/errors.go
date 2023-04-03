package types

import errorsmod "cosmossdk.io/errors"

// DONTCOVER

var (
	// ErrCdpAlreadyExists error for duplicate cdps
	ErrCdpAlreadyExists = errorsmod.Register(ModuleName, 2, "cdp already exists")
	// ErrInvalidCollateralLength error for invalid collateral input length
	ErrInvalidCollateralLength = errorsmod.Register(ModuleName, 3, "only one collateral type per cdp")
	// ErrCollateralNotSupported error for unsupported collateral
	ErrCollateralNotSupported = errorsmod.Register(ModuleName, 4, "collateral not supported")
	// ErrDebtNotSupported error for unsupported debt
	ErrDebtNotSupported = errorsmod.Register(ModuleName, 5, "debt not supported")
	// ErrExceedsDebtLimit error for attempted draws that exceed debt limit
	ErrExceedsDebtLimit = errorsmod.Register(ModuleName, 6, "proposed debt increase would exceed debt limit")
	// ErrInvalidCollateralRatio error for attempted draws that are below liquidation ratio
	ErrInvalidCollateralRatio = errorsmod.Register(ModuleName, 7, "proposed collateral ratio is below liquidation ratio")
	// ErrCdpNotFound error cdp not found
	ErrCdpNotFound = errorsmod.Register(ModuleName, 8, "cdp not found")
	// ErrDepositNotFound error for deposit not found
	ErrDepositNotFound = errorsmod.Register(ModuleName, 9, "deposit not found")
	// ErrInvalidDeposit error for invalid deposit
	ErrInvalidDeposit = errorsmod.Register(ModuleName, 10, "invalid deposit")
	// ErrInvalidPayment error for invalid payment
	ErrInvalidPayment = errorsmod.Register(ModuleName, 11, "invalid payment")
	// ErrDepositNotAvailable error for withdrawing deposits in liquidation
	ErrDepositNotAvailable = errorsmod.Register(ModuleName, 12, "deposit in liquidation")
	// ErrInvalidWithdrawAmount error for invalid withdrawal amount
	ErrInvalidWithdrawAmount = errorsmod.Register(ModuleName, 13, "withdrawal amount exceeds deposit")
	// ErrCdpNotAvailable error for depositing to a CDP in liquidation
	ErrCdpNotAvailable = errorsmod.Register(ModuleName, 14, "cannot modify cdp in liquidation")
	// ErrBelowDebtFloor error for creating a cdp with debt below the minimum
	ErrBelowDebtFloor = errorsmod.Register(ModuleName, 15, "proposed cdp debt is below minimum")
	// ErrLoadingAugmentedCDP error loading augmented cdp
	ErrLoadingAugmentedCDP = errorsmod.Register(ModuleName, 16, "augmented cdp could not be loaded from cdp")
	// ErrInvalidDebtRequest error for invalid principal input length
	ErrInvalidDebtRequest = errorsmod.Register(ModuleName, 17, "only one principal type per cdp")
	// ErrDenomPrefixNotFound error for denom prefix not found
	ErrDenomPrefixNotFound = errorsmod.Register(ModuleName, 18, "denom prefix not found")
	// ErrPricefeedDown error for when a price for the input denom is not found
	ErrPricefeedDown = errorsmod.Register(ModuleName, 19, "no price found for collateral")
	// ErrInvalidCollateral error for when the input collateral denom does not match the expected collateral denom
	ErrInvalidCollateral = errorsmod.Register(ModuleName, 20, "invalid collateral for input collateral type")
	// ErrAccountNotFound error for when no account is found for an input address
	ErrAccountNotFound = errorsmod.Register(ModuleName, 21, "account not found")
	// ErrInsufficientBalance error for when an account does not have enough funds
	ErrInsufficientBalance = errorsmod.Register(ModuleName, 22, "insufficient balance")
	// ErrNotLiquidatable error for when an cdp is not liquidatable
	ErrNotLiquidatable = errorsmod.Register(ModuleName, 23, "cdp collateral ratio not below liquidation ratio")
)
