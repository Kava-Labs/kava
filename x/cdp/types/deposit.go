package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Deposit defines an amount deposited by an account address to a cdp
type Deposit struct {
	CdpID         uint64         `json:"cdp_id" yaml:"cdp_id"`       //  cdpID of the cdp
	Depositor     sdk.AccAddress `json:"depositor" yaml:"depositor"` //  Address of the depositor
	Amount        sdk.Coins      `json:"amount" yaml:"amount"`       //  Deposit amount
	InLiquidation bool           `json:"in_liquidation" yaml:"in_liquidation"`
}

// NewDeposit creates a new Deposit instance
func NewDeposit(cdpID uint64, depositor sdk.AccAddress, amount sdk.Coins) Deposit {
	return Deposit{cdpID, depositor, amount, false}
}

// String implements fmt.Stringer
func (d Deposit) String() string {
	return fmt.Sprintf(`Deposit for CDP %d:
	  Depositor: %s
		Amount: %s
		In Liquidation: %t`,
		d.CdpID, d.Depositor, d.Amount, d.InLiquidation)
}

// Deposits is a collection of Deposit objects
type Deposits []Deposit

// String implements fmt.Stringer
func (d Deposits) String() string {
	if len(d) == 0 {
		return "[]"
	}
	out := fmt.Sprintf("Deposits for CDP %d:", d[0].CdpID)
	for _, dep := range d {
		out += fmt.Sprintf("\n  %s: %s", dep.Depositor, dep.Amount)
		if dep.InLiquidation {
			out += fmt.Sprintf("(in liquidation)")
		}
	}
	return out
}

// Equals returns whether two deposits are equal.
func (d Deposit) Equals(comp Deposit) bool {
	return d.Depositor.Equals(comp.Depositor) && d.CdpID == comp.CdpID && d.Amount.IsEqual(comp.Amount)
}

// Empty returns whether a deposit is empty.
func (d Deposit) Empty() bool {
	return d.Equals(Deposit{})
}
