package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName The name that will be used throughout the module
	ModuleName = "liquidstaking"

	// StoreKey Top level store key where all module items will be stored
	StoreKey = ModuleName

	// RouterKey Top level router key
	RouterKey = ModuleName

	// QuerierRoute is the querier route for gov
	QuerierRoute = ModuleName

	// DefaultParamspace default namestore
	DefaultParamspace = ModuleName

	// ModuleAccountName is the module account's name
	ModuleAccountName = ModuleName
)

var DelegationHoldersKeyPrefix = []byte{0x01}

func GetLiquidStakingTokenDenom(bondDenom string, valAddr sdk.ValAddress) string {
	return fmt.Sprintf("%s-%s", bondDenom, valAddr.String())
}

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
