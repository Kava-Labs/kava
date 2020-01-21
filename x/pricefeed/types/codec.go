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
	cdc.RegisterConcrete(MsgPostPrice{}, "pricefeed/MsgPostPrice", nil)
}
