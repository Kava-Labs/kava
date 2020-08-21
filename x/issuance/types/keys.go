package types

const (
	// ModuleName The name that will be used throughout the module
	ModuleName = "issuance"

	// StoreKey Top level store key where all module items will be stored
	StoreKey = ModuleName

	// RouterKey Top level router key
	RouterKey = ModuleName

	// DefaultParamspace default name for parameter store
	DefaultParamspace = ModuleName

	// QuerierRoute route used for abci queries
	QuerierRoute = ModuleName
)

// KVStore key prefixes
var (
	AssetSupplyPrefix    = []byte{0x01}
	PreviousBlockTimeKey = []byte{0x02}
)
