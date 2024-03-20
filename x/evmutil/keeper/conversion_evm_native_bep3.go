package keeper

import (
	"math/big"

	"github.com/kava-labs/kava/x/evmutil/types"
)

var (
	defaultBEP3ConversionDenoms = []string{"bnb", "busd", "btcb", "xrpb"}
)

// IsEvmNativeBep3Conversion returns true if the given pair is a BEP3 conversion pair.
func IsEvmNativeBep3Conversion(pair types.ConversionPair) bool {
	for _, denom := range defaultBEP3ConversionDenoms {
		if pair.Denom == denom {
			return true
		}
	}
	return false
}

// ConvertBep3CoinAmountToERC20Amount converts a bep3 coin amount with 8 decimals
// to the equivalent ERC20 token with 18 decimals.
func ConvertBep3CoinAmountToERC20Amount(amount *big.Int) *big.Int {
	multiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(10), nil)
	result := new(big.Int).Mul(amount, multiplier) // amount * 10^10
	return result
}

// ConvertBep3ERC20AmountToCoinAmount converts a bep3 ERC20 token with 18 decimals
// to the equivalent coin amount with 8 decimals, and returning the remainder.
func ConvertBep3ERC20AmountToCoinAmount(amount *big.Int) (*big.Int, *big.Int) {
	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(10), nil)
	quotient := new(big.Int).Div(amount, divisor)
	remainder := new(big.Int).Mod(amount, divisor)
	return quotient, remainder
}
