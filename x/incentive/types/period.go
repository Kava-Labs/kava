package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
)

// NewPeriod returns a new vesting period
func NewPeriod(amount sdk.Coins, length int64) vesting.Period {
	return vesting.Period{Amount: amount, Length: length}
}
