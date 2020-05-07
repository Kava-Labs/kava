package types

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Deposit defines an amount of coins deposited by an account to a cdp
type Deposit struct {
	CdpID     uint64         `json:"cdp_id" yaml:"cdp_id"`       //  cdpID of the cdp
	Depositor sdk.AccAddress `json:"depositor" yaml:"depositor"` //  Address of the depositor
	Amount    sdk.Coin       `json:"amount" yaml:"amount"`       //  Deposit amount
}

// NewDeposit creates a new Deposit object
func NewDeposit(cdpID uint64, depositor sdk.AccAddress, amount sdk.Coin) Deposit {
	return Deposit{cdpID, depositor, amount}
}

// String implements fmt.Stringer
func (d Deposit) String() string {
	return fmt.Sprintf(`Deposit for CDP %d:
	  Depositor: %s
		Amount: %s`,
		d.CdpID, d.Depositor, d.Amount)
}

// Validate performs a basic validation of the deposit fields.
func (d Deposit) Validate() error {
	if d.CdpID == 0 {
		return errors.New("deposit's cdp id cannot be 0")
	}
	if d.Depositor.Empty() {
		return errors.New("depositor cannot be empty")
	}
	if d.Amount.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "deposit %s", d.Amount)
	}
	return nil
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
	}
	return out
}

// Validate validates each deposit
func (ds Deposits) Validate() error {
	for _, d := range ds {
		if err := d.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// Equals returns whether two deposits are equal.
func (d Deposit) Equals(comp Deposit) bool {
	return d.Depositor.Equals(comp.Depositor) && d.CdpID == comp.CdpID && d.Amount.IsEqual(comp.Amount)
}

// Empty returns whether a deposit is empty.
func (d Deposit) Empty() bool {
	return d.Equals(Deposit{})
}

// SumCollateral returns the total amount of collateral in the input deposits
func (ds Deposits) SumCollateral() (sum sdk.Int) {
	sum = sdk.ZeroInt()
	for _, d := range ds {
		if !d.Amount.IsZero() {
			sum = sum.Add(d.Amount.Amount)
		}
	}
	return
}
