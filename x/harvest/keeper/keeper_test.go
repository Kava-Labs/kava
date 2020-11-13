package keeper_test

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	supplyexported "github.com/cosmos/cosmos-sdk/x/supply/exported"

	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/harvest/keeper"
	"github.com/kava-labs/kava/x/harvest/types"
)

// Test suite used for all keeper tests
type KeeperTestSuite struct {
	suite.Suite

	keeper keeper.Keeper
	app    app.TestApp
	ctx    sdk.Context
	addrs  []sdk.AccAddress
}

// The default state used by each test
func (suite *KeeperTestSuite) SetupTest() {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)

	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
	tApp.InitializeFromGenesisStates()
	_, addrs := app.GeneratePrivKeyAddressPairs(1)
	keeper := tApp.GetHarvestKeeper()
	suite.app = tApp
	suite.ctx = ctx
	suite.keeper = keeper
	suite.addrs = addrs
}

func (suite *KeeperTestSuite) TestGetSetPreviousBlockTime() {
	now := tmtime.Now()

	_, f := suite.keeper.GetPreviousBlockTime(suite.ctx)
	suite.Require().False(f)

	suite.NotPanics(func() { suite.keeper.SetPreviousBlockTime(suite.ctx, now) })

	pbt, f := suite.keeper.GetPreviousBlockTime(suite.ctx)
	suite.True(f)
	suite.Equal(now, pbt)
}

func (suite *KeeperTestSuite) TestGetSetPreviousDelegatorDistribution() {
	now := tmtime.Now()

	_, f := suite.keeper.GetPreviousDelegatorDistribution(suite.ctx, suite.keeper.BondDenom(suite.ctx))
	suite.Require().False(f)

	suite.NotPanics(func() {
		suite.keeper.SetPreviousDelegationDistribution(suite.ctx, now, suite.keeper.BondDenom(suite.ctx))
	})

	pdt, f := suite.keeper.GetPreviousDelegatorDistribution(suite.ctx, suite.keeper.BondDenom(suite.ctx))
	suite.True(f)
	suite.Equal(now, pdt)
}

func (suite *KeeperTestSuite) TestGetSetDeleteDeposit() {
	dep := types.NewDeposit(sdk.AccAddress("test"), sdk.NewCoin("bnb", sdk.NewInt(100)))

	_, f := suite.keeper.GetDeposit(suite.ctx, sdk.AccAddress("test"), "bnb")
	suite.Require().False(f)

	suite.keeper.SetDeposit(suite.ctx, dep)

	testDeposit, f := suite.keeper.GetDeposit(suite.ctx, sdk.AccAddress("test"), "bnb")
	suite.Require().True(f)
	suite.Require().Equal(dep, testDeposit)

	suite.Require().NotPanics(func() { suite.keeper.DeleteDeposit(suite.ctx, dep) })

	_, f = suite.keeper.GetDeposit(suite.ctx, sdk.AccAddress("test"), "bnb")
	suite.Require().False(f)

}

func (suite *KeeperTestSuite) TestIterateDeposits() {
	for i := 0; i < 5; i++ {
		dep := types.NewDeposit(sdk.AccAddress("test"+string(i)), sdk.NewCoin("bnb", sdk.NewInt(100)))
		suite.Require().NotPanics(func() { suite.keeper.SetDeposit(suite.ctx, dep) })
	}
	var deposits []types.Deposit
	suite.keeper.IterateDeposits(suite.ctx, func(d types.Deposit) bool {
		deposits = append(deposits, d)
		return false
	})
	suite.Require().Equal(5, len(deposits))
}

