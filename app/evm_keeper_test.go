package app_test

import (
	"math/big"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/ethereum/go-ethereum/common"
	"github.com/kava-labs/kava/app"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"
	evmkeeper "github.com/tharsis/ethermint/x/evm/keeper"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
)

var conversionMultiplier int64 = 1_000_000_000_000

type evmKeeperTestSuite struct {
	suite.Suite

	App           app.TestApp
	Ctx           sdk.Context
	EVMBankKeeper app.EVMBankKeeper
	EVMKeeper     evmkeeper.Keeper
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
	suite.EVMBankKeeper = app.NewEVMBankKeeper(tApp.GetBankKeeper())
	suite.EVMKeeper = tApp.GetEVMKeeper()
	suite.AccountKeeper = tApp.GetAccountKeeper()

	suite.EVMKeeper.SetParams(ctx, evmtypes.NewParams("ukava", true, true, evmtypes.DefaultChainConfig()))

	_, addrs := app.GeneratePrivKeyAddressPairs(3)
	suite.Addrs = addrs
}

// AddCoinsToAccount adds coins to an account address
func (suite *evmKeeperTestSuite) AddCoinsToAccount(addr sdk.AccAddress, coins sdk.Coins) {
	acc := suite.AccountKeeper.NewAccountWithAddress(suite.Ctx, addr)
	suite.AccountKeeper.SetAccount(suite.Ctx, acc)

	err := suite.App.FundAccount(suite.Ctx, acc.GetAddress(), coins)
	suite.Require().NoError(err, "failed to fund account")
}

func TestEvmKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(evmKeeperTestSuite))
}

func (suite *evmKeeperTestSuite) TestIdempotentConversion() {
	// Ethermint re-uses coins so this is to test that using the same set of
	// coins does not change the value

	// Make a duplicate set of coins to prevent possible references.
	expCoins := sdk.NewCoins(sdk.NewInt64Coin("ukava", 1234_000_000_000_000))
	coins := sdk.NewCoins(sdk.NewInt64Coin("ukava", 1234_000_000_000_000))

	suite.EVMBankKeeper.MintCoins(suite.Ctx, evmtypes.ModuleName, coins)
	suite.Require().Equal(expCoins, coins)

	suite.EVMBankKeeper.MintCoins(suite.Ctx, evmtypes.ModuleName, coins)
	suite.Require().Equal(expCoins, coins)

	// Burn everything! (same qtys)
	suite.EVMBankKeeper.BurnCoins(suite.Ctx, evmtypes.ModuleName, coins)
	suite.Require().Equal(expCoins, coins)

	suite.EVMBankKeeper.BurnCoins(suite.Ctx, evmtypes.ModuleName, coins)
	suite.Require().Equal(expCoins, coins)

	macc := suite.AccountKeeper.GetModuleAccount(suite.Ctx, evmtypes.ModuleName)
	bal := suite.BankKeeper.GetBalance(suite.Ctx, macc.GetAddress(), "ukava")

	suite.Require().Equal(sdk.ZeroInt(), bal.Amount, "evm module account balance should end in 0")
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
			bal := suite.EVMKeeper.GetBalance(suite.Ctx, common.BytesToAddress(tt.giveAddr.Bytes()))

			suite.Require().Equal(tt.wantEvmCoin.Amount.BigInt(), bal)
		})
	}
}

func (suite *evmKeeperTestSuite) TestSetBalance() {
	addr := common.BytesToAddress(suite.Addrs[0].Bytes())

	tests := []struct {
		name              string
		giveEvmBalance    *big.Int
		wantCosmosBalance sdk.Int
		wantPanic         bool
	}{
		{
			"mint to 0ukava",
			big.NewInt(0 * conversionMultiplier),
			sdk.NewInt(0),
			false,
		},
		{
			"mint to 1ukava",
			big.NewInt(1 * conversionMultiplier),
			sdk.NewInt(1),
			false,
		},
		{
			"mint to 500ukava",
			big.NewInt(500 * conversionMultiplier),
			sdk.NewInt(500),
			false,
		},
		{
			"mint to 50000ukava",
			big.NewInt(50000 * conversionMultiplier),
			sdk.NewInt(50000),
			false,
		},
		{
			"burn to 50ukava",
			big.NewInt(50 * conversionMultiplier),
			sdk.NewInt(50),
			false,
		},
		{
			"burn to 0ukava",
			big.NewInt(0 * conversionMultiplier),
			sdk.NewInt(0),
			false,
		},
		{
			"invalid 0.000000000001ukava",
			big.NewInt(1),
			sdk.ZeroInt(),
			true,
		},
		{
			"invalid 0.999999999999ukava",
			big.NewInt(1*conversionMultiplier - 1),
			sdk.ZeroInt(),
			true,
		},
	}

	// These tests also test the previous state, as balance is preserved between
	// each test case. This tests both increasing and decreasing a balance which
	// mints and burns coins.
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			if tt.wantPanic {
				suite.Require().Panics(func() {
					suite.EVMKeeper.SetBalance(suite.Ctx, addr, tt.giveEvmBalance)
				}, "set balance should fail if smaller than 1ukava")

				return
			}

			// SetBalance mints/burns coins based on current and new balance
			err := suite.EVMKeeper.SetBalance(suite.Ctx, addr, tt.giveEvmBalance)
			suite.Require().NoError(err, "set balance should not fail")

			evmBal := suite.EVMKeeper.GetBalance(suite.Ctx, addr)
			suite.Require().Equal(
				tt.giveEvmBalance,
				evmBal,
				"evm balance should be the same as the set balance",
			)

			cosmosBal := suite.BankKeeper.GetBalance(suite.Ctx, suite.Addrs[0], "ukava")
			suite.Require().Equal(
				tt.wantCosmosBalance,
				cosmosBal.Amount,
				"cosmos balance should equal evm balance / 10^12",
			)
		})
	}
}
