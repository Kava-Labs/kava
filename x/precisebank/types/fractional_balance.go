package types

import (
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

	// Validate the amount with the FractionalAmount wrapper
	return NewFractionalAmountFromInt(fb.Amount).Validate()
}
