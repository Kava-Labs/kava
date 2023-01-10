package types

const (
	// module name
	ModuleName = "community"

	// ModuleAccountName is the name of the module's account
	ModuleAccountName = ModuleName

	// RouterKey is the top-level router key for the module
	RouterKey = ModuleName

	// Query endpoints supported by community
	QueryBalance = "balance"

	// LegacyCommunityPoolModuleName is the module account name used by the legacy community pool
	// It is used to determine the address of the old community pool to be returned with the legacy balance.
	LegacyCommunityPoolModuleName = "distribution"
)
