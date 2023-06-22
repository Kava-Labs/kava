package util

import (
	"github.com/ethereum/go-ethereum/common"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func SdkToEvmAddress(addr sdk.AccAddress) common.Address {
	return common.BytesToAddress(addr.Bytes())
}

func EvmToSdkAddress(addr common.Address) sdk.AccAddress {
	return sdk.AccAddress(addr.Bytes())
}
