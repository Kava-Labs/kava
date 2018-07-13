package paychan

import (
	"github.com/cosmos/cosmos-sdk/wire"
)

func RegisterWire(cdc *wire.Codec) {
	cdc.RegisterConcrete(MsgCreate{}, "paychan/MsgCreate", nil)
	cdc.RegisterConcrete(MsgClose{}, "paychan/MsgClose", nil)
}

var msgCdc = wire.NewCodec()

func init() {
	RegisterWire(msgCdc)
	// TODO is this needed?
	//wire.RegisterCrypto(msgCdc)
}
