package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName The name that will be used throughout the module
	ModuleName = "liquid"

	// RouterKey Top level router key
	RouterKey = ModuleName

	// ModuleAccountName is the module account's name
	ModuleAccountName = ModuleName

	DefaultDerivativeDenom = "bkava"

	DenomSeparator = "-"
)

func GetLiquidStakingTokenDenom(bondDenom string, valAddr sdk.ValAddress) string {
	return fmt.Sprintf("%s%s%s", bondDenom, DenomSeparator, valAddr.String())
}
