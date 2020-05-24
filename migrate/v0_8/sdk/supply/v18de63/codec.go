package v18de63

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// RegisterCodec registers the account types and interface
func RegisterCodec(cdc *codec.Codec) {
	// cdc.RegisterInterface((*exported.ModuleAccountI)(nil), nil) // these types are unnecessary for unmarshalling a genesis.json
	// cdc.RegisterInterface((*exported.SupplyI)(nil), nil)
	cdc.RegisterConcrete(&ModuleAccount{}, "cosmos-sdk/ModuleAccount", nil)
	// cdc.RegisterConcrete(&Supply{}, "cosmos-sdk/Supply", nil)
}
