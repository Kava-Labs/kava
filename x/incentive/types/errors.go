package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
)

// CodeType is the local code type
type CodeType = sdk.CodeType

// Error codes specific to incentive module
const (
	DefaultCodespace        sdk.CodespaceType = ModuleName
	CodeClaimNotFound       CodeType          = 1
	CodeClaimPeriodNotFound CodeType          = 2
	CodeInvalidAccountType  CodeType          = 3
	CodeNoClaimsFound       CodeType          = 4
)

// ErrClaimNotFound error for when a claim with a specific id and denom is not found for a particular owner
func ErrClaimNotFound(codespace sdk.CodespaceType, addr sdk.AccAddress, denom string, id uint64) sdk.Error {
	return sdk.NewError(codespace, CodeClaimNotFound, fmt.Sprintf("no claim with id %d and denom %s found for owner %s", id, denom, addr))
}

// ErrClaimPeriodNotFound error for when a claim period with a specific id and denom is not found
func ErrClaimPeriodNotFound(codespace sdk.CodespaceType, denom string, id uint64) sdk.Error {
	return sdk.NewError(codespace, CodeClaimNotFound, fmt.Sprintf("no claim period with id %d and denom %s found", id, denom))
}

// ErrInvalidAccountType error for invalid account type
func ErrInvalidAccountType(codespace sdk.CodespaceType, acc authexported.Account) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidAccountType, fmt.Sprintf("account type %T not supported", acc))
}

// ErrNoClaimsFound error for when no claims are found for the input denom and address
func ErrNoClaimsFound(codespace sdk.CodespaceType, addr sdk.AccAddress, denom string) sdk.Error {
	return sdk.NewError(codespace, CodeClaimNotFound, fmt.Sprintf("no claims with denom %s found for %s", denom, addr))
}

// ErrInsufficientModAccountBalance error for when module account has insufficient balance to pay claim
func ErrInsufficientModAccountBalance(codespace sdk.CodespaceType, name string) sdk.Error {
	return sdk.NewError(codespace, CodeClaimNotFound, fmt.Sprintf("module account %s has insufficient balance to pay claim", name))
}
