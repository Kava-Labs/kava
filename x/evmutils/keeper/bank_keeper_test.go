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

func (suite *evmKeeperTestSuite) TestGetBalance_ReturnsSpendable() {
	startingCoins := sdk.NewCoins(
		sdk.NewInt64Coin("akava", 100),
		sdk.NewInt64Coin("ukava", 10),
	)

	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)
	bacc := authtypes.NewBaseAccountWithAddress(suite.addrs[0])
	vacc := vesting.NewContinuousVestingAccount(bacc, startingCoins, now.Unix(), endTime.Unix())
	suite.ak.SetAccount(suite.ctx, vacc)

	err := suite.app.FundAccount(suite.ctx, suite.addrs[0], startingCoins)
	suite.Require().NoError(err)
	coin := suite.evmBankKeeper.GetBalance(suite.ctx, suite.addrs[0], "akava")
	suite.Require().Equal(sdk.ZeroInt(), coin.Amount)

	ctx := suite.ctx.WithBlockTime(now.Add(12 * time.Hour))
	coin = suite.evmBankKeeper.GetBalance(ctx, suite.addrs[0], "akava")
	suite.Require().Equal(sdk.NewIntFromUint64(5_000_000_000_050), coin.Amount)
}

func (suite *evmKeeperTestSuite) TestGetBalance_NotEvmDenom() {
	suite.Require().Panics(func() {
		suite.evmBankKeeper.GetBalance(suite.ctx, suite.addrs[0], "ukava")
	})
	suite.Require().Panics(func() {
		suite.evmBankKeeper.GetBalance(suite.ctx, suite.addrs[0], "busd")
	})
}

func (suite *evmKeeperTestSuite) TestGetBalance() {
	tests := []struct {
		name           string
		startingAmount sdk.Coins
		expAmount      sdk.Int
	}{
		{
			"ukava with akava",
			sdk.NewCoins(
				sdk.NewInt64Coin("akava", 100),
				sdk.NewInt64Coin("ukava", 10),
			),
			sdk.NewInt(10_000_000_000_100),
		},
		{
			"just akava",
			sdk.NewCoins(
				sdk.NewInt64Coin("akava", 100),
				sdk.NewInt64Coin("busd", 100),
			),
			sdk.NewInt(100),
		},
		{
			"just ukava",
			sdk.NewCoins(
				sdk.NewInt64Coin("ukava", 10),
				sdk.NewInt64Coin("busd", 100),
			),
			sdk.NewInt(10_000_000_000_000),
		},
		{
			"no ukava or akava",
			sdk.NewCoins(),
			sdk.ZeroInt(),
		},
		{
			"with avaka that is more than 1 ukava",
			sdk.NewCoins(
				sdk.NewInt64Coin("akava", 20_000_000_000_220),
				sdk.NewInt64Coin("ukava", 11),
			),
			sdk.NewInt(31_000_000_000_220),
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.SetupTest()

			err := suite.app.FundAccount(suite.ctx, suite.addrs[0], tt.startingAmount)
			suite.Require().NoError(err)

			coin := suite.evmBankKeeper.GetBalance(suite.ctx, suite.addrs[0], "akava")
			suite.Require().Equal(tt.expAmount, coin.Amount)
		})
	}
}

func (suite *evmKeeperTestSuite) TestSendCoinsFromModuleToAccount() {
	startingCoins := sdk.NewCoins(
		sdk.NewInt64Coin("akava", 200),
		sdk.NewInt64Coin("ukava", 100),
	)
	tests := []struct {
		name           string
		sendCoins      sdk.Coins
		startingAccBal sdk.Coins
		expAccBal      sdk.Coins
		hasErr         bool
	}{
		{
			"send more than 1 ukava",
			sdk.NewCoins(sdk.NewInt64Coin("akava", 12_000_000_000_010)),
			sdk.Coins{},
			sdk.NewCoins(
				sdk.NewInt64Coin("akava", 10),
				sdk.NewInt64Coin("ukava", 12),
			),
			false,
		},
		{
			"send less than 1 ukava",
			sdk.NewCoins(sdk.NewInt64Coin("akava", 122)),
			sdk.Coins{},
			sdk.NewCoins(
				sdk.NewInt64Coin("akava", 122),
				sdk.NewInt64Coin("ukava", 0),
			),
			false,
		},
		{
			"send an exact amount of ukava",
			sdk.NewCoins(sdk.NewInt64Coin("akava", 98_000_000_000_000)),
			sdk.Coins{},
			sdk.NewCoins(
				sdk.NewInt64Coin("akava", 00),
				sdk.NewInt64Coin("ukava", 98),
			),
			false,
		},
		{
			"send no akava",
			sdk.NewCoins(sdk.NewInt64Coin("akava", 0)),
			sdk.Coins{},
			sdk.NewCoins(
				sdk.NewInt64Coin("akava", 0),
				sdk.NewInt64Coin("ukava", 0),
			),
			false,
		},
		{
			"errors if sending other coins",
			sdk.NewCoins(sdk.NewInt64Coin("akava", 500), sdk.NewInt64Coin("busd", 1000)),
			sdk.Coins{},
			sdk.Coins{},
			true,
		},
		{
			"errors if not enough total akava to cover",
			sdk.NewCoins(sdk.NewInt64Coin("akava", 90_000_000_001_000)),
			sdk.Coins{},
			sdk.Coins{},
			true,
		},
		{
			"errors if not enough ukava to cover",
			sdk.NewCoins(sdk.NewInt64Coin("akava", 200_000_000_000_000)),
			sdk.Coins{},
			sdk.Coins{},
			true,
		},
		{
			"converts receiver's akava to ukava there's enough akava after the transfer",
			sdk.NewCoins(sdk.NewInt64Coin("akava", 99_000_000_000_200)),
			sdk.NewCoins(
				sdk.NewInt64Coin("akava", 999_999_999_900),
				sdk.NewInt64Coin("ukava", 1),
			),
			sdk.NewCoins(
				sdk.NewInt64Coin("akava", 100),
				sdk.NewInt64Coin("ukava", 101),
			),
			false,
		},
		{
			"converts all of receiver's akava to ukava even if somehow receiver has more than 1ukava of akava",
			sdk.NewCoins(sdk.NewInt64Coin("akava", 12_000_000_000_100)),
			sdk.NewCoins(
				sdk.NewInt64Coin("akava", 5_999_999_999_990),
				sdk.NewInt64Coin("ukava", 1),
			),
			sdk.NewCoins(
				sdk.NewInt64Coin("akava", 90),
				sdk.NewInt64Coin("ukava", 19),
			),
			false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.SetupTest()

			err := suite.app.FundAccount(suite.ctx, suite.addrs[0], tt.startingAccBal)
			suite.Require().NoError(err)

			suite.Require().NoError(err)
			err = suite.bk.MintCoins(suite.ctx, evmtypes.ModuleName, startingCoins)
			suite.Require().NoError(err)

			err = suite.evmBankKeeper.SendCoinsFromModuleToAccount(suite.ctx, evmtypes.ModuleName, suite.addrs[0], tt.sendCoins)
			if tt.hasErr {
				suite.Require().Error(err)
				return
			} else {
				suite.Require().NoError(err)
			}

			// check ukava
			ukavaSender := suite.bk.GetBalance(suite.ctx, suite.addrs[0], "ukava")
			suite.Require().Equal(tt.expAccBal.AmountOf("ukava").Int64(), ukavaSender.Amount.Int64())

			// check akava
			akavaSender := suite.bk.GetBalance(suite.ctx, suite.addrs[0], "akava")
			suite.Require().Equal(tt.expAccBal.AmountOf("akava").Int64(), akavaSender.Amount.Int64())
		})
	}
}

