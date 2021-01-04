package types

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ValuationMap holds the USD value of various coin types
type ValuationMap struct {
	Usd map[string]sdk.Dec
}

// NewValuationMap returns a new instance of ValuationMap
func NewValuationMap() ValuationMap {
	return ValuationMap{
		Usd: make(map[string]sdk.Dec),
	}
}

// Get returns the USD value for a specific denom
func (m ValuationMap) Get(denom string) sdk.Dec {
	return m.Usd[denom]
}

// SetZero sets the USD value for a specific denom to 0
func (m ValuationMap) SetZero(denom string) {
	m.Usd[denom] = sdk.ZeroDec()
}

// Increment increments the USD value of a denom
func (m ValuationMap) Increment(denom string, amount sdk.Dec) {
	_, ok := m.Usd[denom]
	if !ok {
		m.Usd[denom] = amount
		return
	}
	m.Usd[denom] = m.Usd[denom].Add(amount)
}

// Decrement decrements the USD value of a denom
func (m ValuationMap) Decrement(denom string, amount sdk.Dec) {
	_, ok := m.Usd[denom]
	if !ok {
		m.Usd[denom] = amount
		return
	}
	m.Usd[denom] = m.Usd[denom].Sub(amount)
}

// Sum returns the total USD value of all coins in the map
func (m ValuationMap) Sum() sdk.Dec {
	sum := sdk.ZeroDec()
	for _, v := range m.Usd {
		sum = sum.Add(v)
	}
	return sum
}

// GetSortedKeys returns an array of the map's keys in alphabetical order
func (m ValuationMap) GetSortedKeys() []string {
	keys := make([]string, len(m.Usd))
	i := 0
	for k := range m.Usd {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	return keys
}
