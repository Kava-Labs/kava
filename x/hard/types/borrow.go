package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Borrow defines an amount of coins borrowed from a hard module account
type Borrow struct {
	Borrower sdk.AccAddress        `json:"borrower" yaml:"borrower"`
	Amount   sdk.Coins             `json:"amount" yaml:"amount"`
	Index    BorrowInterestFactors `json:"index" yaml:"index"`
}

// NewBorrow returns a new Borrow instance
func NewBorrow(borrower sdk.AccAddress, amount sdk.Coins, index BorrowInterestFactors) Borrow {
	return Borrow{
		Borrower: borrower,
		Amount:   amount,
		Index:    index,
	}
}

// Validate deposit validation
func (b Borrow) Validate() error {
	if b.Borrower.Empty() {
		return fmt.Errorf("Depositor cannot be empty")
	}
	if !b.Amount.IsValid() {
		return fmt.Errorf("Invalid deposit coins: %s", b.Amount)
	}

	if err := b.Index.Validate(); err != nil {
		return err
	}

	return nil
}

func (b Borrow) String() string {
	return fmt.Sprintf(`Deposit:
	Borrower: %s
	Amount: %s
	Index: %s
	`, b.Borrower, b.Amount, b.Index)
}

// Borrows is a slice of Borrow
type Borrows []Borrow

// Validate validates Borrows
func (bs Borrows) Validate() error {
	borrowDupMap := make(map[string]Borrow)
	for _, b := range bs {
		if err := b.Validate(); err != nil {
			return err
		}
		dup, ok := borrowDupMap[b.Borrower.String()]
		if ok {
			return fmt.Errorf("duplicate borrower: %s\n%s", b, dup)
		}
		borrowDupMap[b.Borrower.String()] = b
	}
	return nil
}

// BorrowInterestFactor defines an individual borrow interest factor
type BorrowInterestFactor struct {
	Denom string  `json:"denom" yaml:"denom"`
	Value sdk.Dec `json:"value" yaml:"value"`
}

// NewBorrowInterestFactor returns a new BorrowInterestFactor instance
func NewBorrowInterestFactor(denom string, value sdk.Dec) BorrowInterestFactor {
	return BorrowInterestFactor{
		Denom: denom,
		Value: value,
	}
}

// Validate validates BorrowInterestFactor values
func (bif BorrowInterestFactor) Validate() error {
	if strings.TrimSpace(bif.Denom) == "" {
		return fmt.Errorf("borrow interest factor denom cannot be empty")
	}
	if bif.Value.IsNegative() {
		return fmt.Errorf("borrow interest factor value cannot be negative: %s", bif)

	}
	return nil
}

func (bif BorrowInterestFactor) String() string {
	return fmt.Sprintf(`[%s,%s]
	`, bif.Denom, bif.Value)
}

// BorrowInterestFactors is a slice of BorrowInterestFactor, because Amino won't marshal maps
type BorrowInterestFactors []BorrowInterestFactor

// GetInterestFactor returns a denom's interest factor value
func (bifs BorrowInterestFactors) GetInterestFactor(denom string) (sdk.Dec, bool) {
	for _, bif := range bifs {
		if bif.Denom == denom {
			return bif.Value, true
		}
	}
	return sdk.ZeroDec(), false
}

// SetInterestFactor sets a denom's interest factor value
func (bifs BorrowInterestFactors) SetInterestFactor(denom string, factor sdk.Dec) BorrowInterestFactors {
	for i, bif := range bifs {
		if bif.Denom == denom {
			bif.Value = factor
			bifs[i] = bif
			return bifs
		}
	}
	return append(bifs, NewBorrowInterestFactor(denom, factor))
}

// Validate validates BorrowInterestFactors
func (bifs BorrowInterestFactors) Validate() error {
	for _, bif := range bifs {
		if err := bif.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (bifs BorrowInterestFactors) String() string {
	out := ""
	for _, bif := range bifs {
		out += bif.String()
	}
	return out
}
