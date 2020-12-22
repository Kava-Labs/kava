package types

import (
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

// SupplyInterestFactors is a slice of SupplyInterestFactor, because Amino won't marshal maps
type SupplyInterestFactors []SupplyInterestFactor
