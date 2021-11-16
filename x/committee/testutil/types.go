package testutil

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Avoid cluttering test cases with long function names
func I(in int64) sdk.Int                    { return sdk.NewInt(in) }
func D(str string) sdk.Dec                  { return sdk.MustNewDecFromStr(str) }
func C(denom string, amount int64) sdk.Coin { return sdk.NewInt64Coin(denom, amount) }
func Cs(coins ...sdk.Coin) sdk.Coins        { return sdk.NewCoins(coins...) }
