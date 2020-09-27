package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
)

// NewPeriod returns a new vesting period
func NewPeriod(amount sdk.Coins, length int64) vesting.Period {
	return vesting.Period{Amount: amount, Length: length}
}

// GetTotalVestingPeriodLength returns the summed length of all vesting periods
func GetTotalVestingPeriodLength(periods vesting.Periods) int64 {
	length := int64(0)
	for _, period := range periods {
		length += period.Length
	}
	return length
}
