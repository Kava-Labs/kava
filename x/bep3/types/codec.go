package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

var ModuleCdc = codec.New()

func init() {
	RegisterCodec(ModuleCdc)
}

// From https://github.com/binance-chain/go-sdk/blob/master/types/msg/wire.go:
// cdc.RegisterConcrete(HTLTMsg{}, "tokens/HTLTMsg", nil)
// cdc.RegisterConcrete(DepositHTLTMsg{}, "tokens/DepositHTLTMsg", nil)
// cdc.RegisterConcrete(ClaimHTLTMsg{}, "tokens/ClaimHTLTMsg", nil)
// cdc.RegisterConcrete(RefundHTLTMsg{}, "tokens/RefundHTLTMsg", nil)

// RegisterCodec registers concrete types on the Amino code
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(HTLTMsg{}, "bep3/HTLTMsg", nil)

	// cdc.RegisterConcrete(MsgDepositHTLT{}, fmt.Sprintf("bep3/%s", AliasMsgDepositHTLT), nil)
	// cdc.RegisterConcrete(MsgClaimHTLT{}, fmt.Sprintf("bep3/%s", AliasMsgClaimHTLT), nil)
	// cdc.RegisterConcrete(MsgRefundHTLT{}, fmt.Sprintf("bep3/%s", AliasMsgRefundHTLT), nil)
}
