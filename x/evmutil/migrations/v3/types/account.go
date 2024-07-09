package types

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewAccount(addr sdk.AccAddress, balance sdkmath.Int) *Account {
	return &Account{
		Address: addr,
		Balance: balance,
	}
}

func (b Account) Validate() error {
	if b.Address.Empty() {
		return fmt.Errorf("address cannot be empty")
	}
	if b.Balance.IsNegative() {
		return fmt.Errorf("balance amount cannot be negative; amount: %d", b.Balance)
	}
	return nil
}
