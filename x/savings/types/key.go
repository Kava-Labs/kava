package types

const (
	// ModuleName The name that will be used throughout the module
	ModuleName = "savings"

	// StoreKey Top level store key where all module items will be stored
	StoreKey = ModuleName

	// RouterKey Top level router key
	RouterKey = ModuleName

	// QuerierRoute is the querier route for gov
	QuerierRoute = ModuleName

	// DefaultParamspace default namestore
	DefaultParamspace = ModuleName

	// ModuleAccountName is the module account's name
	ModuleAccountName = ModuleName
)

var (
// Add store prefixes here
// IndexPrefix = []byte{0x00}
)
