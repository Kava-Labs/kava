package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/x/evmutil/testutil"
	"github.com/kava-labs/kava/x/evmutil/types"
)

type keeperTestSuite struct {
	testutil.Suite
}

func (suite *keeperTestSuite) SetupTest() {
	suite.Suite.SetupTest()
}

func (suite *keeperTestSuite) TestGetAllAccounts() {
	tests := []struct {
		name        string
		expAccounts []types.Account
	}{
		{
			"no accounts",
			[]types.Account{},
		},
		{
			"with accounts",
			[]types.Account{
				{Address: suite.Addrs[0], Balance: sdk.NewInt(100)},
				{Address: suite.Addrs[1], Balance: sdk.NewInt(200)},
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.SetupTest()

			for _, account := range tt.expAccounts {
				suite.Keeper.SetBalance(suite.Ctx, account.Address, account.Balance)
			}

			accounts := suite.Suite.Keeper.GetAllAccounts(suite.Ctx)
			if len(tt.expAccounts) == 0 {
				suite.Require().Len(tt.expAccounts, 0)
			} else {
				suite.Require().Equal(tt.expAccounts, accounts)
			}
		})
	}
}

func (suite *keeperTestSuite) TestSetAccount() {
	existingAccount := types.Account{
		Address: suite.Addrs[0],
		Balance: sdk.NewInt(100),
	}
	tests := []struct {
		name    string
		account types.Account
		success bool
	}{
		{
			"invalid address",
			types.Account{Address: nil, Balance: sdk.NewInt(100)},
			false,
		},
		{
			"invalid balance",
			types.Account{Address: suite.Addrs[0], Balance: sdk.NewInt(-100)},
			false,
		},
		{
			"empty account",
			types.Account{},
			false,
		},
		{
			"valid account",
			types.Account{Address: suite.Addrs[1], Balance: sdk.NewInt(100)},
			true,
		},
		{
			"replaces account",
			types.Account{Address: suite.Addrs[0], Balance: sdk.NewInt(50)},
			true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.SetupTest()

			err := suite.Keeper.SetAccount(suite.Ctx, existingAccount)
			suite.Require().NoError(err)
			err = suite.Keeper.SetAccount(suite.Ctx, tt.account)
			if tt.success {
				suite.Require().NoError(err)
				expAcct := suite.Keeper.GetAccount(suite.Ctx, tt.account.Address)
				suite.Require().Equal(tt.account, *expAcct)
			} else {
				suite.Require().Error(err)
				suite.Require().Nil(suite.Keeper.GetAccount(suite.Ctx, suite.Addrs[1]))
			}
		})
	}
}

