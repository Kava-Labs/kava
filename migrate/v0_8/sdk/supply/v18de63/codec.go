package v18de63

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// Note some types are unnecessary for unmarshalling genesis.json so have not been registered

// RegisterCodec registers the account types and interface
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(&ModuleAccount{}, "cosmos-sdk/ModuleAccount", nil)
}
