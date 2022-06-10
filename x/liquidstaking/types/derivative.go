package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewDerivative returns a new derivative
func NewDerivative(validator sdk.ValAddress, amount sdk.Coin) Derivative {
	return Derivative{
		Validator: validator,
		Amount:    amount,
	}
}

// Validate derivative
func (d Derivative) Validate() error {
	if d.Validator.Empty() {
		return fmt.Errorf("validator cannot be empty")
	}
	if !d.Amount.IsValid() {
		return fmt.Errorf("invalid deposit coin: %s", d.Amount)
	}

	return nil
}

// Derivatives is a slice of Derivative
type Derivatives []Derivative

// Validate validates Derivatives
func (ds Derivatives) Validate() error {
	depositDerMap := make(map[string]Derivative)
	for _, d := range ds {
		if err := d.Validate(); err != nil {
			return err
		}
		dup, ok := depositDerMap[d.Validator.String()]
		if ok {
			return fmt.Errorf("duplicate validator: %s\n%s", d, dup)
		}
		depositDerMap[d.Validator.String()] = d
	}
	return nil
}
