package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Deposit defines an amount of coins deposited into a hard module account
type Deposit struct {
	Depositor sdk.AccAddress        `json:"depositor" yaml:"depositor"`
	Amount    sdk.Coins             `json:"amount" yaml:"amount"`
	Index     SupplyInterestFactors `json:"index" yaml:"index"`
}

// NewDeposit returns a new deposit
func NewDeposit(depositor sdk.AccAddress, amount sdk.Coins, indexes SupplyInterestFactors) Deposit {
	return Deposit{
		Depositor: depositor,
		Amount:    amount,
		Index:     indexes,
	}
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

// Validate deposit validation
func (d Deposit) Validate() error {
	if d.Depositor.Empty() {
		return fmt.Errorf("Depositor cannot be empty")
	}
	if !d.Amount.IsValid() {
		return fmt.Errorf("Invalid deposit coins: %s", d.Amount)
	}

	if err := d.Index.Validate(); err != nil {
		return err
	}

	return nil
}

func (d Deposit) String() string {
	return fmt.Sprintf(`Deposit:
	Depositor: %s
	Amount: %s
	Index: %s
	`, d.Depositor, d.Amount, d.Index)
}

// Deposits is a slice of Deposit
type Deposits []Deposit

// Validate validates Deposits
func (ds Deposits) Validate() error {
	depositDupMap := make(map[string]Deposit)
	for _, d := range ds {
		if err := d.Validate(); err != nil {
			return err
		}
		dup, ok := depositDupMap[d.Depositor.String()]
		if ok {
			return fmt.Errorf("duplicate depositor: %s\n%s", d, dup)
		}
		depositDupMap[d.Depositor.String()] = d
	}
	return nil
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
