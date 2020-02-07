package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
)

var ModuleCdc = codec.New()

func init() {
	RegisterCodec(ModuleCdc)
}

// To match codec registration msg names: https://github.com/binance-chain/go-sdk/blob/master/types/msg/wire.go
const (
	AliasMsgCreateHTLT  = "HTLTMsg"
	AliasMsgDepositHTLT = "DepositHTLTMsg"
	AliasMsgClaimHTLT   = "ClaimHTLTMsg"
	AliasMsgRefundHTLT  = "RefundHTLTMsg"
)

// RegisterCodec registers concrete types on the Amino code
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgCreateHTLT{}, fmt.Sprintf("bep3/%s", AliasMsgCreateHTLT), nil)
	cdc.RegisterConcrete(MsgDepositHTLT{}, fmt.Sprintf("bep3/%s", AliasMsgDepositHTLT), nil)
	cdc.RegisterConcrete(MsgClaimHTLT{}, fmt.Sprintf("bep3/%s", AliasMsgClaimHTLT), nil)
	cdc.RegisterConcrete(MsgRefundHTLT{}, fmt.Sprintf("bep3/%s", AliasMsgRefundHTLT), nil)
}
