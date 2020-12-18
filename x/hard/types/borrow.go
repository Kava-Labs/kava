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

// Borrow defines an amount of coins borrowed from a hard module account
type Borrow struct {
	Borrower sdk.AccAddress  `json:"borrower" yaml:"borrower"`
	Amount   sdk.Coins       `json:"amount" yaml:"amount"`
	Index    InterestFactors `json:"index" yaml:"index"`
}

// NewBorrow returns a new Borrow instance
func NewBorrow(borrower sdk.AccAddress, amount sdk.Coins, index InterestFactors) Borrow {
	return Borrow{
		Borrower: borrower,
		Amount:   amount,
		Index:    index,
	}
}

// InterestFactor defines an individual interest factor
type InterestFactor struct {
	Denom string  `json:"denom" yaml:"denom"`
	Value sdk.Dec `json:"value" yaml:"value"`
}

// NewInterestFactor returns a new InterestFactor instance
func NewInterestFactor(denom string, value sdk.Dec) InterestFactor {
	return InterestFactor{
		Denom: denom,
		Value: value,
	}
}

// InterestFactors is a slice of InterestFactor, because Amino won't marshal maps
type InterestFactors []InterestFactor
