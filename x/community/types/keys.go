package types

const (
	// module name
	ModuleName = "community"

	// ModuleAccountName is the name of the module's account
	ModuleAccountName = ModuleName

	// StoreKey is the default store key for x/community
	StoreKey = ModuleName

	// QuerierRoute is the querier route for x/community
	QuerierRoute = StoreKey

	// Query endpoints supported by community
	QueryParameters = "parameters"
	QueryBalance    = "balance"
)
