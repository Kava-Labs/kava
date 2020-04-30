package types

import (
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
)

// GetTotalVestingPeriodLength returns the summed length of all vesting periods
func GetTotalVestingPeriodLength(periods vesting.Periods) int64 {
	length := int64(0)
	for _, period := range periods {
		length += period.Length
	}
	return length
}
