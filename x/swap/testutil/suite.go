package testutil

import (
	"fmt"
	"time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/swap"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	kv "github.com/tendermint/tendermint/libs/kv"
	tmtime "github.com/tendermint/tendermint/types/time"
)

type Suite struct {
	suite.Suite
	Keeper swap.Keeper
	App    app.TestApp
	Ctx    sdk.Context
}

func (suite *Suite) SetupTest() {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
	keeper := tApp.GetSwapKeeper()

	suite.Ctx = ctx
	suite.App = tApp
	suite.Keeper = keeper
}

func (suite *Suite) GetAccount(initialBalance sdk.Coins) authexported.Account {
	_, addrs := app.GeneratePrivKeyAddressPairs(1)
	ak := suite.App.GetAccountKeeper()

	acc := ak.NewAccountWithAddress(suite.Ctx, addrs[0])
	acc.SetCoins(initialBalance)

	ak.SetAccount(suite.Ctx, acc)
	return acc
}

func (suite *Suite) GetVestingAccount(initialBalance sdk.Coins, vestingBalance sdk.Coins) authexported.Account {
	acc := suite.GetAccount(initialBalance)
	bacc := acc.(*auth.BaseAccount)

	periods := vestingtypes.Periods{
		vestingtypes.Period{
			Length: 31556952,
			Amount: vestingBalance,
		},
	}
	vacc := vestingtypes.NewPeriodicVestingAccount(bacc, time.Now().Unix(), periods)

	return vacc
}

func (suite *Suite) AccountBalanceEqual(acc authexported.Account, coins sdk.Coins) {
	ak := suite.App.GetAccountKeeper()
	acc = ak.GetAccount(suite.Ctx, acc.GetAddress())
	suite.Equal(coins, acc.GetCoins(), fmt.Sprintf("expected account balance to equal coins %s, but got %s", coins, acc.GetCoins()))
}

func (suite *Suite) ModuleAccountBalanceEqual(coins sdk.Coins) {
	sk := suite.App.GetSupplyKeeper()
	macc, _ := sk.GetModuleAccountAndPermissions(suite.Ctx, swap.ModuleName)
	suite.Require().NotNil(macc, "expected module account to be defined")
	suite.Equal(coins, macc.GetCoins(), fmt.Sprintf("expected module account balance to equal coins %s, but got %s", coins, macc.GetCoins()))
}

func (suite *Suite) PoolLiquidityEqual(pool swap.AllowedPool, coins sdk.Coins) {
	storedPool, ok := suite.Keeper.GetPool(suite.Ctx, pool.Name())
	suite.Require().True(ok, "expected pool to exist")
	suite.Equal(coins.AmountOf(pool.TokenA), storedPool.ReservesA.Amount,
		"expected pool reservers of %s%s", coins.AmountOf(pool.TokenA), pool.TokenA,
	)
	suite.Equal(coins.AmountOf(pool.TokenB), storedPool.ReservesB.Amount,
		"expected pool reservers of %s%s", coins.AmountOf(pool.TokenB), pool.TokenB,
	)
}

func (suite *Suite) PoolShareValueEqual(depositor authexported.Account, pool swap.AllowedPool, coins sdk.Coins) {
	storedPool, ok := suite.Keeper.GetPool(suite.Ctx, pool.Name())
	suite.Require().True(ok, fmt.Sprintf("expected pool %s to exist", pool.Name()))
	shares, ok := suite.Keeper.GetDepositorShares(suite.Ctx, depositor.GetAddress(), storedPool.Name())
	suite.Require().True(ok, fmt.Sprintf("expected shares to exist for depositor %s", depositor.GetAddress()))

	value, err := storedPool.ShareValue(shares)
	suite.Nil(err)
	suite.Equal(coins, value, fmt.Sprintf("expected shares to equal %s, but got %s", coins, value))
}

func (suite *Suite) EventsContains(events sdk.Events, expectedEvent sdk.Event) {
	for _, event := range events {
		if event.Type == expectedEvent.Type {
			suite.Equal(attrsToMap(expectedEvent.Attributes), attrsToMap(event.Attributes), fmt.Sprintf("expected event attributes did not match event type %s", event.Type))
			return
		}
	}

	suite.Fail(fmt.Sprintf("event of type %s not found", expectedEvent.Type))
}

func attrsToMap(attrs []kv.Pair) []sdk.Attribute {
	out := []sdk.Attribute{}

	for _, attr := range attrs {
		out = append(out, sdk.NewAttribute(string(attr.Key), string(attr.Value)))
	}

	return out
}