func (suite *KeeperTestSuite) TestIterateDepositsByDenom() {
	for i := 0; i < 5; i++ {
		depA := types.NewDeposit(sdk.AccAddress("test"+string(i)), sdk.NewCoin("bnb", sdk.NewInt(100)))
		suite.Require().NotPanics(func() { suite.keeper.SetDeposit(suite.ctx, depA) })
		depB := types.NewDeposit(sdk.AccAddress("test"+string(i)), sdk.NewCoin("bnb", sdk.NewInt(100)))
		suite.Require().NotPanics(func() { suite.keeper.SetDeposit(suite.ctx, depB) })
		depC := types.NewDeposit(sdk.AccAddress("test"+string(i)), sdk.NewCoin("btcb", sdk.NewInt(100)))
		suite.Require().NotPanics(func() { suite.keeper.SetDeposit(suite.ctx, depC) })
	}

	// Check BNB deposits
	var bnbDeposits []types.Deposit
	suite.keeper.IterateDepositsByDenom(suite.ctx, "bnb", func(d types.Deposit) bool {
		bnbDeposits = append(bnbDeposits, d)
		return false
	})
	suite.Require().Equal(5, len(bnbDeposits))

	// Check BTCB deposits
	var btcbDeposits []types.Deposit
	suite.keeper.IterateDepositsByDenom(suite.ctx, "btcb", func(d types.Deposit) bool {
		btcbDeposits = append(btcbDeposits, d)
		return false
	})
	suite.Require().Equal(5, len(btcbDeposits))

	// Fetch all deposits
	var deposits []types.Deposit
	suite.keeper.IterateDeposits(suite.ctx, func(d types.Deposit) bool {
		deposits = append(deposits, d)
		return false
	})
	suite.Require().Equal(len(bnbDeposits)+len(btcbDeposits), len(deposits))
}

func (suite *KeeperTestSuite) TestGetSetDeleteClaim() {
	claim := types.NewClaim(sdk.AccAddress("test"), "bnb", sdk.NewCoin("hard", sdk.NewInt(100)), "lp")
	_, f := suite.keeper.GetClaim(suite.ctx, sdk.AccAddress("test"), "bnb", "lp")
	suite.Require().False(f)

	suite.Require().NotPanics(func() { suite.keeper.SetClaim(suite.ctx, claim) })
	testClaim, f := suite.keeper.GetClaim(suite.ctx, sdk.AccAddress("test"), "bnb", "lp")
	suite.Require().True(f)
	suite.Require().Equal(claim, testClaim)

	suite.Require().NotPanics(func() { suite.keeper.DeleteClaim(suite.ctx, claim) })
	_, f = suite.keeper.GetClaim(suite.ctx, sdk.AccAddress("test"), "bnb", "lp")
	suite.Require().False(f)
}

func (suite *KeeperTestSuite) TestGetSetDeleteInterestRateModel() {
	denom := "test"
	model := types.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10"))

	_, f := suite.keeper.GetInterestRateModel(suite.ctx, denom)
	suite.Require().False(f)

	suite.keeper.SetInterestRateModel(suite.ctx, denom, model)

	testInterestRateModel, f := suite.keeper.GetInterestRateModel(suite.ctx, denom)
	suite.Require().True(f)
	suite.Require().Equal(model, testInterestRateModel)

	suite.Require().NotPanics(func() { suite.keeper.DeleteInterestRateModel(suite.ctx, denom) })

	_, f = suite.keeper.GetInterestRateModel(suite.ctx, denom)
	suite.Require().False(f)

}

func (suite *KeeperTestSuite) TestIterateInterestRateModels() {
	testDenom := "test"
	var setModels types.InterestRateModels
	var setDenoms []string
	for i := 0; i < 5; i++ {
		denom := testDenom + strconv.Itoa(i)
		model := types.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10"))
		suite.Require().NotPanics(func() { suite.keeper.SetInterestRateModel(suite.ctx, denom, model) })
		// Save the denom and model
		setDenoms = append(setDenoms, denom)
		setModels = append(setModels, model)
	}

	var seenModels types.InterestRateModels
	var seenDenoms []string
	suite.keeper.IterateInterestRateModels(suite.ctx, func(denom string, i types.InterestRateModel) bool {
		seenDenoms = append(seenDenoms, denom)
		seenModels = append(seenModels, i)
		return false
	})

	suite.Require().Equal(setModels, seenModels)
	suite.Require().Equal(setDenoms, seenDenoms)
}

func (suite *KeeperTestSuite) getAccount(addr sdk.AccAddress) authexported.Account {
	ak := suite.app.GetAccountKeeper()
	return ak.GetAccount(suite.ctx, addr)
}

func (suite *KeeperTestSuite) getModuleAccount(name string) supplyexported.ModuleAccountI {
	sk := suite.app.GetSupplyKeeper()
	return sk.GetModuleAccount(suite.ctx, name)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
