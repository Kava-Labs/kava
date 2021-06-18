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

// Suite implements a test suite for the swap module integration tests
type Suite struct {
	suite.Suite
	Keeper swap.Keeper
	App    app.TestApp
	Ctx    sdk.Context
}

// SetupTest instantiates a new app, keepers, and sets suite state
func (suite *Suite) SetupTest() {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
	keeper := tApp.GetSwapKeeper()

	suite.Ctx = ctx
	suite.App = tApp
	suite.Keeper = keeper
}

// GetAccount gets an existing account
func (suite *Suite) GetAccount(addr sdk.AccAddress) authexported.Account {
	ak := suite.App.GetAccountKeeper()
	return ak.GetAccount(suite.Ctx, addr)
}

// CreateAccount creates a new account from the provided balance
func (suite *Suite) CreateAccount(initialBalance sdk.Coins) authexported.Account {
	_, addrs := app.GeneratePrivKeyAddressPairs(1)
	ak := suite.App.GetAccountKeeper()

	acc := ak.NewAccountWithAddress(suite.Ctx, addrs[0])
	acc.SetCoins(initialBalance)

	ak.SetAccount(suite.Ctx, acc)
	return acc
}

// CreateVestingAccount creats a new vesting account from the provided balance and vesting balance
func (suite *Suite) CreateVestingAccount(initialBalance sdk.Coins, vestingBalance sdk.Coins) authexported.Account {
	acc := suite.CreateAccount(initialBalance)
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

// CreatePool creates a pool and stores it in state with the provided reserves
func (suite *Suite) CreatePool(reserves sdk.Coins) error {
	depositor := suite.CreateAccount(reserves)
	pool := swap.NewAllowedPool(reserves[0].Denom, reserves[1].Denom)
	suite.Require().NoError(pool.Validate())
	suite.Keeper.SetParams(suite.Ctx, swap.NewParams(swap.NewAllowedPools(pool), swap.DefaultSwapFee))

	return suite.Keeper.Deposit(suite.Ctx, depositor.GetAddress(), reserves[0], reserves[1])
}

// AccountBalanceEqual asserts that the coins match the account balance
func (suite *Suite) AccountBalanceEqual(acc authexported.Account, coins sdk.Coins) {
	ak := suite.App.GetAccountKeeper()
	acc = ak.GetAccount(suite.Ctx, acc.GetAddress())
	suite.Equal(coins, acc.GetCoins(), fmt.Sprintf("expected account balance to equal coins %s, but got %s", coins, acc.GetCoins()))
}

// ModuleAccountBalanceEqual asserts that the swap module account balance matches the provided coins
func (suite *Suite) ModuleAccountBalanceEqual(coins sdk.Coins) {
	sk := suite.App.GetSupplyKeeper()
	macc, _ := sk.GetModuleAccountAndPermissions(suite.Ctx, swap.ModuleName)
	suite.Require().NotNil(macc, "expected module account to be defined")
	suite.Equal(coins, macc.GetCoins(), fmt.Sprintf("expected module account balance to equal coins %s, but got %s", coins, macc.GetCoins()))
}

// PoolLiquidityEqual asserts that the pool matching the provided coins has those reserves
func (suite *Suite) PoolLiquidityEqual(coins sdk.Coins) {
	poolRecord, ok := suite.Keeper.GetPool(suite.Ctx, swap.PoolIDFromCoins(coins))
	suite.Require().True(ok, "expected pool to exist")
	reserves := sdk.NewCoins(poolRecord.ReservesA, poolRecord.ReservesB)
	suite.Equal(coins, reserves, fmt.Sprintf("expected pool reserves of %s, got %s", coins, reserves))
}

// PoolShareValueEqual asserts that the depositor shares are in state and the value matches the expected coins
func (suite *Suite) PoolShareValueEqual(depositor authexported.Account, pool swap.AllowedPool, coins sdk.Coins) {
	poolRecord, ok := suite.Keeper.GetPool(suite.Ctx, pool.Name())
	suite.Require().True(ok, fmt.Sprintf("expected pool %s to exist", pool.Name()))
	shares, ok := suite.Keeper.GetDepositorShares(suite.Ctx, depositor.GetAddress(), poolRecord.PoolID)
	suite.Require().True(ok, fmt.Sprintf("expected shares to exist for depositor %s", depositor.GetAddress()))

	storedPool, err := swap.NewDenominatedPoolWithExistingShares(sdk.NewCoins(poolRecord.ReservesA, poolRecord.ReservesB), poolRecord.TotalShares)
	suite.Nil(err)
	value := storedPool.ShareValue(shares.SharesOwned)
	suite.Equal(coins, value, fmt.Sprintf("expected shares to equal %s, but got %s", coins, value))
}

// EventsContains asserts that the expected event is in the provided events
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
