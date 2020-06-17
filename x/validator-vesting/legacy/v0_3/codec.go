package v18de63

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// RegisterCodec registers concrete types on the codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(&ValidatorVestingAccount{}, "cosmos-sdk/ValidatorVestingAccount", nil)
}
