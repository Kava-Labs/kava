package types

// MinterKey is the key to use for the keeper store.
var MinterKey = []byte{0x00}

const (
	// module name
	ModuleName = "kavamint"

	// ModuleAccountName is the name of the module's account
	ModuleAccountName = ModuleName

	// StoreKey is the default store key for mint
	StoreKey = ModuleName

	// QuerierRoute is the querier route for the minting store.
	QuerierRoute = StoreKey

	// Query endpoints supported by kavamint
	QueryParameters = "parameters"
	QueryInflation  = "inflation"
)
