package keeper_test

import (
	"fmt"
	"strconv"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	auctionkeeper "github.com/kava-labs/kava/x/auction/keeper"
	"github.com/kava-labs/kava/x/hard/keeper"
	"github.com/kava-labs/kava/x/hard/types"
)

// Test suite used for all keeper tests
type KeeperTestSuite struct {
	suite.Suite
	keeper        keeper.Keeper
	auctionKeeper auctionkeeper.Keeper
	app           app.TestApp
	ctx           sdk.Context
	addrs         []sdk.AccAddress
}

// The default state used by each test
func (suite *KeeperTestSuite) SetupTest() {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)

	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})
	tApp.InitializeFromGenesisStates()
	_, addrs := app.GeneratePrivKeyAddressPairs(1)
	keeper := tApp.GetHardKeeper()
	suite.app = tApp
	suite.ctx = ctx
	suite.keeper = keeper
	suite.addrs = addrs
}

func (suite *KeeperTestSuite) TestGetSetDeleteDeposit() {
	addr := suite.addrs[0]
	dep := types.NewDeposit(
		addr,
		sdk.NewCoins(sdk.NewCoin("bnb", sdkmath.NewInt(100))),
		types.SupplyInterestFactors{types.NewSupplyInterestFactor("bnb", sdk.MustNewDecFromStr("1.12"))},
	)

	_, f := suite.keeper.GetDeposit(suite.ctx, addr)
	suite.Require().False(f)

	suite.keeper.SetDeposit(suite.ctx, dep)

	storedDeposit, f := suite.keeper.GetDeposit(suite.ctx, addr)
	suite.Require().True(f)
	suite.Require().Equal(dep, storedDeposit)

	suite.Require().NotPanics(func() { suite.keeper.DeleteDeposit(suite.ctx, dep) })

	_, f = suite.keeper.GetDeposit(suite.ctx, addr)
	suite.Require().False(f)
}

func (suite *KeeperTestSuite) TestIterateDeposits() {
	var deposits types.Deposits
	for i := 0; i < 5; i++ {
		dep := types.NewDeposit(
			sdk.AccAddress("test"+fmt.Sprint(i)),
			sdk.NewCoins(sdk.NewCoin("bnb", sdkmath.NewInt(100))),
			types.SupplyInterestFactors{types.NewSupplyInterestFactor("bnb", sdk.MustNewDecFromStr("1.12"))},
		)
		deposits = append(deposits, dep)
		suite.keeper.SetDeposit(suite.ctx, dep)
	}
	var storedDeposits types.Deposits
	suite.keeper.IterateDeposits(suite.ctx, func(d types.Deposit) bool {
		storedDeposits = append(storedDeposits, d)
		return false
	})
	suite.Require().Equal(deposits, storedDeposits)
}

func (suite *KeeperTestSuite) TestGetSetDeleteBorrow() {
	addr := suite.addrs[0]

	borrow := types.NewBorrow(
		addr,
		sdk.NewCoins(sdk.NewInt64Coin("bnb", 1e9)),
		types.BorrowInterestFactors{types.NewBorrowInterestFactor("bnb", sdk.MustNewDecFromStr("1.12"))},
	)

	_, f := suite.keeper.GetBorrow(suite.ctx, addr)
	suite.Require().False(f)

	suite.keeper.SetBorrow(suite.ctx, borrow)

	storedBorrow, f := suite.keeper.GetBorrow(suite.ctx, addr)
	suite.Require().True(f)
	suite.Require().Equal(borrow, storedBorrow)

	suite.Require().NotPanics(func() { suite.keeper.DeleteBorrow(suite.ctx, borrow) })

	_, f = suite.keeper.GetBorrow(suite.ctx, addr)
	suite.Require().False(f)
}

func (suite *KeeperTestSuite) TestIterateBorrows() {
	var borrows types.Borrows
	for i := 0; i < 5; i++ {
		borrow := types.NewBorrow(
			sdk.AccAddress("test"+fmt.Sprint(i)),
			sdk.NewCoins(sdk.NewInt64Coin("bnb", 1e9)),
			types.BorrowInterestFactors{types.NewBorrowInterestFactor("bnb", sdk.MustNewDecFromStr("1.12"))},
		)
		borrows = append(borrows, borrow)
		suite.keeper.SetBorrow(suite.ctx, borrow)
	}
	var storedBorrows types.Borrows
	suite.keeper.IterateBorrows(suite.ctx, func(b types.Borrow) bool {
		storedBorrows = append(storedBorrows, b)
		return false
	})
	suite.Require().Equal(borrows, storedBorrows)
}

