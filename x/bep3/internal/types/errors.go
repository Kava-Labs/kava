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

	// CodeInvalid      CodeType = 101
)

// TODO: Fill out some custom errors for the module
// You can see how they are constructed below:
// func ErrInvalid(codespace sdk.CodespaceType) sdk.Error {
// 	return sdk.NewError(codespace, CodeInvalid, "custom error message")
// }
