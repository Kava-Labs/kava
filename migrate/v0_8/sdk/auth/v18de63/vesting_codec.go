package v18de63

import (
	"github.com/cosmos/cosmos-sdk/codec"
	//"github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
)

// RegisterCodec registers concrete types on the codec
func RegisterCodecVesting(cdc *codec.Codec) { // renamed to avoid conflict as packages are combined
	cdc.RegisterInterface((*VestingAccount)(nil), nil)
	cdc.RegisterConcrete(&BaseVestingAccount{}, "cosmos-sdk/BaseVestingAccount", nil)
	cdc.RegisterConcrete(&ContinuousVestingAccount{}, "cosmos-sdk/ContinuousVestingAccount", nil)
	cdc.RegisterConcrete(&DelayedVestingAccount{}, "cosmos-sdk/DelayedVestingAccount", nil)
	cdc.RegisterConcrete(&PeriodicVestingAccount{}, "cosmos-sdk/PeriodicVestingAccount", nil)
}

// // VestingCdc module wide codec
// var VestingCdc *codec.Codec

// func init() {
// 	VestingCdc = codec.New()
// 	RegisterCodec(VestingCdc)
// 	VestingCdc.Seal()
// }
