package types

import (
	"fmt"
	"strings"

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

// ParseLiquidStakingTokenDenom extracts a validator address from a derivative denom.
func ParseLiquidStakingTokenDenom(denom string) (sdk.ValAddress, error) {
	elements := strings.Split(denom, "-")
	if len(elements) != 2 {
		return nil, fmt.Errorf("cannot parse denom %s", denom)
	}
	addr, err := sdk.ValAddressFromBech32(elements[1])
	if err != nil {
		return nil, err
	}
	return addr, nil
}
