package types

import (
	"bytes"
	"fmt"
	"strings"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MaxSortableDec largest sortable sdk.Dec
var MaxSortableDec = sdk.OneDec().Quo(sdk.SmallestDec())

// ValidSortableDec sdk.Dec can't have precision of less than 10^-18
func ValidSortableDec(dec sdk.Dec) bool {
	return dec.Abs().LTE(MaxSortableDec)
}

// SortableDecBytes returns a byte slice representation of a Dec that can be sorted.
// Left and right pads with 0s so there are 18 digits to left and right of the decimal point.
// For this reason, there is a maximum and minimum value for this, enforced by ValidSortableDec.
func SortableDecBytes(dec sdk.Dec) []byte {
	if !ValidSortableDec(dec) {
		panic("dec must be within bounds")
	}
	// Instead of adding an extra byte to all sortable decs in order to handle max sortable, we just
	// makes its bytes be "max" which comes after all numbers in ASCIIbetical order
	if dec.Equal(MaxSortableDec) {
		return []byte("max")
	}
	// For the same reason, we make the bytes of minimum sortable dec be --, which comes before all numbers.
	if dec.Equal(MaxSortableDec.Neg()) {
		return []byte("--")
	}
	// We move the negative sign to the front of all the left padded 0s, to make negative numbers come before positive numbers
	if dec.IsNegative() {
		return append([]byte("-"), []byte(fmt.Sprintf(fmt.Sprintf("%%0%ds", sdk.Precision*2+1), dec.Abs().String()))...)
	}
	return []byte(fmt.Sprintf(fmt.Sprintf("%%0%ds", sdk.Precision*2+1), dec.String()))
}

// ParseDecBytes parses a []byte encoded using SortableDecBytes back to sdk.Dec
func ParseDecBytes(db []byte) (sdk.Dec, error) {
	strFromDecBytes := strings.Trim(string(db[:]), "0")
	if string(strFromDecBytes[0]) == "." {
		strFromDecBytes = "0" + strFromDecBytes
	}
	if string(strFromDecBytes[len(strFromDecBytes)-1]) == "." {
		strFromDecBytes = strFromDecBytes + "0"
	}
	if bytes.Equal(db, []byte("max")) {
		return MaxSortableDec, nil
	}
	if bytes.Equal(db, []byte("--")) {
		return MaxSortableDec.Neg(), nil
	}
	dec, err := sdk.NewDecFromStr(strFromDecBytes)
	if err != nil {
		return sdk.Dec{}, err
	}
	return dec, nil
}

// RelativePow raises x to the power of n, where x (and the result, z) are scaled by factor b.
// For example, RelativePow(210, 2, 100) = 441 (2.1^2 = 4.41)
// Only defined for positive ints.
func RelativePow(x sdkmath.Int, n sdkmath.Int, b sdkmath.Int) (z sdkmath.Int) {
	if x.IsZero() {
		if n.IsZero() {
			z = b // 0^0 = 1
			return
		}
		z = sdk.ZeroInt() // otherwise 0^a = 0
		return
	}

	z = x
	if n.Mod(sdkmath.NewInt(2)).Equal(sdk.ZeroInt()) {
		z = b
	}

	halfOfB := b.Quo(sdkmath.NewInt(2))
	n = n.Quo(sdkmath.NewInt(2))

	for n.GT(sdk.ZeroInt()) {
		xSquared := x.Mul(x)
		xSquaredRounded := xSquared.Add(halfOfB)

		x = xSquaredRounded.Quo(b)

		if n.Mod(sdkmath.NewInt(2)).Equal(sdk.OneInt()) {
			zx := z.Mul(x)
			zxRounded := zx.Add(halfOfB)
			z = zxRounded.Quo(b)
		}
		n = n.Quo(sdkmath.NewInt(2))
	}
	return
}
