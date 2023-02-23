package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewBorrow returns a new Borrow instance
func NewBorrow(borrower sdk.AccAddress, amount sdk.Coins, index BorrowInterestFactors) Borrow {
	return Borrow{
		Borrower: borrower,
		Amount:   amount,
		Index:    index,
	}
}

// NormalizedBorrow is the borrow amounts divided by the interest factors.
//
// Multiplying the normalized borrow by the current global factors gives the current borrow (ie including all interest, ie a synced borrow).
// The normalized borrow is effectively how big the borrow would have been if it had been borrowed at time 0 and not touched since.
//
// An error is returned if the borrow is in an invalid state.
func (b Borrow) NormalizedBorrow() (sdk.DecCoins, error) {
	normalized := sdk.NewDecCoins()

	for _, coin := range b.Amount {

		factor, found := b.Index.GetInterestFactor(coin.Denom)
		if !found {
			return nil, fmt.Errorf("borrowed amount '%s' missing interest factor", coin.Denom)
		}
		if factor.LT(sdk.OneDec()) {
			return nil, fmt.Errorf("interest factor '%s' < 1", coin.Denom)
		}

		normalized = normalized.Add(
			sdk.NewDecCoinFromDec(
				coin.Denom,
				sdk.NewDecFromInt(coin.Amount).Quo(factor),
			),
		)
	}
	return normalized, nil
}

// Validate deposit validation
func (b Borrow) Validate() error {
	if b.Borrower.Empty() {
		return fmt.Errorf("borrower cannot be empty")
	}
	if !b.Amount.IsValid() {
		return fmt.Errorf("invalid borrow coins: %s", b.Amount)
	}

	if err := b.Index.Validate(); err != nil {
		return err
	}

	return nil
}

// ToResponse converts Borrow to BorrowResponse
func (b Borrow) ToResponse() BorrowResponse {
	return NewBorrowResponse(b.Borrower, b.Amount, b.Index)
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

// ToResponse converts Borrows to BorrowResponses
func (bs Borrows) ToResponse() BorrowResponses {
	var bResponses BorrowResponses

	for _, b := range bs {
		bResponses = append(bResponses, b.ToResponse())
	}
	return bResponses
}

// NewBorrowResponse returns a new BorrowResponse instance
func NewBorrowResponse(borrower sdk.AccAddress, amount sdk.Coins, index BorrowInterestFactors) BorrowResponse {
	return BorrowResponse{
		Borrower: borrower.String(),
		Amount:   amount,
		Index:    index.ToResponse(),
	}
}

// BorrowResponses is a slice of BorrowResponse
type BorrowResponses []BorrowResponse

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

// ToResponse converts BorrowInterestFactor to BorrowInterestFactorResponse
func (bif BorrowInterestFactor) ToResponse() BorrowInterestFactorResponse {
	return NewBorrowInterestFactorResponse(bif.Denom, bif.Value)
}

// NewBorrowInterestFactorResponse returns a new BorrowInterestFactorResponse instance
func NewBorrowInterestFactorResponse(denom string, value sdk.Dec) BorrowInterestFactorResponse {
	return BorrowInterestFactorResponse{
		Denom: denom,
		Value: value.String(),
	}
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

// RemoveInterestFactor removes a denom's interest factor value
func (bifs BorrowInterestFactors) RemoveInterestFactor(denom string) (BorrowInterestFactors, bool) {
	for i, bif := range bifs {
		if bif.Denom == denom {
			return append(bifs[:i], bifs[i+1:]...), true
		}
	}
	return bifs, false
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

// ToResponse converts BorrowInterestFactors to BorrowInterestFactorResponses
func (bifs BorrowInterestFactors) ToResponse() BorrowInterestFactorResponses {
	var bifResponses BorrowInterestFactorResponses

	for _, bif := range bifs {
		bifResponses = append(bifResponses, bif.ToResponse())
	}
	return bifResponses
}

// BorrowInterestFactorResponses is a slice of BorrowInterestFactorResponse
type BorrowInterestFactorResponses []BorrowInterestFactorResponse
