package types

import (
	"errors"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewDeposit creates a new Deposit object
func NewDeposit(cdpID uint64, depositor sdk.AccAddress, amount sdk.Coin) Deposit {
	return Deposit{cdpID, depositor, amount}
}

// Validate performs a basic validation of the deposit fields.
func (d Deposit) Validate() error {
	if d.CdpID == 0 {
		return errors.New("deposit's cdp id cannot be 0")
	}
	if d.Depositor.Empty() {
		return errors.New("depositor cannot be empty")
	}
	if !d.Amount.IsValid() {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidCoins, "deposit %s", d.Amount)
	}
	return nil
}

// Deposits a collection of Deposit objects
type Deposits []Deposit

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
func (ds Deposits) SumCollateral() (sum sdkmath.Int) {
	sum = sdk.ZeroInt()
	for _, d := range ds {
		if !d.Amount.IsZero() {
			sum = sum.Add(d.Amount.Amount)
		}
	}
	return
}
