package util

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
)

func SdkToEvmAddress(addr sdk.AccAddress) common.Address {
	return common.BytesToAddress(addr.Bytes())
}

func EvmToSdkAddress(addr common.Address) sdk.AccAddress {
	return sdk.AccAddress(addr.Bytes())
}
