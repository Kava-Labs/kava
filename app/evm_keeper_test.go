package app_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/kava-labs/kava/app"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"
)

var conversionMultiplier int64 = 1_000_000_000_000

type evmKeeperTestSuite struct {
	suite.Suite

	App           app.TestApp
	Ctx           sdk.Context
	EVMKeeper     app.EVMBankKeeper
	BankKeeper    bankkeeper.Keeper
	AccountKeeper authkeeper.AccountKeeper
	Addrs         []sdk.AccAddress
}

func (suite *evmKeeperTestSuite) SetupTest() {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})

	suite.Ctx = ctx
	suite.App = tApp
	suite.BankKeeper = tApp.GetBankKeeper()
	suite.EVMKeeper = app.NewEVMBankKeeper(tApp.GetBankKeeper())
	suite.AccountKeeper = tApp.GetAccountKeeper()

	_, addrs := app.GeneratePrivKeyAddressPairs(3)
	suite.Addrs = addrs
}

// AddCoinsToAccount adds coins to an account address
func (suite *evmKeeperTestSuite) AddCoinsToAccount(addr sdk.AccAddress, coins sdk.Coins) {
	acc := suite.AccountKeeper.NewAccountWithAddress(suite.Ctx, addr)
	suite.AccountKeeper.SetAccount(suite.Ctx, acc)

	err := suite.App.FundAccount(suite.Ctx, acc.GetAddress(), coins)
	suite.Require().NoError(err)
}

func TestEvmKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(evmKeeperTestSuite))
}

func (suite *evmKeeperTestSuite) TestGetBalance() {
	tests := []struct {
		name        string
		giveAddr    sdk.AccAddress
		giveCoin    sdk.Coin
		wantEvmCoin sdk.Coin
	}{
		{
			"0ukava",
			suite.Addrs[0],
			sdk.NewInt64Coin("ukava", 0),
			sdk.NewInt64Coin("ukava", 0),
		},
		{
			"1ukava",
			suite.Addrs[1],
			sdk.NewInt64Coin("ukava", 1),
			sdk.NewInt64Coin("ukava", 1*conversionMultiplier),
		},
		{
			"500ukava",
			suite.Addrs[2],
			sdk.NewInt64Coin("ukava", 500),
			sdk.NewInt64Coin("ukava", 500*conversionMultiplier),
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.AddCoinsToAccount(tt.giveAddr, sdk.NewCoins(tt.giveCoin))
			evmBal := suite.EVMKeeper.GetBalance(suite.Ctx, tt.giveAddr, tt.giveCoin.Denom)

			suite.Require().Equal(tt.wantEvmCoin, evmBal)
		})
	}
}
