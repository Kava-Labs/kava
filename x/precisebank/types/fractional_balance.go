package types

import (
	fmt "fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// FractionalBalance contains __only__ the fractional balance of an address.
// We want to extend the current KAVA decimal digits from 6 to 18, thus 12 more
// digits are added to the fractional balance. With 12 digits, the maximum
// value of the fractional balance is 1_000_000_000_000 - 1.
// We subtract 1, as 1 more will roll over to the integer balance.
var maxFractionalAmount = sdkmath.NewInt(1_000_000_000_000).SubRaw(1)

// GetMaxFractionalAmount returns the maximum value of a FractionalBalance.
func GetMaxFractionalAmount() sdkmath.Int {
	return maxFractionalAmount
}

// FractionalBalance returns a new FractionalBalance with the given address and
// amount.
func NewFractionalBalance(address string, amount sdkmath.Int) FractionalBalance {
	return FractionalBalance{
		Address: address,
		Amount:  amount,
	}
}

// Validate returns an error if the FractionalBalance has an invalid address or
// negative amount.
func (fb FractionalBalance) Validate() error {
	if _, err := sdk.AccAddressFromBech32(fb.Address); err != nil {
		return err
	}

	if fb.Amount.IsNil() {
		return fmt.Errorf("nil amount")
	}

	if !fb.Amount.IsPositive() {
		return fmt.Errorf("non-positive amount %v", fb.Amount)
	}

	if fb.Amount.GT(maxFractionalAmount) {
		return fmt.Errorf("amount %v exceeds max of %v", fb.Amount, maxFractionalAmount)
	}

	return nil
}
