package testutil

import (
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/kavamint/keeper"
)

type KavamintTestSuite struct {
	suite.Suite

	App           app.TestApp
	Ctx           sdk.Context
	Keeper        keeper.Keeper
	StakingKeeper stakingkeeper.Keeper

	BondDenom string
}

func (suite *KavamintTestSuite) SetupTest() {
	app.SetSDKConfig()
	suite.App = app.NewTestApp()
	suite.App.InitializeFromGenesisStates()
	suite.Ctx = suite.App.BaseApp.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})
	suite.Keeper = suite.App.GetKavamintKeeper()
	suite.StakingKeeper = suite.App.GetStakingKeeper()

	suite.BondDenom = suite.StakingKeeper.BondDenom(suite.Ctx)
}

// SetBondedTokenRatio mints the total supply to an account and creates a validator with a self
// delegation that makes the total staked token ratio set as desired.
// EndBlocker must be run in order for tokens to become bonded.
// returns total supply coins
func (suite *KavamintTestSuite) SetBondedTokenRatio(ratio sdk.Dec) sdk.Coins {
	address := app.RandomAddress()

	supplyAmount := sdk.NewInt(1e10)
	totalSupply := sdk.NewCoins(sdk.NewCoin(suite.BondDenom, supplyAmount))
	amountToBond := ratio.MulInt(supplyAmount).TruncateInt()

	// fund account that will create validator with total supply
	err := suite.App.FundAccount(suite.Ctx, address, totalSupply)
	suite.Require().NoError(err)

	if ratio.IsZero() {
		return totalSupply
	}

	// create a validator with self delegation such that ratio is achieved
	err = suite.App.CreateNewUnbondedValidator(suite.Ctx, sdk.ValAddress(address), amountToBond)
	suite.Require().NoError(err)

	return totalSupply
}
