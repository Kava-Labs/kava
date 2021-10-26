package math

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RelativePow raises x to the power of n, where x (and the result, z) are scaled by factor b.
// For example, RelativePow(210, 2, 100) = 441 (2.1^2 = 4.41)
// Only defined for positive ints.
func RelativePow(x sdk.Int, n sdk.Int, b sdk.Int) (z sdk.Int) {
	if x.IsZero() {
		if n.IsZero() {
			z = b // 0^0 = 1
			return
		}
		z = sdk.ZeroInt() // otherwise 0^a = 0
		return
	}

	z = x
	if n.Mod(sdk.NewInt(2)).Equal(sdk.ZeroInt()) {
		z = b
	}

	halfOfB := b.Quo(sdk.NewInt(2))
	n = n.Quo(sdk.NewInt(2))

	for n.GT(sdk.ZeroInt()) {
		xSquared := x.Mul(x)
		xSquaredRounded := xSquared.Add(halfOfB)

		x = xSquaredRounded.Quo(b)

		if n.Mod(sdk.NewInt(2)).Equal(sdk.OneInt()) {
			zx := z.Mul(x)
			zxRounded := zx.Add(halfOfB)
			z = zxRounded.Quo(b)
		}
		n = n.Quo(sdk.NewInt(2))
	}
	return
}
