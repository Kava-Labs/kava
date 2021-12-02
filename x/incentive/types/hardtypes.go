package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

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

func (bifs BorrowInterestFactors) String() string {
	out := ""
	for _, bif := range bifs {
		out += bif.String()
	}
	return out
}

// Borrow defines an amount of coins borrowed from a hard module account
type Borrow struct {
	Borrower sdk.AccAddress        `json:"borrower" yaml:"borrower"`
	Amount   sdk.Coins             `json:"amount" yaml:"amount"`
	Index    BorrowInterestFactors `json:"index" yaml:"index"`
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
				coin.Amount.ToDec().Quo(factor),
			),
		)
	}
	return normalized, nil
}

// Deposit defines an amount of coins deposited into a hard module account
type Deposit struct {
	Depositor sdk.AccAddress        `json:"depositor" yaml:"depositor"`
	Amount    sdk.Coins             `json:"amount" yaml:"amount"`
	Index     SupplyInterestFactors `json:"index" yaml:"index"`
}

// NormalizedDeposit is the deposit amounts divided by the interest factors.
//
// Multiplying the normalized deposit by the current global factors gives the current deposit (ie including all interest, ie a synced deposit).
// The normalized deposit is effectively how big the deposit would have been if it had been supplied at time 0 and not touched since.
//
// An error is returned if the deposit is in an invalid state.
func (b Deposit) NormalizedDeposit() (sdk.DecCoins, error) {

	normalized := sdk.NewDecCoins()

	for _, coin := range b.Amount {

		factor, found := b.Index.GetInterestFactor(coin.Denom)
		if !found {
			return nil, fmt.Errorf("deposited amount '%s' missing interest factor", coin.Denom)
		}
		if factor.LT(sdk.OneDec()) {
			return nil, fmt.Errorf("interest factor '%s' < 1", coin.Denom)
		}

		normalized = normalized.Add(
			sdk.NewDecCoinFromDec(
				coin.Denom,
				coin.Amount.ToDec().Quo(factor),
			),
		)
	}
	return normalized, nil
}

// SupplyInterestFactor defines an individual borrow interest factor
type SupplyInterestFactor struct {
	Denom string  `json:"denom" yaml:"denom"`
	Value sdk.Dec `json:"value" yaml:"value"`
}

// NewSupplyInterestFactor returns a new SupplyInterestFactor instance
func NewSupplyInterestFactor(denom string, value sdk.Dec) SupplyInterestFactor {
	return SupplyInterestFactor{
		Denom: denom,
		Value: value,
	}
}

// Validate validates SupplyInterestFactor values
func (sif SupplyInterestFactor) Validate() error {
	if strings.TrimSpace(sif.Denom) == "" {
		return fmt.Errorf("supply interest factor denom cannot be empty")
	}
	if sif.Value.IsNegative() {
		return fmt.Errorf("supply interest factor value cannot be negative: %s", sif)

	}
	return nil
}

func (sif SupplyInterestFactor) String() string {
	return fmt.Sprintf(`[%s,%s]
	`, sif.Denom, sif.Value)
}

// SupplyInterestFactors is a slice of SupplyInterestFactor, because Amino won't marshal maps
type SupplyInterestFactors []SupplyInterestFactor

// GetInterestFactor returns a denom's interest factor value
func (sifs SupplyInterestFactors) GetInterestFactor(denom string) (sdk.Dec, bool) {
	for _, sif := range sifs {
		if sif.Denom == denom {
			return sif.Value, true
		}
	}
	return sdk.ZeroDec(), false
}

// SetInterestFactor sets a denom's interest factor value
func (sifs SupplyInterestFactors) SetInterestFactor(denom string, factor sdk.Dec) SupplyInterestFactors {
	for i, sif := range sifs {
		if sif.Denom == denom {
			sif.Value = factor
			sifs[i] = sif
			return sifs
		}
	}
	return append(sifs, NewSupplyInterestFactor(denom, factor))
}

// RemoveInterestFactor removes a denom's interest factor value
func (sifs SupplyInterestFactors) RemoveInterestFactor(denom string) (SupplyInterestFactors, bool) {
	for i, sif := range sifs {
		if sif.Denom == denom {
			return append(sifs[:i], sifs[i+1:]...), true
		}
	}
	return sifs, false
}

// Validate validates SupplyInterestFactors
func (sifs SupplyInterestFactors) Validate() error {
	for _, sif := range sifs {
		if err := sif.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (sifs SupplyInterestFactors) String() string {
	out := ""
	for _, sif := range sifs {
		out += sif.String()
	}
	return out
}
