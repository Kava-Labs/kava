package testutil

import (
	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtime "github.com/cometbft/cometbft/types/time"

	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/auction/keeper"
	"github.com/kava-labs/kava/x/auction/types"
)

// Suite implements a test suite for the kavadist module integration tests
type Suite struct {
	suite.Suite

	Keeper        keeper.Keeper
	BankKeeper    bankkeeper.Keeper
	AccountKeeper authkeeper.AccountKeeper
	App           app.TestApp
	Ctx           sdk.Context
	Addrs         []sdk.AccAddress
	ModAcc        *authtypes.ModuleAccount
}

// SetupTest instantiates a new app, keepers, and sets suite state
func (suite *Suite) SetupTest(numAddrs int) {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)
	tApp := app.NewTestApp()

	_, addrs := app.GeneratePrivKeyAddressPairs(numAddrs)

	// Fund liquidator module account
	coins := sdk.NewCoins(
		sdk.NewCoin("token1", sdkmath.NewInt(100)),
		sdk.NewCoin("token2", sdkmath.NewInt(100)),
	)

	ctx := tApp.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})

	modName := "liquidator"
	modBaseAcc := authtypes.NewBaseAccount(authtypes.NewModuleAddress(modName), nil, 0, 0)
	modAcc := authtypes.NewModuleAccount(modBaseAcc, modName, []string{authtypes.Minter, authtypes.Burner}...)
	suite.ModAcc = modAcc

	authGS := app.NewFundedGenStateWithSameCoinsWithModuleAccount(tApp.AppCodec(), coins, addrs, modAcc)

	params := types.NewParams(
		types.DefaultMaxAuctionDuration,
		types.DefaultForwardBidDuration,
		types.DefaultReverseBidDuration,
		types.DefaultIncrement,
		types.DefaultIncrement,
		types.DefaultIncrement,
	)

	auctionGs, err := types.NewGenesisState(types.DefaultNextAuctionID, params, []types.GenesisAuction{})
	suite.Require().NoError(err)

	moduleGs := tApp.AppCodec().MustMarshalJSON(auctionGs)
	gs := app.GenesisState{types.ModuleName: moduleGs}
	tApp.InitializeFromGenesisStates(authGS, gs)

	suite.App = tApp
	suite.Ctx = ctx
	suite.Addrs = addrs
	suite.Keeper = tApp.GetAuctionKeeper()
	suite.BankKeeper = tApp.GetBankKeeper()
	suite.AccountKeeper = tApp.GetAccountKeeper()
}

// AddCoinsToModule adds coins to a named module account
func (suite *Suite) AddCoinsToNamedModule(moduleName string, amount sdk.Coins) {
	// Does not use suite.BankKeeper.MintCoins as module account would not have permission to mint
	err := suite.App.FundModuleAccount(suite.Ctx, moduleName, amount)
	suite.Require().NoError(err)
}

// CheckAccountBalanceEqual asserts that
func (suite *Suite) CheckAccountBalanceEqual(owner sdk.AccAddress, expectedCoins sdk.Coins) {
	balances := suite.BankKeeper.GetAllBalances(suite.Ctx, owner)
	suite.Equal(expectedCoins, balances)
}
