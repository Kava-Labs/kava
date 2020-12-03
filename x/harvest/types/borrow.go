package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BorrowIndexItem defines an individual borrow index
type BorrowIndexItem struct {
	Denom string  `json:"denom" yaml:"denom"`
	Value sdk.Dec `json:"value" yaml:"value"`
}

// NewBorrowIndexItem returns a new BorrowIndexItem instance
func NewBorrowIndexItem(denom string, value sdk.Dec) BorrowIndexItem {
	return BorrowIndexItem{
		Denom: denom,
		Value: value,
	}
}

// BorrowIndexes is a slice of BorrowIndexItem, because Amino won't marshal maps
type BorrowIndexes []BorrowIndexItem

// Borrow defines an amount of coins borrowed from a harvest module account
type Borrow struct {
	Borrower sdk.AccAddress `json:"borrower" yaml:"borrower"`
	Amount   sdk.Coins      `json:"amount" yaml:"amount"`
	Index    BorrowIndexes  `json:"index" yaml:"index"`
}

// NewBorrow returns a new Borrow instance
func NewBorrow(borrower sdk.AccAddress, amount sdk.Coins, index BorrowIndexes) Borrow {
	return Borrow{
		Borrower: borrower,
		Amount:   amount,
		Index:    index,
	}
}
