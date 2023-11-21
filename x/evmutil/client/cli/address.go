package cli

import (
	"context"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/kava-labs/kava/x/evmutil/types"
)

// ParseAddrFromHexOrBech32 parses a string address that can be either a hex or
// Bech32 string.
func ParseAddrFromHexOrBech32(addrString string) (common.Address, error) {
	if common.IsHexAddress(addrString) {
		return common.HexToAddress(addrString), nil
	}

	cfg := sdk.GetConfig()

	if !strings.HasPrefix(addrString, cfg.GetBech32AccountAddrPrefix()) {
		return common.Address{}, fmt.Errorf("receiver '%s' is not a hex or bech32 address (prefix does not match)", addrString)
	}

	accAddr, err := sdk.AccAddressFromBech32(addrString)
	if err != nil {
		return common.Address{}, fmt.Errorf("receiver '%s' is not a hex or bech32 address (could not parse as bech32 string)", addrString)
	}

	return common.BytesToAddress(accAddr), nil

}

// ParseOrQueryConversionPairAddress returns an EVM address of the provided
// ERC20 contract address string or denom. If an address string, just returns
// the parsed address. If a denom, fetches params, searches the enabled
// conversion pairs, and returns corresponding ERC20 contract address.
func ParseOrQueryConversionPairAddress(
	queryClient types.QueryClient,
	addrOrDenom string,
) (common.Address, error) {
	if common.IsHexAddress(addrOrDenom) {
		return common.HexToAddress(addrOrDenom), nil
	}

	if err := sdk.ValidateDenom(addrOrDenom); err != nil {
		return common.Address{}, fmt.Errorf(
			"Kava ERC20 '%s' is not a valid hex address or denom",
			addrOrDenom,
		)
	}

	// Valid denom, try looking up as denom to get corresponding Kava ERC20 address
	paramsRes, err := queryClient.Params(
		context.Background(),
		&types.QueryParamsRequest{},
	)
	if err != nil {
		return common.Address{}, err
	}

	for _, pair := range paramsRes.Params.EnabledConversionPairs {
		if pair.Denom == addrOrDenom {
			return pair.GetAddress().Address, nil
		}
	}

	return common.Address{}, fmt.Errorf(
		"Kava ERC20 '%s' is not a valid hex address or denom (did not match any denoms in queried enabled conversion pairs)",
		addrOrDenom,
	)
}
