package types

import (
	"fmt"
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Pair defines a tradable token pair
type Pair struct {
	TokenA    string  `json:"token_a" yaml:"token_a"`
	TokenB    string  `json:"token_b" yaml:"token_b"`
	RewardAPY sdk.Dec `json:"reward_apy" yaml:"reward_apy"`
	SwapFee   sdk.Dec `json:"swap_fee" yaml:"swap_fee"`
}

// NewPair returns a new Pair object
func NewPair(tokenA, tokenB string, rewardAPY, swapFee sdk.Dec) Pair {
	return Pair{
		TokenA:    tokenA,
		TokenB:    tokenB,
		RewardAPY: rewardAPY,
		SwapFee:   swapFee,
	}
}

func (p Pair) Validate() error {
	err := sdk.ValidateDenom(p.TokenA)
	if err != nil {
		return err
	}

	err = sdk.ValidateDenom(p.TokenB)
	if err != nil {
		return err
	}

	if strings.Compare(strings.ToLower(p.TokenA), strings.ToLower(p.TokenB)) == 0 {
		return fmt.Errorf(
			"pair cannot have two tokens of the same type, received '%s' and '%s'",
			strings.ToLower(p.TokenA), strings.ToLower(p.TokenB),
		)
	}

	if p.RewardAPY.IsNil() || p.RewardAPY.IsNegative() {
		return fmt.Errorf(fmt.Sprintf("invalid reward apy: %s:", p.RewardAPY))
	}

	if p.SwapFee.IsNil() || p.SwapFee.IsNegative() || p.SwapFee.GT(sdk.OneDec()) {
		return fmt.Errorf(fmt.Sprintf("invalid swap fee: %s:", p.SwapFee))
	}

	return nil
}

func (p Pair) String() string {
	return fmt.Sprintf(`Pair:
	Token A: %s
	Token B: %s
	Reward APY: %s
	Swap fee: %s
	`, p.TokenA, p.TokenB, p.RewardAPY, p.SwapFee)
}

// Pairs is a slice of Pair
type Pairs []Pair

func (p Pairs) Validate() error {
	pairMap := make(map[string]bool)

	for _, pair := range p {
		// Generate token pair as alphabetically sorted lowercase token names
		tokens := []string{strings.ToLower(pair.TokenA), strings.ToLower(pair.TokenB)}
		sort.Strings(tokens)
		tokenPair := strings.Join(tokens, "/")

		if pairMap[tokenPair] {
			return fmt.Errorf("duplicate pair %s", tokenPair)
		}
		pairMap[tokenPair] = true

		err := p.Validate()
		if err != nil {
			return err
		}
	}
	return nil
}
