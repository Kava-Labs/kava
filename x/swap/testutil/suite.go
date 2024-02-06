package testutil

import (
	"fmt"
	"reflect"
	"time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/swap/keeper"
	"github.com/kava-labs/kava/x/swap/types"

	sdkmath "cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtime "github.com/cometbft/cometbft/types/time"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	BankKeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/suite"
)

var defaultSwapFee = sdk.MustNewDecFromStr("0.003")

// Suite implements a test suite for the swap module integration tests
type Suite struct {
	suite.Suite
	Keeper        keeper.Keeper
	App           app.TestApp
	Ctx           sdk.Context
	BankKeeper    BankKeeper.Keeper
	AccountKeeper authkeeper.AccountKeeper
}

// SetupTest instantiates a new app, keepers, and sets suite state
func (suite *Suite) SetupTest() {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})

	suite.Ctx = ctx
	suite.App = tApp
	suite.Keeper = tApp.GetSwapKeeper()
	suite.BankKeeper = tApp.GetBankKeeper()
	suite.AccountKeeper = tApp.GetAccountKeeper()
}

// GetEvents returns emitted events on the sdk context
func (suite *Suite) GetEvents() sdk.Events {
	return suite.Ctx.EventManager().Events()
}

// AddCoinsToModule adds coins to the swap module account
func (suite *Suite) AddCoinsToModule(amount sdk.Coins) {
	// Does not use suite.BankKeeper.MintCoins as module account would not have permission to mint
	err := suite.App.FundModuleAccount(suite.Ctx, types.ModuleName, amount)
	suite.Require().NoError(err)
}

// RemoveCoinsFromModule removes coins to the swap module account
func (suite *Suite) RemoveCoinsFromModule(amount sdk.Coins) {
	// Swap module does not have BurnCoins permission so we need to transfer to gov first to burn
	err := suite.BankKeeper.SendCoinsFromModuleToModule(suite.Ctx, types.ModuleAccountName, govtypes.ModuleName, amount)
	suite.Require().NoError(err)
	err = suite.BankKeeper.BurnCoins(suite.Ctx, govtypes.ModuleName, amount)
	suite.Require().NoError(err)
}

// CreateAccount creates a new account from the provided balance
func (suite *Suite) CreateAccount(initialBalance sdk.Coins) authtypes.AccountI {
	_, addrs := app.GeneratePrivKeyAddressPairs(1)
	ak := suite.App.GetAccountKeeper()

	acc := ak.NewAccountWithAddress(suite.Ctx, addrs[0])
	ak.SetAccount(suite.Ctx, acc)

	err := suite.App.FundAccount(suite.Ctx, acc.GetAddress(), initialBalance)
	suite.Require().NoError(err)

	return acc
}

// NewAccountFromAddr creates a new account from the provided address with the provided balance
func (suite *Suite) NewAccountFromAddr(addr sdk.AccAddress, balance sdk.Coins) authtypes.AccountI {
	ak := suite.App.GetAccountKeeper()

	acc := ak.NewAccountWithAddress(suite.Ctx, addr)
	ak.SetAccount(suite.Ctx, acc)

	err := suite.App.FundAccount(suite.Ctx, acc.GetAddress(), balance)
	suite.Require().NoError(err)

	return acc
}

// CreateVestingAccount creates a new vesting account from the provided balance and vesting balance
func (suite *Suite) CreateVestingAccount(initialBalance sdk.Coins, vestingBalance sdk.Coins) authtypes.AccountI {
	acc := suite.CreateAccount(initialBalance)
	bacc := acc.(*authtypes.BaseAccount)

	periods := vestingtypes.Periods{
		vestingtypes.Period{
			Length: 31556952,
			Amount: vestingBalance,
		},
	}
	vacc := vestingtypes.NewPeriodicVestingAccount(bacc, initialBalance, time.Now().Unix(), periods) // TODO is initialBalance correct for originalVesting?

	return vacc
}

// CreatePool creates a pool and stores it in state with the provided reserves
func (suite *Suite) CreatePool(reserves sdk.Coins) error {
	depositor := suite.CreateAccount(reserves)
	pool := types.NewAllowedPool(reserves[0].Denom, reserves[1].Denom)
	suite.Require().NoError(pool.Validate())
	suite.Keeper.SetParams(suite.Ctx, types.NewParams(types.AllowedPools{pool}, defaultSwapFee))

	return suite.Keeper.Deposit(suite.Ctx, depositor.GetAddress(), reserves[0], reserves[1], sdk.MustNewDecFromStr("1"))
}

