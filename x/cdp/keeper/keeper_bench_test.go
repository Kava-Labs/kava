package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/cdp"
	"github.com/kava-labs/kava/x/cdp/keeper"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"
)

// saving the result to a module level variable ensures the compiler doesn't optimize the test away
var coinsResult sdk.Coins
var coinResult sdk.Coin

// Note - the iteration benchmarks take a long time to stabilize, to get stable results use:
// go test -benchmem -bench ^(BenchmarkAccountIteration)$ -benchtime 60s  -timeout 2h
// go test -benchmem -bench ^(BenchmarkCdpIteration)$ -benchtime 60s  -timeout 2h

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

func createCdps(n int) (app.TestApp, sdk.Context, keeper.Keeper) {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
	_, addrs := app.GeneratePrivKeyAddressPairs(n)
	coins := []sdk.Coins{}
	for i := 0; i < n; i++ {
		coins = append(coins, cs(c("btc", 100000000)))
	}
	authGS := app.NewAuthGenState(
		addrs, coins)
	tApp.InitializeFromGenesisStates(
		authGS,
		NewPricefeedGenStateMulti(),
		NewCDPGenStateMulti(),
	)
	cdpKeeper := tApp.GetCDPKeeper()
	for i := 0; i < n; i++ {
		err := cdpKeeper.AddCdp(ctx, addrs[i], coins[i][0], c("usdx", 100000000))
		if err != nil {
			panic("failed to create cdp")
		}
	}
	return tApp, ctx, cdpKeeper
}

func BenchmarkCdpIteration(b *testing.B) {
	benchmarks := []struct {
		name       string
		numberCdps int
	}{
		{"1000 Cdps", 1000},
		{"10000 Cdps", 10000},
		{"100000 Cdps", 100000},
	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			_, ctx, cdpKeeper := createCdps(bm.numberCdps)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				cdpKeeper.IterateAllCdps(ctx, func(c cdp.CDP) (stop bool) {
					coinResult = c.Principal
					return false
				})
			}
		})
	}

}

var errResult error

func BenchmarkCdpCreation(b *testing.B) {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
	_, addrs := app.GeneratePrivKeyAddressPairs(b.N)
	coins := []sdk.Coins{}
	for i := 0; i < b.N; i++ {
		coins = append(coins, cs(c("btc", 100000000)))
	}
	authGS := app.NewAuthGenState(
		addrs, coins)
	tApp.InitializeFromGenesisStates(
		authGS,
		NewPricefeedGenStateMulti(),
		NewCDPGenStateMulti(),
	)
	cdpKeeper := tApp.GetCDPKeeper()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := cdpKeeper.AddCdp(ctx, addrs[i], coins[i][0], c("usdx", 100000000))
		if err != nil {
			b.Error("unexpected error")
		}
		errResult = err
	}
}
