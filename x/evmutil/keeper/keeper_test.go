package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
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
				{Address: suite.Addrs[0], Balance: sdkmath.NewInt(100)},
				{Address: suite.Addrs[1], Balance: sdkmath.NewInt(200)},
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

func (suite *keeperTestSuite) TestSetAccount_ZeroBalance() {
	existingAccount := types.Account{
		Address: suite.Addrs[0],
		Balance: sdkmath.NewInt(100),
	}
	err := suite.Keeper.SetAccount(suite.Ctx, existingAccount)
	suite.Require().NoError(err)
	err = suite.Keeper.SetAccount(suite.Ctx, types.Account{
		Address: suite.Addrs[0],
		Balance: sdk.ZeroInt(),
	})
	suite.Require().NoError(err)
	bal := suite.Keeper.GetBalance(suite.Ctx, suite.Addrs[0])
	suite.Require().Equal(sdk.ZeroInt(), bal)
	expAcct := suite.Keeper.GetAccount(suite.Ctx, suite.Addrs[0])
	suite.Require().Nil(expAcct)
}

func (suite *keeperTestSuite) TestSetAccount() {
	existingAccount := types.Account{
		Address: suite.Addrs[0],
		Balance: sdkmath.NewInt(100),
	}
	tests := []struct {
		name    string
		account types.Account
		success bool
	}{
		{
			"invalid address",
			types.Account{Address: nil, Balance: sdkmath.NewInt(100)},
			false,
		},
		{
			"invalid balance",
			types.Account{Address: suite.Addrs[0], Balance: sdkmath.NewInt(-100)},
			false,
		},
		{
			"empty account",
			types.Account{},
			false,
		},
		{
			"valid account",
			types.Account{Address: suite.Addrs[1], Balance: sdkmath.NewInt(100)},
			true,
		},
		{
			"replaces account",
			types.Account{Address: suite.Addrs[0], Balance: sdkmath.NewInt(50)},
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
	startingSenderBal := sdkmath.NewInt(100)
	startingRecipientBal := sdkmath.NewInt(50)
	tests := []struct {
		name            string
		amt             sdkmath.Int
		expSenderBal    sdkmath.Int
		expRecipientBal sdkmath.Int
		success         bool
	}{
		{
			"fails when sending negative amount",
			sdkmath.NewInt(-5),
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
			sdkmath.NewInt(101),
			sdk.ZeroInt(),
			sdk.ZeroInt(),
			false,
		},
		{
			"send valid amount",
			sdkmath.NewInt(80),
			sdkmath.NewInt(20),
			sdkmath.NewInt(130),
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
		Balance: sdkmath.NewInt(100),
	}
	tests := []struct {
		name    string
		address sdk.AccAddress
		balance sdkmath.Int
		success bool
	}{
		{
			"invalid balance",
			suite.Addrs[0],
			sdkmath.NewInt(-100),
			false,
		},
		{
			"set new account balance",
			suite.Addrs[1],
			sdkmath.NewInt(100),
			true,
		},
		{
			"replace account balance",
			suite.Addrs[0],
			sdkmath.NewInt(50),
			true,
		},
		{
			"invalid address",
			nil,
			sdkmath.NewInt(100),
			false,
		},
		{
			"zero balance",
			suite.Addrs[0],
			sdk.ZeroInt(),
			true,
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

				if tt.balance.IsZero() {
					account := suite.Keeper.GetAccount(suite.Ctx, tt.address)
					suite.Require().Nil(account)
				}
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
		Balance: sdkmath.NewInt(100),
	}
	tests := []struct {
		name    string
		amt     sdkmath.Int
		expBal  sdkmath.Int
		success bool
	}{
		{
			"fails if amount is negative",
			sdkmath.NewInt(-10),
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
			sdkmath.NewInt(101),
			sdk.ZeroInt(),
			false,
		},
		{
			"remove full balance",
			sdkmath.NewInt(100),
			sdk.ZeroInt(),
			true,
		},
		{
			"remove some balance",
			sdkmath.NewInt(10),
			sdkmath.NewInt(90),
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
		Balance: sdkmath.NewInt(100),
	}
	tests := []struct {
		name   string
		addr   sdk.AccAddress
		expBal sdkmath.Int
	}{
		{
			"returns 0 balance if account does not exist",
			suite.Addrs[1],
			sdk.ZeroInt(),
		},
		{
			"returns account balance",
			suite.Addrs[0],
			sdkmath.NewInt(100),
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

func (suite *keeperTestSuite) TestDeployedCosmosCoinContractStoreState() {
	suite.Run("returns nil for nonexistent denom", func() {
		suite.SetupTest()
		addr, found := suite.Keeper.GetDeployedCosmosCoinContract(suite.Ctx, "undeployed-denom")
		suite.False(found)
		suite.Equal(addr, types.InternalEVMAddress{})
	})

	suite.Run("handles setting & getting a contract address", func() {
		suite.SetupTest()
		denom := "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2"
		address := testutil.RandomInternalEVMAddress()

		err := suite.Keeper.SetDeployedCosmosCoinContract(suite.Ctx, denom, address)
		suite.NoError(err)

		stored, found := suite.Keeper.GetDeployedCosmosCoinContract(suite.Ctx, denom)
		suite.True(found)
		suite.Equal(address, stored)
	})

	suite.Run("fails when setting an invalid denom", func() {
		suite.SetupTest()
		invalidDenom := ""
		err := suite.Keeper.SetDeployedCosmosCoinContract(suite.Ctx, invalidDenom, testutil.RandomInternalEVMAddress())
		suite.ErrorContains(err, "invalid cosmos denom")
	})

	suite.Run("fails when setting 0 address", func() {
		suite.SetupTest()
		invalidAddr := types.InternalEVMAddress{}
		err := suite.Keeper.SetDeployedCosmosCoinContract(suite.Ctx, "denom", invalidAddr)
		suite.ErrorContains(err, "attempting to register empty contract address")
	})
}

func (suite *keeperTestSuite) TestIterateAllDeployedCosmosCoinContracts() {
	suite.SetupTest()
	address := testutil.RandomInternalEVMAddress()
	denoms := []string{}
	register := func(denom string) {
		addr := testutil.RandomInternalEVMAddress()
		if denom == "waldo" {
			addr = address
		}
		err := suite.Keeper.SetDeployedCosmosCoinContract(suite.Ctx, denom, addr)
		suite.NoError(err)
		denoms = append(denoms, denom)
	}

	// register some contracts
	register("magic")
	register("popcorn")
	register("waldo")
	register("zzz")
	register("ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2")

	suite.Run("stops when told", func() {
		// test out stopping the iteration
		// NOTE: don't actually look for a single contract this way. the keys are deterministic by denom.
		var contract types.DeployedCosmosCoinContract
		suite.Keeper.IterateAllDeployedCosmosCoinContracts(suite.Ctx, func(c types.DeployedCosmosCoinContract) bool {
			contract = c
			return c.CosmosDenom == "waldo"
		})
		suite.Equal(types.NewDeployedCosmosCoinContract("waldo", address), contract)
	})

	suite.Run("iterates all contracts", func() {
		foundDenoms := make([]string, 0, len(denoms))
		suite.Keeper.IterateAllDeployedCosmosCoinContracts(suite.Ctx, func(c types.DeployedCosmosCoinContract) bool {
			foundDenoms = append(foundDenoms, c.CosmosDenom)
			return false
		})
		suite.Len(foundDenoms, len(denoms))
		suite.ElementsMatch(denoms, foundDenoms)
	})
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(keeperTestSuite))
}