// AccountBalanceEqual asserts that the coins match the account balance
func (suite *Suite) AccountBalanceEqual(addr sdk.AccAddress, coins sdk.Coins) {
	balance := suite.BankKeeper.GetAllBalances(suite.Ctx, addr)
	suite.Equal(coins, balance, fmt.Sprintf("expected account balance to equal coins %s, but got %s", coins, balance))
}

// ModuleAccountBalanceEqual asserts that the swap module account balance matches the provided coins
func (suite *Suite) ModuleAccountBalanceEqual(coins sdk.Coins) {
	balance := suite.BankKeeper.GetAllBalances(
		suite.Ctx,
		suite.AccountKeeper.GetModuleAddress(types.ModuleAccountName),
	)
	suite.Equal(coins, balance, fmt.Sprintf("expected module account balance to equal coins %s, but got %s", coins, balance))
}

// PoolLiquidityEqual asserts that the pool matching the provided coins has those reserves
func (suite *Suite) PoolLiquidityEqual(coins sdk.Coins) {
	poolRecord, ok := suite.Keeper.GetPool(suite.Ctx, types.PoolIDFromCoins(coins))
	suite.Require().True(ok, "expected pool to exist")
	reserves := sdk.NewCoins(poolRecord.ReservesA, poolRecord.ReservesB)
	suite.Equal(coins, reserves, fmt.Sprintf("expected pool reserves of %s, got %s", coins, reserves))
}

// PoolDeleted asserts that the pool does not exist
func (suite *Suite) PoolDeleted(denomA, denomB string) {
	_, ok := suite.Keeper.GetPool(suite.Ctx, types.PoolID(denomA, denomB))
	suite.Require().False(ok, "expected pool to not exist")
}

// PoolShareTotalEqual asserts the total shares match the stored pool
func (suite *Suite) PoolShareTotalEqual(poolID string, totalShares sdkmath.Int) {
	poolRecord, found := suite.Keeper.GetPool(suite.Ctx, poolID)
	suite.Require().True(found, fmt.Sprintf("expected pool %s to exist", poolID))
	suite.Equal(totalShares, poolRecord.TotalShares, "expected pool total shares to be equal")
}

// PoolDepositorSharesEqual asserts the depositor owns the shares for the provided pool
func (suite *Suite) PoolDepositorSharesEqual(depositor sdk.AccAddress, poolID string, shares sdkmath.Int) {
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
func (suite *Suite) PoolShareValueEqual(depositor authtypes.AccountI, pool types.AllowedPool, coins sdk.Coins) {
	poolRecord, ok := suite.Keeper.GetPool(suite.Ctx, pool.Name())
	suite.Require().True(ok, fmt.Sprintf("expected pool %s to exist", pool.Name()))
	shares, ok := suite.Keeper.GetDepositorShares(suite.Ctx, depositor.GetAddress(), poolRecord.PoolID)
	suite.Require().True(ok, fmt.Sprintf("expected shares to exist for depositor %s", depositor.GetAddress()))

	storedPool, err := types.NewDenominatedPoolWithExistingShares(sdk.NewCoins(poolRecord.ReservesA, poolRecord.ReservesB), poolRecord.TotalShares)
	suite.Nil(err)
	value := storedPool.ShareValue(shares.SharesOwned)
	suite.Equal(coins, value, fmt.Sprintf("expected shares to equal %s, but got %s", coins, value))
}

// PoolSharesDeleted asserts that the pool shares have been removed
func (suite *Suite) PoolSharesDeleted(depositor sdk.AccAddress, denomA, denomB string) {
	_, ok := suite.Keeper.GetDepositorShares(suite.Ctx, depositor, types.PoolID(denomA, denomB))
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

func attrsToMap(attrs []abci.EventAttribute) []sdk.Attribute { // new cosmos changed the event attribute type
	out := []sdk.Attribute{}

	for _, attr := range attrs {
		out = append(out, sdk.NewAttribute(string(attr.Key), string(attr.Value)))
	}

	return out
}
