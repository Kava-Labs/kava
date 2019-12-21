// DONTCOVER
package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// codes for cdp errors
const (
	DefaultCodespace            sdk.CodespaceType = ModuleName
	CodeCdpAlreadyExists        sdk.CodeType      = 1
	CodeCollateralLengthInvalid sdk.CodeType      = 2
	CodeCollateralNotSupported  sdk.CodeType      = 3
	CodeDebtNotSupported        sdk.CodeType      = 4
	CodeExceedsDebtLimit        sdk.CodeType      = 5
	CodeInvalidCollateralRatio  sdk.CodeType      = 6
	CodeCdpNotFound             sdk.CodeType      = 7
	CodeDepositNotFound         sdk.CodeType      = 8
	CodeInvalidDepositDenom     sdk.CodeType      = 9
	CodeInvalidPaymentDenom     sdk.CodeType      = 10
	CodeDepositNotAvailable     sdk.CodeType      = 11
	CodeInvalidCollateralDenom  sdk.CodeType      = 12
	CodeInvalidWithdrawAmount   sdk.CodeType      = 13
)

// ErrCdpAlreadyExists error for duplicate cdps
func ErrCdpAlreadyExists(codespace sdk.CodespaceType, owner sdk.AccAddress, denom string) sdk.Error {
	return sdk.NewError(codespace, CodeCdpAlreadyExists, fmt.Sprintf("cdp for owner %s and collateral %s already exists", owner, denom))
}

// ErrInvalidCollateralLength error for invalid collateral input length
func ErrInvalidCollateralLength(codespace sdk.CodespaceType, length int) sdk.Error {
	return sdk.NewError(codespace, CodeCollateralLengthInvalid, fmt.Sprintf("only one collateral type per cdp, has %d", length))
}

// ErrCollateralNotSupported error for unsupported collateral
func ErrCollateralNotSupported(codespace sdk.CodespaceType, denom string) sdk.Error {
	return sdk.NewError(codespace, CodeCollateralNotSupported, fmt.Sprintf("collateral %s not supported", denom))
}

// ErrDebtNotSupported error for unsupported collateral
func ErrDebtNotSupported(codespace sdk.CodespaceType, denom string) sdk.Error {
	return sdk.NewError(codespace, CodeDebtNotSupported, fmt.Sprintf("collateral %s not supported", denom))
}

// ErrExceedsDebtLimit error for attempted draws that exceed debt limit
func ErrExceedsDebtLimit(codespace sdk.CodespaceType, proposed sdk.Coins, limit sdk.Coins) sdk.Error {
	return sdk.NewError(codespace, CodeExceedsDebtLimit, fmt.Sprintf("proposed debt increase %s would exceed debt limit of %s", proposed, limit))
}

// ErrInvalidCollateralRatio error for attempted draws that are below liquidation ratio
func ErrInvalidCollateralRatio(codespace sdk.CodespaceType, denom string, collateralRatio sdk.Dec, liquidationRatio sdk.Dec) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidCollateralRatio, fmt.Sprintf("proposed collateral ratio of %s is below liqudation ratio of %s for collateral %s", collateralRatio, liquidationRatio, denom))
}

// ErrCdpNotFound error cdp not found
func ErrCdpNotFound(codespace sdk.CodespaceType, owner sdk.AccAddress, denom string) sdk.Error {
	return sdk.NewError(codespace, CodeCdpNotFound, fmt.Sprintf("cdp for owner %s and collateral %s not found", owner, denom))
}

// ErrDepositNotFound error for deposit not found
func ErrDepositNotFound(codespace sdk.CodespaceType, depositor sdk.AccAddress, cdpID uint64) sdk.Error {
	return sdk.NewError(codespace, CodeDepositNotFound, fmt.Sprintf("deposit for cdp %d not found for %s", cdpID, depositor))
}

// ErrInvalidDepositDenom error for invalid deposit denoms
func ErrInvalidDepositDenom(codespace sdk.CodespaceType, cdpID uint64, expected string, actual string) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidDepositDenom, fmt.Sprintf("invalid deposit for cdp %d, expects %s, got  %s", cdpID, expected, actual))
}

// ErrInvalidPaymentDenom error for invalid deposit denoms
func ErrInvalidPaymentDenom(codespace sdk.CodespaceType, cdpID uint64, expected []string, actual []string) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidPaymentDenom, fmt.Sprintf("invalid payment for cdp %d, expects %s, got  %s", cdpID, expected, actual))
}

//ErrDepositNotAvailable error for withdrawing deposits in liquidation
func ErrDepositNotAvailable(codespace sdk.CodespaceType, cdpID uint64, depositor sdk.AccAddress) sdk.Error {
	return sdk.NewError(codespace, CodeDepositNotAvailable, fmt.Sprintf("deposit from %s for cdp %d in liquidation", depositor, cdpID))
}

// ErrInvalidPaymentDenom error for invalid deposit denoms
func ErrInvalidCollateralDenom(codespace sdk.CodespaceType, denom string) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidDepositDenom, fmt.Sprintf("invalid denom:  %s", denom))
}

// ErrInvalidWithdrawAmount error for invalid withdrawal amount
func ErrInvalidWithdrawAmount(codespace sdk.CodespaceType, withdraw sdk.Coins, deposit sdk.Coins) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidWithdrawAmount, fmt.Sprintf("withdrawal amount of %s exceeds deposit of %s", withdraw, deposit))
}
