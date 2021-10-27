package app

import (
	"github.com/cosmos/cosmos-sdk/std"

	"github.com/kava-labs/kava/app/params"
)

// TODO why is this separate from params?

// MakeEncodingConfig creates an EncodingConfig and registers the app's types on it.
func MakeEncodingConfig() params.EncodingConfig {
	encodingConfig := params.MakeEncodingConfig()
	std.RegisterLegacyAminoCodec(encodingConfig.Amino)
	std.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	ModuleBasics.RegisterLegacyAminoCodec(encodingConfig.Amino)
	ModuleBasics.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	return encodingConfig
}
