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
	cdc.RegisterConcrete(DelegatorClaim{}, "incentive/DelegatorClaim", nil)
	cdc.RegisterConcrete(SwapClaim{}, "incentive/SwapClaim", nil)

	// Register msgs
	cdc.RegisterConcrete(MsgClaimUSDXMintingReward{}, "incentive/MsgClaimUSDXMintingReward", nil)
	cdc.RegisterConcrete(MsgClaimHardReward{}, "incentive/MsgClaimHardReward", nil)
	cdc.RegisterConcrete(MsgClaimDelegatorReward{}, "incentive/MsgClaimDelegatorReward", nil)
	cdc.RegisterConcrete(MsgClaimSwapReward{}, "incentive/MsgClaimSwapReward", nil)

	cdc.RegisterConcrete(MsgClaimUSDXMintingRewardVVesting{}, "incentive/MsgClaimUSDXRewardVVesting", nil)
	cdc.RegisterConcrete(MsgClaimHardRewardVVesting{}, "incentive/MsgClaimHardRewardVVesting", nil)
	cdc.RegisterConcrete(MsgClaimDelegatorRewardVVesting{}, "incentive/MsgClaimDelegatorRewardVVesting", nil)
	cdc.RegisterConcrete(MsgClaimSwapRewardVVesting{}, "incentive/MsgClaimSwapRewardVVesting", nil)
}
