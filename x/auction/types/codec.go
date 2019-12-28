package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// ModuleCdc module level codec
var ModuleCdc = codec.New()

func init() {
	RegisterCodec(ModuleCdc)
}

// RegisterCodec registers concrete types on the codec.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgPlaceBid{}, "auction/MsgPlaceBid", nil)

	// Register the Auction interface and concrete types
	cdc.RegisterInterface((*Auction)(nil), nil)
	cdc.RegisterConcrete(ForwardAuction{}, "auction/ForwardAuction", nil)
	cdc.RegisterConcrete(ReverseAuction{}, "auction/ReverseAuction", nil)
	cdc.RegisterConcrete(ForwardReverseAuction{}, "auction/ForwardReverseAuction", nil)
}
