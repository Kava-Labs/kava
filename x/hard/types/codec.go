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

// RegisterCodec registers the necessary types for hard module
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgClaimReward{}, "hard/MsgClaimReward", nil)
	cdc.RegisterConcrete(MsgDeposit{}, "hard/MsgDeposit", nil)
	cdc.RegisterConcrete(MsgWithdraw{}, "hard/MsgWithdraw", nil)
	cdc.RegisterConcrete(MsgBorrow{}, "hard/MsgBorrow", nil)
	cdc.RegisterConcrete(MsgLiquidate{}, "hard/MsgLiquidate", nil)
	cdc.RegisterConcrete(MsgRepay{}, "hard/MsgRepay", nil)
}
