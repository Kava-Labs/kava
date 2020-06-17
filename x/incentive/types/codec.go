package types

import "github.com/cosmos/cosmos-sdk/codec"

// ModuleCdc generic sealed codec to be used throughout module
var ModuleCdc *codec.Codec

func init() {
	cdc := codec.New()
	RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	cdc.Seal()
	ModuleCdc = cdc
}

// RegisterCodec registers the necessary types for incentive module
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgClaimReward{}, "incentive/MsgClaimReward", nil)
	cdc.RegisterConcrete(GenesisClaimPeriodID{}, "incentive/GenesisClaimPeriodID", nil)
	cdc.RegisterConcrete(RewardPeriod{}, "incentive/RewardPeriod", nil)
	cdc.RegisterConcrete(ClaimPeriod{}, "incentive/ClaimPeriod", nil)
	cdc.RegisterConcrete(Claim{}, "incentive/Claim", nil)
}
