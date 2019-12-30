package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Deposit defines an amount of coins deposited by an account to a cdp
type Deposit struct {
	CdpID         uint64         `json:"cdp_id" yaml:"cdp_id"`       //  cdpID of the cdp
	Depositor     sdk.AccAddress `json:"depositor" yaml:"depositor"` //  Address of the depositor
	Amount        sdk.Coins      `json:"amount" yaml:"amount"`       //  Deposit amount
	InLiquidation bool           `json:"in_liquidation" yaml:"in_liquidation"`
}

// DepositStatus is a type alias that represents a deposit status as a byte
type DepositStatus byte

// Valid Deposit statuses
const (
	StatusNil        DepositStatus = 0x00
	StatusLiquidated DepositStatus = 0x01
)

// AsByte returns the status as byte
func (ds DepositStatus) AsByte() byte {
	return byte(ds)
}

// StatusFromByte returns the status from its byte representation
func StatusFromByte(b byte) DepositStatus {
	switch b {
	case 0x00:
		return StatusNil
	case 0x01:
		return StatusLiquidated
	default:
		panic(fmt.Sprintf("unrecognized deposit status, %v", b))
	}
}

// NewDeposit creates a new Deposit object
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

// Deposits a collection of Deposit objects
type Deposits []Deposit

// String implements fmt.Stringer
func (ds Deposits) String() string {
	if len(ds) == 0 {
		return "[]"
	}
	out := fmt.Sprintf("Deposits for CDP %d:", ds[0].CdpID)
	for _, dep := range ds {
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
