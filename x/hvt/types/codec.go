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

// RegisterCodec registers the necessary types for hvt module
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgClaimReward{}, "hvt/MsgClaimReward", nil)
	cdc.RegisterConcrete(MsgDeposit{}, "hvt/MsgDeposit", nil)
	cdc.RegisterConcrete(DistributionSchedule{}, "hvt/DistributionSchedule", nil)
}
