package testutil

import (
	"fmt"
	"reflect"
	"time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/swap"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/supply"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	kv "github.com/tendermint/tendermint/libs/kv"
	tmtime "github.com/tendermint/tendermint/types/time"
)

var DefaultSwapFee = sdk.MustNewDecFromStr("0.003")

// Suite implements a test suite for the swap module integration tests
type Suite struct {
	suite.Suite
	Keeper       swap.Keeper
	App          app.TestApp
	Ctx          sdk.Context
	bankKeeper   bank.Keeper
	supplyKeeper supply.Keeper
}

// SetupTest instantiates a new app, keepers, and sets suite state
func (suite *Suite) SetupTest() {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
	keeper := tApp.GetSwapKeeper()
	bankKeeper := tApp.GetBankKeeper()
	supplyKeeper := tApp.GetSupplyKeeper()

	suite.Ctx = ctx
	suite.App = tApp
	suite.Keeper = keeper
	suite.bankKeeper = bankKeeper
	suite.supplyKeeper = supplyKeeper
}

// AddCoinsToModule adds coins to the swap module account
func (suite *Suite) AddCoinsToModule(amount sdk.Coins) {
	macc, _ := suite.supplyKeeper.GetModuleAccountAndPermissions(suite.Ctx, swap.ModuleName)
	_, err := suite.bankKeeper.AddCoins(suite.Ctx, macc.GetAddress(), amount)
	suite.Require().NoError(err)
}

// RemoveCoinsFromModule adds coins to the swap module account
func (suite *Suite) RemoveCoinsFromModule(amount sdk.Coins) {
	macc, _ := suite.supplyKeeper.GetModuleAccountAndPermissions(suite.Ctx, swap.ModuleName)
	_, err := suite.bankKeeper.SubtractCoins(suite.Ctx, macc.GetAddress(), amount)
	suite.Require().NoError(err)
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
	err := acc.SetCoins(initialBalance)
	suite.Require().NoError(err)

	ak.SetAccount(suite.Ctx, acc)
	return acc
}

// NewAccountFromAddr creates a new account from the provided address with the provided balance
func (suite *Suite) NewAccountFromAddr(addr sdk.AccAddress, balance sdk.Coins) authexported.Account {
	ak := suite.App.GetAccountKeeper()

	acc := ak.NewAccountWithAddress(suite.Ctx, addr)
	err := acc.SetCoins(balance)
	suite.Require().NoError(err)

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
	suite.Keeper.SetParams(suite.Ctx, swap.NewParams(swap.NewAllowedPools(pool), DefaultSwapFee))

	return suite.Keeper.Deposit(suite.Ctx, depositor.GetAddress(), reserves[0], reserves[1], sdk.MustNewDecFromStr("1"))
}

// AccountBalanceEqual asserts that the coins match the account balance
func (suite *Suite) AccountBalanceEqual(acc authexported.Account, coins sdk.Coins) {
	ak := suite.App.GetAccountKeeper()
	acc = ak.GetAccount(suite.Ctx, acc.GetAddress())
	suite.Equal(coins, acc.GetCoins(), fmt.Sprintf("expected account balance to equal coins %s, but got %s", coins, acc.GetCoins()))
}

// AccountBalanceDelta asserts that the coins are within delta of the account balance
func (suite *Suite) AccountBalanceDelta(acc authexported.Account, coins sdk.Coins, delta float64) {
	ak := suite.App.GetAccountKeeper()
	acc = ak.GetAccount(suite.Ctx, acc.GetAddress())
	accCoins := acc.GetCoins()
	allCoins := coins.Add(accCoins...)
	for _, coin := range allCoins {
		suite.InDelta(
			coins.AmountOf(coin.Denom).Int64(),
			accCoins.AmountOf(coin.Denom).Int64(),
			delta,
			fmt.Sprintf("expected module account balance to be in delta %f of coins %s, but got %s", delta, coins, accCoins),
		)
	}
}

// ModuleAccountBalanceEqual asserts that the swap module account balance matches the provided coins
func (suite *Suite) ModuleAccountBalanceEqual(coins sdk.Coins) {
	macc, _ := suite.supplyKeeper.GetModuleAccountAndPermissions(suite.Ctx, swap.ModuleName)
	suite.Require().NotNil(macc, "expected module account to be defined")
	suite.Equal(coins, macc.GetCoins(), fmt.Sprintf("expected module account balance to equal coins %s, but got %s", coins, macc.GetCoins()))
}

// ModuleAccountBalanceDelta asserts that the swap module account balance is within acceptable delta of the provided coins
func (suite *Suite) ModuleAccountBalanceDelta(coins sdk.Coins, delta float64) {
	macc, _ := suite.supplyKeeper.GetModuleAccountAndPermissions(suite.Ctx, swap.ModuleName)
	suite.Require().NotNil(macc, "expected module account to be defined")

	allCoins := coins.Add(macc.GetCoins()...)
	for _, coin := range allCoins {
		suite.InDelta(
			coins.AmountOf(coin.Denom).Int64(),
			macc.GetCoins().AmountOf(coin.Denom).Int64(),
			delta,
			fmt.Sprintf("expected module account balance to be in delta %f of coins %s, but got %s", delta, coins, macc.GetCoins()),
		)
	}
}

