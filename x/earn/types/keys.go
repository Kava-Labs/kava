package types

import sdk "github.com/cosmos/cosmos-sdk/types"

const (
	// ModuleName name that will be used throughout the module
	ModuleName = "earn"

	// ModuleAccountName name of module account used to hold liquidity
	ModuleAccountName = "earn"

	// StoreKey Top level store key where all module items will be stored
	StoreKey = ModuleName

	// RouterKey Top level router key
	RouterKey = ModuleName

	// QuerierRoute Top level query string
	QuerierRoute = ModuleName

	// DefaultParamspace default name for parameter store
	DefaultParamspace = ModuleName
)

// key prefixes for store
var (
	VaultRecordKeyPrefix = []byte{0x01} // denom -> vault
	VaultSharePrefix     = []byte{0x02}

	sep = []byte("|")
)

// Vault returns a key generated from a vault denom
func VaultKey(denom string) []byte {
	return []byte(denom)
}

// DepositorVaultSharesKey returns a key from a depositor and vault denom
func DepositorVaultSharesKey(depositor sdk.AccAddress, vaultDenom string) []byte {
	return createKey(depositor, sep, []byte(vaultDenom))
}

func createKey(bytes ...[]byte) (r []byte) {
	for _, b := range bytes {
		r = append(r, b...)
	}
	return
}
