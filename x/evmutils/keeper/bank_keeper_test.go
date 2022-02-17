package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/evmutils/keeper"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"

	"github.com/kava-labs/kava/x/evmutils/types"
)

type evmKeeperTestSuite struct {
	suite.Suite

	app           app.TestApp
	ctx           sdk.Context
	bk            types.BankKeeper
	ak            authkeeper.AccountKeeper
	evmBankKeeper keeper.EvmBankKeeper
	addrs         []sdk.AccAddress
	evmModuleAddr sdk.AccAddress
}

func (suite *evmKeeperTestSuite) SetupTest() {
	tApp := app.NewTestApp()

	suite.ctx = tApp.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})
	suite.app = tApp
	suite.bk = tApp.GetBankKeeper()
	suite.ak = tApp.GetAccountKeeper()
	suite.evmBankKeeper = keeper.NewEvmBankKeeper(tApp.GetBankKeeper())
	suite.evmModuleAddr = suite.ak.GetModuleAddress(evmtypes.ModuleName)

	_, addrs := app.GeneratePrivKeyAddressPairs(4)
	suite.addrs = addrs
}

func (suite *evmKeeperTestSuite) TestBurnCoins() {
	startingUkava := sdk.NewInt(100)
	tests := []struct {
		name       string
		burnCoins  sdk.Coins
		ukava      sdk.Int
		akava      sdk.Int
		hasErr     bool
		akavaStart sdk.Int
	}{
		{
			"burn more than 1 ukava",
			sdk.NewCoins(sdk.NewInt64Coin("akava", 12_021_000_000_002)),
			sdk.NewInt(88),
			sdk.NewInt(100_000_000_000),
			false,
			sdk.NewInt(121_000_000_002),
		},
		{
			"burn less than 1 ukava",
			sdk.NewCoins(sdk.NewInt64Coin("akava", 122)),
			sdk.NewInt(100),
			sdk.NewInt(878),
			false,
			sdk.NewInt(1000),
		},
		{
			"burn an exact amount of ukava",
			sdk.NewCoins(sdk.NewInt64Coin("akava", 98_000_000_000_000)),
			sdk.NewInt(2),
			sdk.NewInt(10),
			false,
			sdk.NewInt(10),
		},
		{
			"burn no akava",
			sdk.NewCoins(sdk.NewInt64Coin("akava", 0)),
			startingUkava,
			sdk.ZeroInt(),
			false,
			sdk.ZeroInt(),
		},
		{
			"errors if burning other coins",
			sdk.NewCoins(sdk.NewInt64Coin("akava", 500), sdk.NewInt64Coin("busd", 1000)),
			startingUkava,
			sdk.NewInt(100),
			true,
			sdk.NewInt(100),
		},
		{
			"errors if have dup coins",
			sdk.Coins{
				sdk.NewInt64Coin("akava", 12_000_000_000_000),
				sdk.NewInt64Coin("akava", 2_000_000_000_000),
			},
			startingUkava,
			sdk.ZeroInt(),
			true,
			sdk.ZeroInt(),
		},
		{
			"errors if burn amount is negative",
			sdk.Coins{sdk.Coin{Denom: "akava", Amount: sdk.NewInt(-100)}},
			startingUkava,
			sdk.NewInt(50),
			true,
			sdk.NewInt(50),
		},
		{
			"errors if not enough akava to cover burn",
			sdk.NewCoins(sdk.NewInt64Coin("akava", 10_999_000_000_000)),
			sdk.NewInt(100),
			sdk.NewInt(99_000_000_000),
			true,
			sdk.NewInt(99_000_000_000),
		},
		{
			"errors if not enough ukava to cover burn",
			sdk.NewCoins(sdk.NewInt64Coin("akava", 200_000_000_000_000)),
			sdk.NewInt(100),
			sdk.ZeroInt(),
			true,
			sdk.ZeroInt(),
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.SetupTest()

			startingCoins := sdk.NewCoins(
				sdk.NewCoin("ukava", startingUkava),
				sdk.NewCoin("akava", tt.akavaStart),
			)
			err := suite.bk.MintCoins(suite.ctx, evmtypes.ModuleName, startingCoins)
			suite.Require().NoError(err)

			err = suite.evmBankKeeper.BurnCoins(suite.ctx, evmtypes.ModuleName, tt.burnCoins)
			if tt.hasErr {
				suite.Require().Error(err)
				return
			} else {
				suite.Require().NoError(err)
			}

			// check ukava
			ukavaActual := suite.bk.GetBalance(suite.ctx, suite.evmModuleAddr, "ukava")
			suite.Require().Equal(tt.ukava, ukavaActual.Amount)

			// check akava
			akavaActual := suite.bk.GetBalance(suite.ctx, suite.evmModuleAddr, "akava")
			suite.Require().Equal(tt.akava, akavaActual.Amount)
		})
	}
}

