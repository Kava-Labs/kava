package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/kava-labs/kava/app"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"
)

// savings the result to a module level variable ensures the compiler doesn't optimize the test away
var coinsResult sdk.Coins

func BenchmarkAccountIteration(b *testing.B) {
	benchmarks := []struct {
		name           string
		numberAccounts int
		coins          bool
	}{
		{name: "10000 Accounts, No Coins", numberAccounts: 10000, coins: false},
		{name: "100000 Accounts, No Coins", numberAccounts: 100000, coins: false},
		{name: "1000000 Accounts, No Coins", numberAccounts: 1000000, coins: false},
		{name: "10000 Accounts, With Coins", numberAccounts: 10000, coins: true},
		{name: "100000 Accounts, With Coins", numberAccounts: 100000, coins: true},
		{name: "1000000 Accounts, With Coins", numberAccounts: 1000000, coins: true},
	}
	coins := sdk.Coins{
		sdk.NewCoin("xrp", sdk.NewInt(1000000000)),
		sdk.NewCoin("usdx", sdk.NewInt(1000000000)),
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			tApp := app.NewTestApp()
			ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
			ak := tApp.GetAccountKeeper()
			tApp.InitializeFromGenesisStates()
			for i := 0; i < bm.numberAccounts; i++ {
				arr := []byte{byte((i & 0xFF0000) >> 16), byte((i & 0xFF00) >> 8), byte(i & 0xFF)}
				addr := sdk.AccAddress(arr)
				acc := ak.NewAccountWithAddress(ctx, addr)
				if bm.coins {
					acc.SetCoins(coins)
				}
				ak.SetAccount(ctx, acc)
			}
			// reset timer ensures we don't count setup time
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ak.IterateAccounts(ctx,
					func(acc exported.Account) (stop bool) {
						coins := acc.GetCoins()
						coinsResult = coins
						return false
					})
			}
		})
	}
}
