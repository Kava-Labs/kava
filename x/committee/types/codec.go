package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// ModuleCdc generic sealed codec to be used throughout module
var ModuleCdc *codec.Codec

func init() {
	cdc := codec.New()
	RegisterCodec(cdc)
	ModuleCdc = cdc.Seal()
}

// RegisterCodec registers the necessary types for the module
func RegisterCodec(cdc *codec.Codec) {

	// TODO need to register Content interface, however amino panics if you try and register it twice and helpfully doesn't provide a way to query registered types
	//cdc.RegisterInterface((*gov.Content)(nil), nil)

	cdc.RegisterInterface((*Permission)(nil), nil)
	cdc.RegisterConcrete(GodPermission{}, "kava/GodPermission", nil)
}
