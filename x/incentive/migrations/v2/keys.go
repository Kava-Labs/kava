package v2

// Legacy store key prefixes
var (
	EarnClaimKeyPrefix                     = []byte{0x18} // prefix for keys that store earn claims
	EarnRewardIndexesKeyPrefix             = []byte{0x19} // prefix for key that stores earn reward indexes
	PreviousEarnRewardAccrualTimeKeyPrefix = []byte{0x20} // prefix for key that stores the previous time earn rewards accrued
)
