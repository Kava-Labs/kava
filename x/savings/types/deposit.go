package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewDeposit returns a new deposit
func NewDeposit(depositor sdk.AccAddress, amount sdk.Coins) Deposit {
	return Deposit{
		Depositor: depositor,
		Amount:    amount,
	}
}

// Validate deposit validation
func (d Deposit) Validate() error {
	if d.Depositor.Empty() {
		return fmt.Errorf("depositor cannot be empty")
	}
	if !d.Amount.IsValid() {
		return fmt.Errorf("invalid deposit coins: %s", d.Amount)
	}

	return nil
}

// Deposits is a slice of Deposit
type Deposits []Deposit

// Validate validates Deposits
func (ds Deposits) Validate() error {
	depositDupMap := make(map[string]Deposit)
	for _, d := range ds {
		if err := d.Validate(); err != nil {
			return err
		}
		dup, ok := depositDupMap[d.Depositor.String()]
		if ok {
			return fmt.Errorf("duplicate depositor: %s\n%s", d, dup)
		}
		depositDupMap[d.Depositor.String()] = d
	}
	return nil
}
