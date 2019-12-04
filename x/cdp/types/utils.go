package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ValidSortableDec sdk.Dec can't have precision of less than 10^-18
func ValidSortableDec(dec sdk.Dec) bool {
	return dec.LTE(sdk.OneDec().Quo(sdk.SmallestDec()))
}

// FormattingStringSortableBytes is the string used in Sprintf to left and right pad the sdk.Dec
// It adjust based on the Precision constant
var FormattingStringSortableBytes = fmt.Sprintf("%%0%ds", sdk.Precision*2+1)

// SortableDecBytes returns a byte slice representation of an sdk.Dec that can be sorted.
// Left and right pads with 0s so there are 18 digits to left and right of decimal point
// For this reason, there is a maximum and minimum value for this,  enforced by ValidSortableDec
func SortableDecBytes(dec sdk.Dec) []byte {
	if !ValidSortableDec(dec) {
		panic("dec must be within bounds")
	}
	return []byte(fmt.Sprintf(FormattingStringSortableBytes, dec.String()))
}