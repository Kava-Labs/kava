package types

const (
	// ModuleName The name that will be used throughout the module
	ModuleName = "incentive"

	// StoreKey Top level store key where all module items will be stored
	StoreKey = ModuleName

	// RouterKey Top level router key
	RouterKey = ModuleName

	// DefaultParamspace default name for parameter store
	DefaultParamspace = ModuleName

	// QuerierRoute route used for abci queries
	QuerierRoute = ModuleName
)

// Key Prefixes
var (
	ClaimKeyPrefix  = []byte{0x01} // prefix for keys that store claims
	BlockTimeKey    = []byte{0x02} // prefix for key that stores the blocktime
	RewardFactorKey = []byte{0x03} // prefix for key that stores reward factors

	USDXMintingRewardDenom = "ukava"
)
