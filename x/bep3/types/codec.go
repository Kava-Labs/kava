package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

var ModuleCdc = codec.New()

func init() {
	RegisterCodec(ModuleCdc)
}

// RegisterCodec registers concrete types on the Amino code
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgCreateAtomicSwap{}, "bep3/MsgCreateAtomicSwap", nil)
	cdc.RegisterConcrete(MsgDepositAtomicSwap{}, "bep3/MsgDepositAtomicSwap", nil)
	cdc.RegisterConcrete(MsgRefundAtomicSwap{}, "bep3/MsgRefundAtomicSwap", nil)
	cdc.RegisterConcrete(MsgClaimAtomicSwap{}, "bep3/MsgClaimAtomicSwap", nil)
}
