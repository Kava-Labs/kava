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

// TODO: Could store previous reward time accruals under the same key indexed by type ('usd_minting', 'hard_supply', etc.)

// Key Prefixes
var (
	USDXMintingClaimKeyPrefix                       = []byte{0x01} // prefix for keys that store USDX minting claims
	USDXMintingRewardFactorKeyPrefix                = []byte{0x02} // prefix for key that stores USDX minting reward factors
	PreviousUSDXMintingRewardAccrualTimeKeyPrefix   = []byte{0x03} // prefix for key that stores the blocktime
	HardLiquidityClaimKeyPrefix                     = []byte{0x04} // prefix for keys that store Hard liquidity claims
	HardSupplyRewardFactorKeyPrefix                 = []byte{0x05} // prefix for key that stores Hard supply reward factors
	PreviousHardSupplyRewardAccrualTimeKeyPrefix    = []byte{0x06} // prefix for key that stores the previous time Hard supply rewards accrued
	HardBorrowRewardFactorKeyPrefix                 = []byte{0x07} // prefix for key that stores Hard borrow reward factors
	PreviousHardBorrowRewardAccrualTimeKeyPrefix    = []byte{0x08} // prefix for key that stores the previous time Hard borrow rewards accrued
	HardDelegatorRewardFactorKeyPrefix              = []byte{0x09} // prefix for key that stores Hard delegator reward factors
	PreviousHardDelegatorRewardAccrualTimeKeyPrefix = []byte{0x10} // prefix for key that stores the previous time Hard delegator rewards accrued

	USDXMintingRewardDenom = "ukava"
)
