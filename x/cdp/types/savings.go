package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// USDXSavingsRateClaim represents the claim that owner has on usdx savings rate coins
type USDXSavingsRateClaim struct {
	Owner  sdk.AccAddress `json:"owner" yaml:"owner"`
	Factor sdk.Dec        `json:"factor" yaml:"factor"`
}

// NewUSDXSavingsRateClaim returns a new NewUSDXSavingsRateClaim
func NewUSDXSavingsRateClaim(owner sdk.AccAddress, factor sdk.Dec) USDXSavingsRateClaim {
	return USDXSavingsRateClaim{
		Owner:  owner,
		Factor: factor,
	}
}

// Validate performs validation for USDXSavingsRateClaim object
func (uc USDXSavingsRateClaim) Validate() error {
	if uc.Owner.Empty() {
		return fmt.Errorf("usdx savings claim owner should not be empty")
	}
	if uc.Factor.IsNegative() {
		return fmt.Errorf("usdx savings claim factor should not be negative, is %s", uc.Factor)
	}
	return nil
}

// USDXSavingsRateClaims array of USDXSavingsRateClaim
type USDXSavingsRateClaims []USDXSavingsRateClaim

// Validate performs validation for USDXSavingsRateClaims objects
func (ucs USDXSavingsRateClaims) Validate() error {
	for _, uc := range ucs {
		err := uc.Validate()
		if err != nil {
			return err
		}
	}
	return nil
}
