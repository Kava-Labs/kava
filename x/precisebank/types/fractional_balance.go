package types

import (
	fmt "fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// maxFractionalAmount is the largest valid value in a FractionalBalance amount.
// This is for direct internal use so that there are no extra allocations.
var maxFractionalAmount = sdkmath.NewInt(1_000_000_000_000).SubRaw(1)

// MaxFractionalAmount returns the largest valid value in a FractionalBalance
// amount.
// FractionalBalance contains **only** the fractional balance of an address.
// We want to extend the current KAVA decimal digits from 6 to 18, or in other
// words add 12 fractional digits to ukava.
// With 12 digits, the valid amount is 1 - 999_999_999_999.
func MaxFractionalAmount() sdkmath.Int {
	// BigInt() returns a copy of the internal big.Int, so it's safe to directly
	// use it for a new Int instead of creating another big.Int internally.
	// We need to copy it because the internal value can be accessed and
	// modified via Int.BigIntMut()
	return sdkmath.NewIntFromBigIntMut(maxFractionalAmount.BigInt())
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
