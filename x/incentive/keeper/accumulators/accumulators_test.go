package accumulators_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var distantFuture = time.Date(9000, 1, 1, 0, 0, 0, 0, time.UTC)

func i(in int64) sdk.Int                    { return sdk.NewInt(in) }
func d(str string) sdk.Dec                  { return sdk.MustNewDecFromStr(str) }
func c(denom string, amount int64) sdk.Coin { return sdk.NewInt64Coin(denom, amount) }
func dc(denom string, amount string) sdk.DecCoin {
	return sdk.NewDecCoinFromDec(denom, sdk.MustNewDecFromStr(amount))
}
func cs(coins ...sdk.Coin) sdk.Coins        { return sdk.NewCoins(coins...) }
func toDcs(coins ...sdk.Coin) sdk.DecCoins  { return sdk.NewDecCoinsFromCoins(coins...) }
func dcs(coins ...sdk.DecCoin) sdk.DecCoins { return sdk.NewDecCoins(coins...) }
