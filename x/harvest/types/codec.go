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

// RegisterCodec registers the necessary types for harvest module
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgClaimReward{}, "harvest/MsgClaimReward", nil)
	cdc.RegisterConcrete(MsgDeposit{}, "harvest/MsgDeposit", nil)
	cdc.RegisterConcrete(MsgWithdraw{}, "harvest/MsgWithdraw", nil)
	cdc.RegisterConcrete(MsgBorrow{}, "harvest/MsgBorrow", nil)
	cdc.RegisterConcrete(MsgLiquidate{}, "harvest/MsgLiquidate", nil)
	cdc.RegisterConcrete(DistributionSchedule{}, "harvest/DistributionSchedule", nil)
}
