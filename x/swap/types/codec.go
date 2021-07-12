package types

import "github.com/cosmos/cosmos-sdk/codec"

// ModuleCdc generic sealed codec to be used throughout module
var ModuleCdc *codec.Codec

func init() {
	cdc := codec.New()
	RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	ModuleCdc = cdc.Seal()
}

// RegisterCodec registers the necessary types for swap module
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgDeposit{}, "swap/MsgDeposit", nil)
	cdc.RegisterConcrete(MsgWithdraw{}, "swap/MsgWithdraw", nil)
}
