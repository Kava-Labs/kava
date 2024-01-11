package types

const (
	// ModuleName name that will be used throughout the module
	ModuleName = "kavadist"

	// StoreKey Top level store key where all module items will be stored
	StoreKey = ModuleName

	// RouterKey Top level router key
	RouterKey = ModuleName

	// DefaultParamspace default name for parameter store
	DefaultParamspace = ModuleName

	// KavaDistMacc module account for kavadist
	KavaDistMacc = ModuleName

	// Treasury
	FundModuleAccount = "kava-fund"
)

var (
	CurrentDistPeriodKey = []byte{0x00}
	PreviousBlockTimeKey = []byte{0x01}
)
