package keeper_test

import (
	"fmt"
	"sort"
	"strconv"
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	supplyexported "github.com/cosmos/cosmos-sdk/x/supply/exported"

	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	aucKeeper "github.com/kava-labs/kava/x/auction/keeper"
	"github.com/kava-labs/kava/x/hard/keeper"
	"github.com/kava-labs/kava/x/hard/types"
	pfKeeper "github.com/kava-labs/kava/x/pricefeed/keeper"
)

// Test suite used for all keeper tests
type KeeperTestSuite struct {
	suite.Suite
	keeper          keeper.Keeper
	auctionKeeper   aucKeeper.Keeper
	pricefeedKeeper pfKeeper.Keeper
	app             app.TestApp
	ctx             sdk.Context
	addrs           []sdk.AccAddress
}

// The default state used by each test
func (suite *KeeperTestSuite) SetupTest() {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)

	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
	tApp.InitializeFromGenesisStates()
	_, addrs := app.GeneratePrivKeyAddressPairs(1)
	keeper := tApp.GetHardKeeper()
	suite.app = tApp
	suite.ctx = ctx
	suite.keeper = keeper
	suite.addrs = addrs
}

func (suite *KeeperTestSuite) TestGetSetDeleteDeposit() {
	dep := types.NewDeposit(sdk.AccAddress("test"), sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(100))),
		types.SupplyInterestFactors{types.NewSupplyInterestFactor("", sdk.MustNewDecFromStr("0"))})

	_, f := suite.keeper.GetDeposit(suite.ctx, sdk.AccAddress("test"))
	suite.Require().False(f)

	suite.keeper.SetDeposit(suite.ctx, dep)

	testDeposit, f := suite.keeper.GetDeposit(suite.ctx, sdk.AccAddress("test"))
	suite.Require().True(f)
	suite.Require().Equal(dep, testDeposit)

	suite.Require().NotPanics(func() { suite.keeper.DeleteDeposit(suite.ctx, dep) })

	_, f = suite.keeper.GetDeposit(suite.ctx, sdk.AccAddress("test"))
	suite.Require().False(f)

}

func (suite *KeeperTestSuite) TestIterateDeposits() {
	for i := 0; i < 5; i++ {
		dep := types.NewDeposit(sdk.AccAddress("test"+fmt.Sprint(i)), sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(100))), types.SupplyInterestFactors{})
		suite.Require().NotPanics(func() { suite.keeper.SetDeposit(suite.ctx, dep) })
	}
	var deposits []types.Deposit
	suite.keeper.IterateDeposits(suite.ctx, func(d types.Deposit) bool {
		deposits = append(deposits, d)
		return false
	})
	suite.Require().Equal(5, len(deposits))
}

func (suite *KeeperTestSuite) TestGetSetDeleteInterestRateModel() {
	denom := "test"
	model := types.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10"))
	borrowLimit := types.NewBorrowLimit(false, sdk.MustNewDecFromStr("0.2"), sdk.MustNewDecFromStr("0.5"))
	moneyMarket := types.NewMoneyMarket(denom, borrowLimit, denom+":usd", sdk.NewInt(1000000), sdk.NewInt(1000000000), model, sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec())

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
		moneyMarket := types.NewMoneyMarket(denom, borrowLimit, denom+":usd", sdk.NewInt(1000000), sdk.NewInt(1000000000), model, sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec())

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

func (suite *KeeperTestSuite) getAccount(addr sdk.AccAddress) authexported.Account {
	ak := suite.app.GetAccountKeeper()
	return ak.GetAccount(suite.ctx, addr)
}

func (suite *KeeperTestSuite) getAccountAtCtx(addr sdk.AccAddress, ctx sdk.Context) authexported.Account {
	ak := suite.app.GetAccountKeeper()
	return ak.GetAccount(ctx, addr)
}

func (suite *KeeperTestSuite) getModuleAccount(name string) supplyexported.ModuleAccountI {
	sk := suite.app.GetSupplyKeeper()
	return sk.GetModuleAccount(suite.ctx, name)
}

func (suite *KeeperTestSuite) getModuleAccountAtCtx(name string, ctx sdk.Context) supplyexported.ModuleAccountI {
	sk := suite.app.GetSupplyKeeper()
	return sk.GetModuleAccount(ctx, name)
}

func addressSort(addrs []sdk.AccAddress) (sortedAddrs []sdk.AccAddress) {
	addrStrs := []string{}
	for _, addr := range addrs {
		addrStrs = append(addrStrs, addr.String())
	}

	sort.Strings(addrStrs)

	for _, addrStr := range addrStrs {
		addr, _ := sdk.AccAddressFromBech32(addrStr)
		sortedAddrs = append(sortedAddrs, addr)
	}
	return sortedAddrs
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
