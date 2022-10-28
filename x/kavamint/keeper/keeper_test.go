package keeper_test

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

	tApp          app.TestApp
	ctx           sdk.Context
	keeper        keeper.Keeper
	stakingKeeper stakingkeeper.Keeper

	bondDenom string
}

func (suite *KavamintTestSuite) SetupTest() {
	app.SetSDKConfig()
	suite.tApp = app.NewTestApp()
	suite.tApp.InitializeFromGenesisStates()
	suite.ctx = suite.tApp.BaseApp.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})
	suite.keeper = suite.tApp.GetKavamintKeeper()
	suite.stakingKeeper = suite.tApp.GetStakingKeeper()

	suite.bondDenom = suite.keeper.BondDenom(suite.ctx)
}

// SetBondedTokenRatio mints the total supply to an account and creates a validator with a self
// delegation that makes the total staked token ratio set as desired.
// EndBlocker must be run in order for tokens to become bonded.
func (suite *KavamintTestSuite) SetBondedTokenRatio(ratio sdk.Dec) {
	address := app.RandomAddress()

	supplyAmount := sdk.NewInt(1e10)
	totalSupply := sdk.NewCoins(sdk.NewCoin(suite.bondDenom, supplyAmount))
	amountToBond := ratio.MulInt(supplyAmount).TruncateInt()

	// fund account that will create validator with total supply
	err := suite.tApp.FundAccount(suite.ctx, address, totalSupply)
	suite.Require().NoError(err)

	// create a validator with self delegation such that ratio is achieved
	err = suite.tApp.CreateNewUnbondedValidator(suite.ctx, sdk.ValAddress(address), amountToBond)
	suite.Require().NoError(err)
}
