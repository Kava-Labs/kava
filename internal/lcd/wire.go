// Copyright 2016 All in Bits, inc
// Modifications copyright 2018 Kava Labs

package lcd

import (
	amino "github.com/tendermint/go-amino"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

var cdc = amino.NewCodec()

func init() {
	ctypes.RegisterAmino(cdc)
}