func (suite *KeeperTestSuite) TestGetSetDeleteInterestRateModel() {
	denom := "test"
	model := types.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10"))
	borrowLimit := types.NewBorrowLimit(false, sdk.MustNewDecFromStr("0.2"), sdk.MustNewDecFromStr("0.5"))
	moneyMarket := types.NewMoneyMarket(denom, borrowLimit, denom+":usd", sdkmath.NewInt(1000000), model, sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec())

	_, f := suite.keeper.GetMoneyMarket(suite.ctx, denom)
	suite.Require().False(f)

	suite.keeper.SetMoneyMarket(suite.ctx, denom, moneyMarket)

	testMoneyMarket, f := suite.keeper.GetMoneyMarket(suite.ctx, denom)
	suite.Require().True(f)
	suite.Require().Equal(moneyMarket, testMoneyMarket)

	suite.Require().NotPanics(func() { suite.keeper.DeleteMoneyMarket(suite.ctx, denom) })

	_, f = suite.keeper.GetMoneyMarket(suite.ctx, denom)
	suite.Require().False(f)
}

func (suite *KeeperTestSuite) TestIterateInterestRateModels() {
	testDenom := "test"
	var setMMs types.MoneyMarkets
	var setDenoms []string
	for i := 0; i < 5; i++ {
		// Initialize a new money market
		denom := testDenom + strconv.Itoa(i)
		model := types.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10"))
		borrowLimit := types.NewBorrowLimit(false, sdk.MustNewDecFromStr("0.2"), sdk.MustNewDecFromStr("0.5"))
		moneyMarket := types.NewMoneyMarket(denom, borrowLimit, denom+":usd", sdkmath.NewInt(1000000), model, sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec())

		// Store money market in the module's store
		suite.Require().NotPanics(func() { suite.keeper.SetMoneyMarket(suite.ctx, denom, moneyMarket) })

		// Save the denom and model
		setDenoms = append(setDenoms, denom)
		setMMs = append(setMMs, moneyMarket)
	}

	var seenMMs types.MoneyMarkets
	var seenDenoms []string
	suite.keeper.IterateMoneyMarkets(suite.ctx, func(denom string, i types.MoneyMarket) bool {
		seenDenoms = append(seenDenoms, denom)
		seenMMs = append(seenMMs, i)
		return false
	})

	suite.Require().Equal(setMMs, seenMMs)
	suite.Require().Equal(setDenoms, seenDenoms)
}

func (suite *KeeperTestSuite) TestGetSetBorrowedCoins() {
	suite.keeper.SetBorrowedCoins(suite.ctx, sdk.Coins{c("ukava", 123)})

	coins, found := suite.keeper.GetBorrowedCoins(suite.ctx)
	suite.Require().True(found)
	suite.Require().Len(coins, 1)
	suite.Require().Equal(coins, cs(c("ukava", 123)))
}

func (suite *KeeperTestSuite) TestGetSetBorrowedCoins_Empty() {
	coins, found := suite.keeper.GetBorrowedCoins(suite.ctx)
	suite.Require().False(found)
	suite.Require().Empty(coins)

	// None set and setting empty coins should both be the same
	suite.keeper.SetBorrowedCoins(suite.ctx, sdk.Coins{})

	coins, found = suite.keeper.GetBorrowedCoins(suite.ctx)
	suite.Require().False(found)
	suite.Require().Empty(coins)
}

func (suite *KeeperTestSuite) getAccountCoins(acc authtypes.AccountI) sdk.Coins {
	bk := suite.app.GetBankKeeper()
	return bk.GetAllBalances(suite.ctx, acc.GetAddress())
}

func (suite *KeeperTestSuite) getAccount(addr sdk.AccAddress) authtypes.AccountI {
	ak := suite.app.GetAccountKeeper()
	return ak.GetAccount(suite.ctx, addr)
}

func (suite *KeeperTestSuite) getAccountAtCtx(addr sdk.AccAddress, ctx sdk.Context) authtypes.AccountI {
	ak := suite.app.GetAccountKeeper()
	return ak.GetAccount(ctx, addr)
}

func (suite *KeeperTestSuite) getModuleAccount(name string) authtypes.ModuleAccountI {
	ak := suite.app.GetAccountKeeper()
	return ak.GetModuleAccount(suite.ctx, name)
}

func (suite *KeeperTestSuite) getModuleAccountAtCtx(name string, ctx sdk.Context) authtypes.ModuleAccountI {
	ak := suite.app.GetAccountKeeper()
	return ak.GetModuleAccount(ctx, name)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