func (suite *keeperTestSuite) TestSendBalance() {
	startingSenderBal := sdk.NewInt(100)
	startingRecipientBal := sdk.NewInt(50)
	tests := []struct {
		name            string
		amt             sdk.Int
		expSenderBal    sdk.Int
		expRecipientBal sdk.Int
		success         bool
	}{
		{
			"fails when sending negative amount",
			sdk.NewInt(-5),
			sdk.ZeroInt(),
			sdk.ZeroInt(),
			false,
		},
		{
			"send zero amount",
			sdk.ZeroInt(),
			startingSenderBal,
			startingRecipientBal,
			true,
		},
		{
			"fails when sender does not have enough balance",
			sdk.NewInt(101),
			sdk.ZeroInt(),
			sdk.ZeroInt(),
			false,
		},
		{
			"send valid amount",
			sdk.NewInt(80),
			sdk.NewInt(20),
			sdk.NewInt(130),
			true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.SetupTest()

			err := suite.Keeper.SetBalance(suite.Ctx, suite.Addrs[0], startingSenderBal)
			suite.Require().NoError(err)
			err = suite.Keeper.SetBalance(suite.Ctx, suite.Addrs[1], startingRecipientBal)
			suite.Require().NoError(err)

			err = suite.Keeper.SendBalance(suite.Ctx, suite.Addrs[0], suite.Addrs[1], tt.amt)
			if tt.success {
				suite.Require().NoError(err)
				suite.Require().Equal(tt.expSenderBal, suite.Keeper.GetBalance(suite.Ctx, suite.Addrs[0]))
				suite.Require().Equal(tt.expRecipientBal, suite.Keeper.GetBalance(suite.Ctx, suite.Addrs[1]))
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *keeperTestSuite) TestSetBalance() {
	existingAccount := types.Account{
		Address: suite.Addrs[0],
		Balance: sdk.NewInt(100),
	}
	tests := []struct {
		name    string
		address sdk.AccAddress
		balance sdk.Int
		success bool
	}{
		{
			"invalid balance",
			suite.Addrs[0],
			sdk.NewInt(-100),
			false,
		},
		{
			"set new account balance",
			suite.Addrs[1],
			sdk.NewInt(100),
			true,
		},
		{
			"replace account balance",
			suite.Addrs[0],
			sdk.NewInt(50),
			true,
		},
		{
			"invalid address",
			nil,
			sdk.NewInt(100),
			false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.SetupTest()

			err := suite.Keeper.SetAccount(suite.Ctx, existingAccount)
			suite.Require().NoError(err)
			err = suite.Keeper.SetBalance(suite.Ctx, tt.address, tt.balance)
			if tt.success {
				suite.Require().NoError(err)
				expBal := suite.Keeper.GetBalance(suite.Ctx, tt.address)
				suite.Require().Equal(expBal, tt.balance)
			} else {
				suite.Require().Error(err)
				expBal := suite.Keeper.GetBalance(suite.Ctx, existingAccount.Address)
				suite.Require().Equal(expBal, existingAccount.Balance)
			}
		})
	}
}

func (suite *keeperTestSuite) TestRemoveBalance() {
	existingAccount := types.Account{
		Address: suite.Addrs[0],
		Balance: sdk.NewInt(100),
	}
	tests := []struct {
		name    string
		amt     sdk.Int
		expBal  sdk.Int
		success bool
	}{
		{
			"fails if amount is negative",
			sdk.NewInt(-10),
			sdk.ZeroInt(),
			false,
		},
		{
			"remove zero amount",
			sdk.ZeroInt(),
			existingAccount.Balance,
			true,
		},
		{
			"not enough balance",
			sdk.NewInt(101),
			sdk.ZeroInt(),
			false,
		},
		{
			"remove full balance",
			sdk.NewInt(100),
			sdk.ZeroInt(),
			true,
		},
		{
			"remove some balance",
			sdk.NewInt(10),
			sdk.NewInt(90),
			true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.SetupTest()

			err := suite.Keeper.SetAccount(suite.Ctx, existingAccount)
			suite.Require().NoError(err)
			err = suite.Keeper.RemoveBalance(suite.Ctx, existingAccount.Address, tt.amt)
			if tt.success {
				suite.Require().NoError(err)
				expBal := suite.Keeper.GetBalance(suite.Ctx, existingAccount.Address)
				suite.Require().Equal(expBal, tt.expBal)
			} else {
				suite.Require().Error(err)
				expBal := suite.Keeper.GetBalance(suite.Ctx, existingAccount.Address)
				suite.Require().Equal(expBal, existingAccount.Balance)
			}
		})
	}
}

func (suite *keeperTestSuite) TestGetBalance() {
	existingAccount := types.Account{
		Address: suite.Addrs[0],
		Balance: sdk.NewInt(100),
	}
	tests := []struct {
		name   string
		addr   sdk.AccAddress
		expBal sdk.Int
	}{
		{
			"returns 0 balance if account does not exist",
			suite.Addrs[1],
			sdk.ZeroInt(),
		},
		{
			"returns account balance",
			suite.Addrs[0],
			sdk.NewInt(100),
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.SetupTest()

			err := suite.Keeper.SetAccount(suite.Ctx, existingAccount)
			suite.Require().NoError(err)
			balance := suite.Keeper.GetBalance(suite.Ctx, tt.addr)
			suite.Require().Equal(tt.expBal, balance)
		})
	}
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(keeperTestSuite))
}
