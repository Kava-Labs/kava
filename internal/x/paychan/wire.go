package paychan

import (
	"github.com/cosmos/cosmos-sdk/wire"
)

func RegisterWire(cdc *wire.Codec) {
	cdc.RegisterConcrete(MsgCreate{}, "paychan/MsgCreate", nil)
	cdc.RegisterConcrete(MsgSubmitUpdate{}, "paychan/MsgSubmitUpdate", nil)
}

// TODO move this to near the msg definitions?
var msgCdc = wire.NewCodec()

func init() {
	wire.RegisterCrypto(msgCdc)
	RegisterWire(msgCdc)
}
