package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Pair defines a tradable token pair
type Pair struct {
	TokenA string `json:"token_a" yaml:"token_a"`
	TokenB string `json:"token_b" yaml:"token_b"`
}

// NewPair returns a new Pair object
func NewPair(tokenA, tokenB string) Pair {
	return Pair{
		TokenA: tokenA,
		TokenB: tokenB,
	}
}

// Validate validates pair attributes and returns an error if invalid
func (p Pair) Validate() error {
	err := sdk.ValidateDenom(p.TokenA)
	if err != nil {
		return err
	}

	err = sdk.ValidateDenom(p.TokenB)
	if err != nil {
		return err
	}

	if p.TokenA == p.TokenB {
		return fmt.Errorf(
			"pair cannot have two tokens of the same type, received '%s' and '%s'",
			p.TokenA, p.TokenB,
		)
	}

	return nil
}

// Name returns a unique name for a pair in alphabetical order
func (p Pair) Name() string {
	if p.TokenA < p.TokenB {
		return fmt.Sprintf("%s/%s", p.TokenA, p.TokenB)
	}

	return fmt.Sprintf("%s/%s", p.TokenB, p.TokenA)
}

// String pretty prints the pair
func (p Pair) String() string {
	return fmt.Sprintf(`Pair:
  Name: %s
	Token A: %s
	Token B: %s
`, p.Name(), p.TokenA, p.TokenB)
}

// Pairs is a slice of Pair
type Pairs []Pair

// NewPairs returns Pairs from the provided values
func NewPairs(pairs ...Pair) Pairs {
	return Pairs(pairs)
}

// Validate validates each pair and returns an error if there are any duplicates
func (p Pairs) Validate() error {
	seenPairs := make(map[string]bool)
	for _, pair := range p {
		err := pair.Validate()
		if err != nil {
			return err
		}

		if seen := seenPairs[pair.Name()]; seen {
			return fmt.Errorf("duplicate pair: %s", pair.Name())
		}
		seenPairs[pair.Name()] = true
	}

	return nil
}
