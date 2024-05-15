package types

import (
	fmt "fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	// conversionFactor is used to convert the fractional balance to integer
	// balances.
	conversionFactor = sdkmath.NewInt(1_000_000_000_000)
	// maxFractionalAmount is the largest valid value in a FractionalBalance amount.
	// This is for direct internal use so that there are no extra allocations.
	maxFractionalAmount = conversionFactor.SubRaw(1)
)

// ConversionFactor returns a copy of the conversionFactor used to convert the
// fractional balance to integer balances. This is also 1 greater than the max
// valid fractional amount (999_999_999_999):
// 0 < FractionalBalance < conversionFactor
func ConversionFactor() sdkmath.Int {
	return sdkmath.NewIntFromBigIntMut(conversionFactor.BigInt())
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
