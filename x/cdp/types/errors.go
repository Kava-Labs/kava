package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// DONTCOVER

var (
	// ErrCdpAlreadyExists error for duplicate cdps
	ErrCdpAlreadyExists = sdkerrors.Register(ModuleName, 2, "cdp already exists")
	// ErrInvalidCollateralLength error for invalid collateral input length
	ErrInvalidCollateralLength = sdkerrors.Register(ModuleName, 3, "only one collateral type per cdp")
	// ErrCollateralNotSupported error for unsupported collateral
	ErrCollateralNotSupported = sdkerrors.Register(ModuleName, 4, "collateral not supported")
	// ErrDebtNotSupported error for unsupported debt
	ErrDebtNotSupported = sdkerrors.Register(ModuleName, 5, "debt not supported")
	// ErrExceedsDebtLimit error for attempted draws that exceed debt limit
	ErrExceedsDebtLimit = sdkerrors.Register(ModuleName, 6, "proposed debt increase would exceed debt limit")
	// ErrInvalidCollateralRatio error for attempted draws that are below liquidation ratio
	ErrInvalidCollateralRatio = sdkerrors.Register(ModuleName, 7, "proposed collateral ratio is below liquidation ratio")
	// ErrCdpNotFound error cdp not found
	ErrCdpNotFound = sdkerrors.Register(ModuleName, 8, "cdp not found")
	// ErrDepositNotFound error for deposit not found
	ErrDepositNotFound = sdkerrors.Register(ModuleName, 9, "deposit not found")
	// ErrInvalidDeposit error for invalid deposit
	ErrInvalidDeposit = sdkerrors.Register(ModuleName, 10, "invalid deposit")
	// ErrInvalidCollateral error for invalid collateral
	ErrInvalidCollateral = sdkerrors.Register(ModuleName, 11, "collateral not supported")
	// ErrInvalidPayment error for invalid payment
	ErrInvalidPayment = sdkerrors.Register(ModuleName, 12, "invalid payment")
	//ErrDepositNotAvailable error for withdrawing deposits in liquidation
	ErrDepositNotAvailable = sdkerrors.Register(ModuleName, 13, "deposit in liquidation")
	// ErrInvalidWithdrawAmount error for invalid withdrawal amount
	ErrInvalidWithdrawAmount = sdkerrors.Register(ModuleName, 14, "withdrawal amount exceeds deposit")
	//ErrCdpNotAvailable error for depositing to a CDP in liquidation
	ErrCdpNotAvailable = sdkerrors.Register(ModuleName, 15, "cannot modify cdp in liquidation")
	// ErrBelowDebtFloor error for creating a cdp with debt below the minimum
	ErrBelowDebtFloor = sdkerrors.Register(ModuleName, 16, "proposed cdp debt is below minimum")
	// ErrLoadingAugmentedCDP error loading augmented cdp
	ErrLoadingAugmentedCDP = sdkerrors.Register(ModuleName, 17, "augmented cdp could not be loaded from cdp")
	// ErrInvalidDebtRequest error for invalid principal input length
	ErrInvalidDebtRequest = sdkerrors.Register(ModuleName, 18, "only one principal type per cdp")
)
