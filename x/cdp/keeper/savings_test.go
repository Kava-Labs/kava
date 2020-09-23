package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/cdp/keeper"
	"github.com/kava-labs/kava/x/cdp/types"
)

type SavingsTestSuite struct {
	suite.Suite

	keeper             keeper.Keeper
	app                app.TestApp
	ctx                sdk.Context
	addrs              []sdk.AccAddress
	amountToDistribute int64
}

func (suite *SavingsTestSuite) SetupTest() {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})

	_, addrs := app.GeneratePrivKeyAddressPairs(3)
	authGS := app.NewAuthGenState(
		addrs,
		[]sdk.Coins{
			cs(c("usdx", 100000)), cs(c("usdx", 50000)), cs(c("usdx", 50000)),
		},
	)

	tApp.InitializeFromGenesisStates(
		authGS,
		NewPricefeedGenStateMulti(),
		NewCDPGenStateMulti(),
	)

	sk := tApp.GetSupplyKeeper()
	macc := sk.GetModuleAccount(ctx, types.SavingsRateMacc)
	distAmount := int64(10000)
	err := sk.MintCoins(ctx, macc.GetName(), cs(c("usdx", distAmount)))
	suite.NoError(err)

	keeper := tApp.GetCDPKeeper()
	suite.app = tApp
	suite.keeper = keeper
	suite.ctx = ctx
	suite.addrs = addrs
	suite.amountToDistribute = distAmount
}

func (suite *SavingsTestSuite) TestApplySavingsRate() {
	preSavingsRateDistAmount := suite.keeper.GetSavingsRateDistributed(suite.ctx)

	err := suite.keeper.DistributeSavingsRate(suite.ctx, "usdx")
	suite.NoError(err)

	ak := suite.app.GetAccountKeeper()
	acc0 := ak.GetAccount(suite.ctx, suite.addrs[0])
	suite.Equal(cs(c("usdx", 105000)), acc0.GetCoins())
	acc1 := ak.GetAccount(suite.ctx, suite.addrs[1])
	suite.Equal(cs(c("usdx", 52500)), acc1.GetCoins())
	acc2 := ak.GetAccount(suite.ctx, suite.addrs[2])
	suite.Equal(cs(c("usdx", 52500)), acc2.GetCoins())

	sk := suite.app.GetSupplyKeeper()
	macc := sk.GetModuleAccount(suite.ctx, types.SavingsRateMacc)
	suite.True(macc.GetCoins().AmountOf("usdx").IsZero())

	expectedPostSavingsRateDistAmount := preSavingsRateDistAmount.Add(sdk.NewInt(suite.amountToDistribute))
	postSavingsRateDistAmount := suite.keeper.GetSavingsRateDistributed(suite.ctx)
	suite.True(expectedPostSavingsRateDistAmount.Equal(postSavingsRateDistAmount))
}

func (suite *SavingsTestSuite) TestGetSetPreviousDistributionTime() {
	now := tmtime.Now()

	_, f := suite.keeper.GetPreviousSavingsDistribution(suite.ctx)
	suite.Require().False(f) // distr time not set at genesis when the default genesis is used

	suite.NotPanics(func() { suite.keeper.SetPreviousSavingsDistribution(suite.ctx, now) })

	pdt, f := suite.keeper.GetPreviousSavingsDistribution(suite.ctx)
	suite.True(f)
	suite.Equal(now, pdt)
}

func (suite *SavingsTestSuite) TestGetSetSavingsRateDistributed() {
	// Savings rate dist set to 0 when the default genesis is used
	preSavingsRateDistAmount := suite.keeper.GetSavingsRateDistributed(suite.ctx)
	suite.True(preSavingsRateDistAmount.Equal(types.DefaultSavingsRateDistributed))

	// Adding new dist amount to existing dist so default genesis value can be updated in the future
	amountToDistribute := sdk.NewInt(9876543210)
	newTotalDistributed := preSavingsRateDistAmount.Add(amountToDistribute)

	suite.NotPanics(func() { suite.keeper.SetSavingsRateDistributed(suite.ctx, newTotalDistributed) })

	postSavingsRateDistAmount := suite.keeper.GetSavingsRateDistributed(suite.ctx)
	suite.Equal(newTotalDistributed, postSavingsRateDistAmount)
}

func TestSavingsTestSuite(t *testing.T) {
	suite.Run(t, new(SavingsTestSuite))
}
