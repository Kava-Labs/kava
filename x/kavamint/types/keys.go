package types

const (
	// module name
	ModuleName = "kavamint"

	// ModuleAccountName is the name of the module's account
	ModuleAccountName = ModuleName

	// StoreKey is the default store key for mint
	StoreKey = ModuleName

	// QuerierRoute is the querier route for the minting store.
	QuerierRoute = StoreKey
)

// PreviousBlockTimeKey is the store key for the previous block time
var PreviousBlockTimeKey = []byte{0x00}
