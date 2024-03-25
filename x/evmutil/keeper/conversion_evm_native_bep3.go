package keeper

import (
	"math/big"

	errorsmod "cosmossdk.io/errors"
	"github.com/kava-labs/kava/x/evmutil/types"
)

var (
	bep3Denoms = map[string]bool{
		"bnb":  true,
		"busd": true,
		"btcb": true,
		"xrpb": true,
	}

	bep3ConversionFactor = new(big.Int).Exp(big.NewInt(10), big.NewInt(10), nil)
)

func isBep3Asset(denom string) bool {
	return bep3Denoms[denom]
}

// convertBep3CoinAmountToERC20Amount converts a bep3 coin amount with 8 decimals
// to the equivalent ERC20 token with 18 decimals.
func convertBep3CoinAmountToERC20Amount(amount *big.Int) (erc20Amount *big.Int) {
	return new(big.Int).Mul(amount, bep3ConversionFactor)
}

// convertBep3ERC20AmountToCoinAmount converts a bep3 ERC20 token with 18 decimals
// to the equivalent coin amount with 8 decimals, and dropping the remainder.
func convertBep3ERC20AmountToCoinAmount(amount *big.Int) (coinAmount *big.Int) {
	return new(big.Int).Div(amount, bep3ConversionFactor)
}

// bep3ERC20AmountToCoinMintAndERC20LockAmount converts 18 decimals erc20 bep3
// amount to the equivalent 8 decimals coin amount to mint, and the
// 18 decimals erc20 bep3 amount to lock for the converted coin amount.
func bep3ERC20AmountToCoinMintAndERC20LockAmount(amount *big.Int) (
	amountToMint *big.Int, amountToLock *big.Int, err error,
) {
	amountToMint = convertBep3ERC20AmountToCoinAmount(amount)

	// make sure we have at least 1 sdk.Coin to mint
	if amountToMint.Cmp(big.NewInt(0)) == 0 {
		err := errorsmod.Wrapf(
			types.ErrInsufficientConversionAmount,
			"unable to convert bep3 coin due converting less than 1 native unit",
		)
		return nil, nil, err
	}
	amountToLock = convertBep3CoinAmountToERC20Amount(amountToMint)
	return amountToMint, amountToLock, nil
}
