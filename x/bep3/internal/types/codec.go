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
	cdc.RegisterConcrete(MsgHTLC{}, "bep3/MsgHTLC", nil)
	cdc.RegisterConcrete(MsgDepositHTLC{}, "bep3/MsgDepositHTLC", nil)
	cdc.RegisterConcrete(MsgClaimHTLC{}, "bep3/MsgClaimHTLC", nil)
	cdc.RegisterConcrete(MsgRefundHTLC{}, "bep3/MsgRefundHTLC", nil)
}