// PoolLiquidityEqual asserts that the pool matching the provided coins has those reserves
func (suite *Suite) PoolLiquidityEqual(coins sdk.Coins) {
	poolRecord, ok := suite.Keeper.GetPool(suite.Ctx, swap.PoolIDFromCoins(coins))
	suite.Require().True(ok, "expected pool to exist")
	reserves := sdk.NewCoins(poolRecord.ReservesA, poolRecord.ReservesB)
	suite.Equal(coins, reserves, fmt.Sprintf("expected pool reserves of %s, got %s", coins, reserves))
}

// PoolDeleted asserts that the pool does not exist
func (suite *Suite) PoolDeleted(denomA, denomB string) {
	_, ok := suite.Keeper.GetPool(suite.Ctx, swap.PoolID(denomA, denomB))
	suite.Require().False(ok, "expected pool to not exist")
}

// PoolLiquidityDelta asserts that the pool matching the provided coins has those reserves within delta
func (suite *Suite) PoolLiquidityDelta(coins sdk.Coins, delta float64) {
	poolRecord, ok := suite.Keeper.GetPool(suite.Ctx, swap.PoolIDFromCoins(coins))
	suite.Require().True(ok, "expected pool to exist")

	suite.InDelta(
		poolRecord.ReservesA.Amount.Int64(),
		coins.AmountOf(poolRecord.ReservesA.Denom).Int64(),
		delta,
		fmt.Sprintf("expected pool reserves within delta %f of %s, got %s", delta, coins, poolRecord.Reserves()),
	)
	suite.InDelta(
		poolRecord.ReservesB.Amount.Int64(),
		coins.AmountOf(poolRecord.ReservesB.Denom).Int64(),
		delta,
		fmt.Sprintf("expected pool reserves within delta %f of %s, got %s", delta, coins, poolRecord.Reserves()),
	)
}

// PoolShareTotalEqual asserts the total shares match the stored pool
func (suite *Suite) PoolShareTotalEqual(poolID string, totalShares sdk.Int) {
	poolRecord, found := suite.Keeper.GetPool(suite.Ctx, poolID)
	suite.Require().True(found, fmt.Sprintf("expected pool %s to exist", poolID))
	suite.Equal(totalShares, poolRecord.TotalShares, "expected pool total shares to be equal")
}

// PoolDepositorSharesEqual asserts the depositor owns the shares for the provided pool
func (suite *Suite) PoolDepositorSharesEqual(depositor sdk.AccAddress, poolID string, shares sdk.Int) {
	shareRecord, found := suite.Keeper.GetDepositorShares(suite.Ctx, depositor, poolID)
	suite.Require().True(found, fmt.Sprintf("expected share record to exist for depositor %s and pool %s", depositor.String(), poolID))
	suite.Equal(shares, shareRecord.SharesOwned)
}

// PoolReservesEqual assets the stored pool reserves are equal to the provided reserves
func (suite *Suite) PoolReservesEqual(poolID string, reserves sdk.Coins) {
	poolRecord, found := suite.Keeper.GetPool(suite.Ctx, poolID)
	suite.Require().True(found, fmt.Sprintf("expected pool %s to exist", poolID))
	suite.Equal(reserves, poolRecord.Reserves(), "expected pool reserves to be equal")
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

// PoolShareValueDelta asserts that the depositor shares are in state and the value is within delta of the expected coins
func (suite *Suite) PoolShareValueDelta(depositor authexported.Account, pool swap.AllowedPool, coins sdk.Coins, delta float64) {
	poolRecord, ok := suite.Keeper.GetPool(suite.Ctx, pool.Name())
	suite.Require().True(ok, fmt.Sprintf("expected pool %s to exist", pool.Name()))
	shares, ok := suite.Keeper.GetDepositorShares(suite.Ctx, depositor.GetAddress(), poolRecord.PoolID)
	suite.Require().True(ok, fmt.Sprintf("expected shares to exist for depositor %s", depositor.GetAddress()))

	storedPool, err := swap.NewDenominatedPoolWithExistingShares(sdk.NewCoins(poolRecord.ReservesA, poolRecord.ReservesB), poolRecord.TotalShares)
	suite.Nil(err)
	value := storedPool.ShareValue(shares.SharesOwned)

	for _, coin := range coins {
		suite.InDelta(
			coin.Amount.Int64(),
			value.AmountOf(coin.Denom).Int64(),
			delta,
			fmt.Sprintf("expected shares to be within delta %f of %s, but got %s", delta, coins, value),
		)
	}
}

// PoolSharesDeleted asserts that the pool shares have been removed
func (suite *Suite) PoolSharesDeleted(depositor sdk.AccAddress, denomA, denomB string) {
	_, ok := suite.Keeper.GetDepositorShares(suite.Ctx, depositor, swap.PoolID(denomA, denomB))
	suite.Require().False(ok, "expected pool shares to not exist")
}

// EventsContains asserts that the expected event is in the provided events
func (suite *Suite) EventsContains(events sdk.Events, expectedEvent sdk.Event) {
	foundMatch := false
	for _, event := range events {
		if event.Type == expectedEvent.Type {
			if reflect.DeepEqual(attrsToMap(expectedEvent.Attributes), attrsToMap(event.Attributes)) {
				foundMatch = true
			}
		}
	}

	suite.True(foundMatch, fmt.Sprintf("event of type %s not found or did not match", expectedEvent.Type))
}

func attrsToMap(attrs []kv.Pair) []sdk.Attribute {
	out := []sdk.Attribute{}

	for _, attr := range attrs {
		out = append(out, sdk.NewAttribute(string(attr.Key), string(attr.Value)))
	}

	return out
}
