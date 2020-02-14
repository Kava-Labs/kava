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
	cdc.RegisterConcrete(MsgCreateHTLT{}, "bep3/MsgCreateHTLT", nil)
	cdc.RegisterConcrete(MsgDepositHTLT{}, "bep3/MsgDepositHTLT", nil)
	cdc.RegisterConcrete(MsgRefundHTLT{}, "bep3/MsgRefundHTLT", nil)
	cdc.RegisterConcrete(MsgClaimHTLT{}, "bep3/MsgClaimHTLT", nil)
}
