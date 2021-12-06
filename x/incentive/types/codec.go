package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"

	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterLegacyAminoCodec registers all the necessary types and interfaces for the
// governance module.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgClaimUSDXMintingReward{}, "incentive/MsgClaimUSDXMintingReward", nil)
	cdc.RegisterConcrete(&MsgClaimHardReward{}, "incentive/MsgClaimHardReward", nil)
	cdc.RegisterConcrete(&MsgClaimDelegatorReward{}, "incentive/MsgClaimDelegatorReward", nil)
	cdc.RegisterConcrete(&MsgClaimSwapReward{}, "incentive/MsgClaimSwapReward", nil)
	cdc.RegisterConcrete(&MsgClaimUSDXMintingRewardVVesting{}, "incentive/MsgClaimUSDXRewardVVesting", nil)
	cdc.RegisterConcrete(&MsgClaimHardRewardVVesting{}, "incentive/MsgClaimHardRewardVVesting", nil)
	cdc.RegisterConcrete(&MsgClaimDelegatorRewardVVesting{}, "incentive/MsgClaimDelegatorRewardVVesting", nil)
	cdc.RegisterConcrete(&MsgClaimSwapRewardVVesting{}, "incentive/MsgClaimSwapRewardVVesting", nil)

	cdc.RegisterInterface((*Claim)(nil), nil)

	cdc.RegisterConcrete(&USDXMintingClaim{}, "incentive/USDXMintingClaim", nil)
	cdc.RegisterConcrete(&HardLiquidityProviderClaim{}, "incentive/HardLiquidityProviderClaim", nil)
	cdc.RegisterConcrete(&DelegatorClaim{}, "incentive/DelegatorClaim", nil)
	cdc.RegisterConcrete(&SwapClaim{}, "incentive/SwapClaim", nil)
}

func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgClaimUSDXMintingReward{},
		&MsgClaimHardReward{},
		&MsgClaimDelegatorReward{},
		&MsgClaimSwapReward{},
		&MsgClaimUSDXMintingRewardVVesting{},
		&MsgClaimHardRewardVVesting{},
		&MsgClaimDelegatorRewardVVesting{},
		&MsgClaimSwapRewardVVesting{},
	)

	registry.RegisterInterface(
		"kava.auction.v1beta1.Claim",
		(*Claim)(nil),
		&USDXMintingClaim{},
		// TODO: These Claims do not actually match Claim interface, GetReward()
		// responds with sdk.Coins instead of a single sdk.Coin
		// &HardLiquidityProviderClaim{},
		// &DelegatorClaim{},
		// &SwapClaim{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	amino = codec.NewLegacyAmino()

	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
}
