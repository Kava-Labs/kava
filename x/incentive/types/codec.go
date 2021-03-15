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
	cdc.RegisterInterface((*Claim)(nil), nil)
	cdc.RegisterConcrete(USDXMintingClaim{}, "incentive/USDXMintingClaim", nil)
	cdc.RegisterConcrete(HardLiquidityProviderClaim{}, "incentive/HardLiquidityProviderClaim", nil)

	// Register msgs
	cdc.RegisterConcrete(MsgClaimUSDXMintingReward{}, "incentive/MsgClaimUSDXMintingReward", nil)
	cdc.RegisterConcrete(MsgClaimHardReward{}, "incentive/MsgClaimHardReward", nil)
}
