package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

var ModuleCdc = codec.New()

func init() {
	RegisterCodec(ModuleCdc)
}

// RegisterCodec registers concrete types on the Amino code
func RegisterCodec(cdc *codec.Codec) {
	// TODO: Register messages
	// cdc.RegisterConcrete(HTLTMsg{}, "bep3/HTLTMsg", nil)
	// cdc.RegisterConcrete(DepositHTLTMsg{}, "bep3/DepositHTLTMsg", nil)
	// cdc.RegisterConcrete(ClaimHTLTMsg{}, "bep3/ClaimHTLTMsg", nil)
	// cdc.RegisterConcrete(RefundHTLTMsg{}, "bep3/RefundHTLTMsg", nil)
}
