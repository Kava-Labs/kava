package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Local code type
type CodeType = sdk.CodeType

const (
	// Default bep3 codespace
	DefaultCodespace sdk.CodespaceType = ModuleName

	// CodeInvalidLockTime error code for lock time < min lock time
	CodeInvalidLockTime       CodeType = 1
	CodeInvalidInitialKHTLTID CodeType = 2
)

// ErrInvalidLockTime Error constructor
func ErrInvalidLockTime(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidLockTime, fmt.Sprintf("invalid lock time: must be greater than minimum lock time"))
}

// ErrInvalidInitialHTErrInvalidInitialKHTLTIDLTID Error constructor
func ErrInvalidInitialKHTLTID(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidInitialKHTLTID, fmt.Sprintf("invalid initial HTLT ID"))
}
