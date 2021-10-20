package app

import sdk "github.com/cosmos/cosmos-sdk/types"

const (
	// Bech32MainPrefix defines the Bech32 prefix for account addresses
	Bech32MainPrefix = "kava"
	// Bech32PrefixAccPub defines the Bech32 prefix of an account's public key
	Bech32PrefixAccPub = Bech32MainPrefix + "pub"
	// Bech32PrefixValAddr defines the Bech32 prefix of a validator's operator address
	Bech32PrefixValAddr = Bech32MainPrefix + "val" + "oper"
	// Bech32PrefixValPub defines the Bech32 prefix of a validator's operator public key
	Bech32PrefixValPub = Bech32MainPrefix + "val" + "oper" + "pub"
	// Bech32PrefixConsAddr defines the Bech32 prefix of a consensus node address
	Bech32PrefixConsAddr = Bech32MainPrefix + "val" + "cons"
	// Bech32PrefixConsPub defines the Bech32 prefix of a consensus node public key
	Bech32PrefixConsPub = Bech32MainPrefix + "val" + "cons" + "pub"

	Bip44CoinType = 459 // see https://github.com/satoshilabs/slips/blob/master/slip-0044.md
)

// SetSDKConfig configures the global config with kava app specific parameters.
// It does not seal the config to allow modification in tests.
func SetSDKConfig() *sdk.Config {
	config := sdk.GetConfig()
	SetBech32AddressPrefixes(config)
	SetBip44CoinType(config)
	return config
}

// SetBech32AddressPrefixes sets the global prefix to be used when serializing addresses to bech32 strings.
func SetBech32AddressPrefixes(config *sdk.Config) {
	config.SetBech32PrefixForAccount(Bech32MainPrefix, Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(Bech32PrefixValAddr, Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(Bech32PrefixConsAddr, Bech32PrefixConsPub)
}

// SetBip44CoinType sets the global coin type to be used in hierarchical deterministic wallets.
func SetBip44CoinType(config *sdk.Config) {
	config.SetCoinType(Bip44CoinType)
}
