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

// RegisterCodec registers the necessary types for incentive module
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgClaimReward{}, "incentive/MsgClaimReward", nil)
	cdc.RegisterConcrete(GenesisAccumulationTime{}, "incentive/GenesisAccumulationTime", nil)
	cdc.RegisterConcrete(GenesisAccumulationTimes{}, "incentive/GenesisAccumulationTimes", nil)
	cdc.RegisterConcrete(RewardPeriod{}, "incentive/RewardPeriod", nil)
	cdc.RegisterConcrete(USDXMintingClaim{}, "incentive/USDXMintingClaim", nil)
	cdc.RegisterConcrete(USDXMintingClaims{}, "incentive/USDXMintingClaims", nil)

}