func (suite *evmKeeperTestSuite) TestMintCoins() {
	tests := []struct {
		name       string
		mintCoins  sdk.Coins
		ukava      sdk.Int
		akava      sdk.Int
		hasErr     bool
		akavaStart sdk.Int
	}{
		{
			"mint more than 1 ukava",
			sdk.NewCoins(sdk.NewInt64Coin("akava", 12_021_000_000_002)),
			sdk.NewInt(12),
			sdk.NewInt(21_000_000_002),
			false,
			sdk.ZeroInt(),
		},
		{
			"mint less than 1 ukava",
			sdk.NewCoins(sdk.NewInt64Coin("akava", 901_000_000_001)),
			sdk.ZeroInt(),
			sdk.NewInt(901_000_000_001),
			false,
			sdk.ZeroInt(),
		},
		{
			"mint an exact amount of ukava",
			sdk.NewCoins(sdk.NewInt64Coin("akava", 123_000_000_000_000_000)),
			sdk.NewInt(123_000),
			sdk.ZeroInt(),
			false,
			sdk.ZeroInt(),
		},
		{
			"mint no akava",
			sdk.NewCoins(sdk.NewInt64Coin("akava", 0)),
			sdk.ZeroInt(),
			sdk.ZeroInt(),
			false,
			sdk.ZeroInt(),
		},
		{
			"errors if minting other coins",
			sdk.NewCoins(sdk.NewInt64Coin("akava", 500), sdk.NewInt64Coin("busd", 1000)),
			sdk.ZeroInt(),
			sdk.NewInt(100),
			true,
			sdk.NewInt(100),
		},
		{
			"errors if have dup coins",
			sdk.Coins{
				sdk.NewInt64Coin("akava", 12_000_000_000_000),
				sdk.NewInt64Coin("akava", 2_000_000_000_000),
			},
			sdk.ZeroInt(),
			sdk.ZeroInt(),
			true,
			sdk.ZeroInt(),
		},
		{
			"errors if mint amount is negative",
			sdk.Coins{sdk.Coin{Denom: "akava", Amount: sdk.NewInt(-100)}},
			sdk.ZeroInt(),
			sdk.NewInt(50),
			true,
			sdk.NewInt(50),
		},
		{
			"adds to existing akava balance",
			sdk.NewCoins(sdk.NewInt64Coin("akava", 12_021_000_000_002)),
			sdk.NewInt(12),
			sdk.NewInt(21_000_000_102),
			false,
			sdk.NewInt(100),
		},
		{
			"does not convert akava balance to ukava if it exceeds 1 ukava",
			sdk.NewCoins(sdk.NewInt64Coin("akava", 10_999_000_000_000)),
			sdk.NewInt(10),
			sdk.NewInt(2_001_200_000_001),
			false,
			sdk.NewInt(1_002_200_000_001),
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.SetupTest()

			err := suite.bk.MintCoins(suite.ctx, evmtypes.ModuleName, sdk.NewCoins(sdk.NewCoin("akava", tt.akavaStart)))
			suite.Require().NoError(err)

			err = suite.evmBankKeeper.MintCoins(suite.ctx, evmtypes.ModuleName, tt.mintCoins)
			if tt.hasErr {
				suite.Require().Error(err)
				return
			} else {
				suite.Require().NoError(err)
			}

			// check ukava
			ukavaActual := suite.bk.GetBalance(suite.ctx, suite.evmModuleAddr, "ukava")
			suite.Require().Equal(tt.ukava, ukavaActual.Amount)

			// check akava
			akavaActual := suite.bk.GetBalance(suite.ctx, suite.evmModuleAddr, "akava")
			suite.Require().Equal(tt.akava, akavaActual.Amount)
		})
	}
}

func TestEvmKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(evmKeeperTestSuite))
}
