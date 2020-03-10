package types

import "github.com/cosmos/cosmos-sdk/codec"

// ModuleCdc generic sealed codec to be used throughout module
var ModuleCdc *codec.Codec

func init() {
	cdc := codec.New()
	RegisterCodec(cdc)
	ModuleCdc = cdc.Seal()
}

// RegisterCodec registers the necessary types for the module
func RegisterCodec(cdc *codec.Codec) {
	// TODO
	// cdc.RegisterConcrete(MsgCreateCDP{}, "cdp/MsgCreateCDP", nil)
	// cdc.RegisterConcrete(MsgDeposit{}, "cdp/MsgDeposit", nil)
	// cdc.RegisterConcrete(MsgWithdraw{}, "cdp/MsgWithdraw", nil)
	// cdc.RegisterConcrete(MsgDrawDebt{}, "cdp/MsgDrawDebt", nil)
	// cdc.RegisterConcrete(MsgRepayDebt{}, "cdp/MsgRepayDebt", nil)
}
