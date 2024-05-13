package types

import sdkmath "cosmossdk.io/math"

// SplitBalance represents a full extended balance split into the corresponding
// integer and fractional parts. IntegerAmount is managed by x/bank while the
// fractional part is managed by x/precisebank.
type SplitBalance struct {
	IntegerAmount    sdkmath.Int
	FractionalAmount sdkmath.Int
}

// NewSplitBalance creates a new SplitBalance.
func NewSplitBalance(integerAmount, fractionalAmount sdkmath.Int) SplitBalance {
	return SplitBalance{
		IntegerAmount:    integerAmount,
		FractionalAmount: fractionalAmount,
	}
}

// NewSplitBalanceFromFullAmount creates a new SplitBalance from a full amount.
func NewSplitBalanceFromFullAmount(fullAmount sdkmath.Int) SplitBalance {
	// TODO: Since we may not always need both parts:
	// This can be optimized in the future if we want to prevent
	// unnecessary conversions and allocations, by storing the full amount
	// and only converting & caching when needed in IntegerAmount() and
	// FractionalAmount() methods (fields will be private then).
	return NewSplitBalance(
		fullAmount.Quo(conversionFactor),
		fullAmount.Mod(conversionFactor),
	)
}