func (suite *evmKeeperTestSuite) TestSendCoinsFromAccountToModule() {
	startingCoins := sdk.NewCoins(
		sdk.NewInt64Coin("akava", 200),
		sdk.NewInt64Coin("ukava", 100),
	)
	tests := []struct {
		name      string
		sendCoins sdk.Coins
		ukava     sdk.Int
		akava     sdk.Int
		hasErr    bool
	}{
		{
			"send more than 1 ukava",
			sdk.NewCoins(sdk.NewInt64Coin("akava", 12_000_000_000_010)),
			sdk.NewInt(88),
			sdk.NewInt(190),
			false,
		},
		{
			"send less than 1 ukava",
			sdk.NewCoins(sdk.NewInt64Coin("akava", 122)),
			sdk.NewInt(100),
			sdk.NewInt(78),
			false,
		},
		{
			"send an exact amount of ukava",
			sdk.NewCoins(sdk.NewInt64Coin("akava", 98_000_000_000_000)),
			sdk.NewInt(2),
			sdk.NewInt(200),
			false,
		},
		{
			"send no akava",
			sdk.NewCoins(sdk.NewInt64Coin("akava", 0)),
			sdk.NewInt(100),
			sdk.NewInt(200),
			false,
		},
		{
			"errors if sending other coins",
			sdk.NewCoins(sdk.NewInt64Coin("akava", 500), sdk.NewInt64Coin("busd", 1000)),
			sdk.NewInt(100),
			sdk.NewInt(200),
			true,
		},
		{
			"errors if have dup coins",
			sdk.Coins{
				sdk.NewInt64Coin("akava", 12_000_000_000_000),
				sdk.NewInt64Coin("akava", 2_000_000_000_000),
			},
			sdk.NewInt(100),
			sdk.NewInt(200),
			true,
		},
		{
			"errors if not enough total akava to cover",
			sdk.NewCoins(sdk.NewInt64Coin("akava", 100_000_000_001_000)),
			sdk.NewInt(100),
			sdk.NewInt(200),
			true,
		},
		{
			"errors if not enough ukava to cover",
			sdk.NewCoins(sdk.NewInt64Coin("akava", 200_000_000_000_000)),
			sdk.NewInt(100),
			sdk.NewInt(200),
			true,
		},
		{
			"converts 1 ukava to akava if not enough akava to cover",
			sdk.NewCoins(sdk.NewInt64Coin("akava", 99_900_000_000_000)),
			sdk.NewInt(0),
			sdk.NewInt(100_000_000_200),
			false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.SetupTest()

			err := suite.app.FundAccount(suite.ctx, suite.addrs[0], startingCoins)
			suite.Require().NoError(err)

			err = suite.evmBankKeeper.SendCoinsFromAccountToModule(suite.ctx, suite.addrs[0], evmtypes.ModuleName, tt.sendCoins)
			if tt.hasErr {
				suite.Require().Error(err)
				return
			} else {
				suite.Require().NoError(err)
			}

			// check ukava
			ukavaSender := suite.bk.GetBalance(suite.ctx, suite.addrs[0], "ukava")
			suite.Require().Equal(tt.ukava, ukavaSender.Amount)

			// check akava
			akavaSender := suite.bk.GetBalance(suite.ctx, suite.addrs[0], "akava")
			suite.Require().Equal(tt.akava, akavaSender.Amount)
		})
	}
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
