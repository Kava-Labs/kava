package types

import (
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

// Borrows is a slice of Borrow
type Borrows []Borrow

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

// BorrowInterestFactors is a slice of BorrowInterestFactor, because Amino won't marshal maps
type BorrowInterestFactors []BorrowInterestFactor
