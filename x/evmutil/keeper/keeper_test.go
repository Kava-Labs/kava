package keeper_test

import (
	"math/big"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/ethermint/x/evm/legacystatedb"
	"github.com/evmos/ethermint/x/evm/statedb"
	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/x/evmutil/keeper"
	"github.com/kava-labs/kava/x/evmutil/testutil"
	"github.com/kava-labs/kava/x/evmutil/types"
)

var (
	blockHash     common.Hash      = common.BigToHash(big.NewInt(9999))
	emptyTxConfig statedb.TxConfig = statedb.NewEmptyTxConfig(blockHash)
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

func (suite *keeperTestSuite) TestSupplyLoss() {
	// Native evm statedb can cause a net 1ukava burn instead of a net 0ukava
	// transfer of funds in tx. Example in height 8591928 on mainnet.

	// addr1 := common.HexToAddress("0x3E577e11198087Ee019e94D3b98e86c8EEb37a1C")
	// addr2 := common.HexToAddress("0x26DA582889f59EaaE9dA1f063bE0140CD93E6a4f")
	addr3 := common.HexToAddress("0x1a44076050125825900e736c501f859c50fE728c")
	// addr4 := common.HexToAddress("0x83Fb937054918cB7AccB15bd6cD9234dF9ebb357")

	init := func() {
		// Mint some coins for x/evmutil
		err := suite.BankKeeper.MintCoins(suite.Ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(10))))
		suite.Require().NoError(err)

		// Initial states
		// db := statedb.New(suite.Ctx, suite.App.GetEvmKeeper(), emptyTxConfig)
		// db.AddBalance(addr1, big.NewInt(150311671923249087))
		// db.AddBalance(addr2, big.NewInt(8523454805415132295))
		// db.AddBalance(addr3, keeper.ConversionMultiplier.MulRaw(2).BigInt())

		// Too big for big.NewInt int64
		// bigBal, success := new(big.Int).SetString("248358861987150296473", 10)
		// suite.Require().True(success)
		// db.AddBalance(addr4, bigBal)
		// suite.Require().NoError(db.Commit())

		// Verify balances are correctly set
		// ek := suite.App.GetEvmKeeper()
		// suite.Require().Equal(int64(150311671923249087), ek.GetBalance(suite.Ctx, addr1).Int64())
		// suite.Require().Equal(int64(8523454805415132295), ek.GetBalance(suite.Ctx, addr2).Int64())
		// suite.Require().Equal(int64(0), ek.GetBalance(suite.Ctx, addr3).Int64())
		// suite.Require().Equal("248358861987150296473", ek.GetBalance(suite.Ctx, addr4).String())
	}

	run := func(db vm.StateDB) {
		// 1ukava - 0.1ukava
		// NET = 0.9ukava, 0ukava should be minted

		// +1ukava - MINT 1
		db.AddBalance(addr3, keeper.ConversionMultiplier.BigInt())
		// -0.1ukava
		db.SubBalance(addr3, big.NewInt(100000000000))
	}

	init()
	balInit := suite.BankKeeper.GetBalance(suite.Ctx, suite.App.GetAccountKeeper().GetModuleAddress(types.ModuleName), "ukava")
	balUserInit := suite.BankKeeper.GetBalance(suite.Ctx, sdk.AccAddress(addr3.Bytes()), "ukava")

	// Create a new one to not reuse after Commit()
	db := statedb.New(suite.Ctx, suite.App.GetEvmKeeper(), emptyTxConfig)
	run(db)
	suite.Require().NoError(db.Commit())

	// Check balances
	ak := suite.App.GetAccountKeeper()
	bal1 := suite.BankKeeper.GetBalance(suite.Ctx, ak.GetModuleAddress(types.ModuleName), "ukava")
	userBal1 := suite.BankKeeper.GetBalance(suite.Ctx, sdk.AccAddress(addr3.Bytes()), "ukava")

	suite.Require().Equal(balInit.Amount.AddRaw(1), bal1.Amount)
	suite.T().Logf("x/evmutil bal (before, after): %v -> %v", balInit, bal1)
	suite.T().Logf("user x/bank bal: %v -> %v", balUserInit, userBal1)

	akavaAccs1 := suite.App.GetEvmutilKeeper().GetAllAccounts(suite.Ctx)

	// --------
	// Reset state
	suite.SetupTest()
	init()
	balInit = suite.BankKeeper.GetBalance(suite.Ctx, suite.App.GetAccountKeeper().GetModuleAddress(types.ModuleName), "ukava")
	balUserInit = suite.BankKeeper.GetBalance(suite.Ctx, sdk.AccAddress(addr3.Bytes()), "ukava")

	ldb := legacystatedb.New(suite.Ctx, suite.App.GetEvmKeeper(), emptyTxConfig)
	run(ldb)
	suite.Require().NoError(ldb.Commit())

	bal2 := suite.BankKeeper.GetBalance(suite.Ctx, ak.GetModuleAddress(types.ModuleName), "ukava")
	userBal2 := suite.BankKeeper.GetBalance(suite.Ctx, sdk.AccAddress(addr3.Bytes()), "ukava")
	suite.Assert().Equalf(bal1.Amount, bal2.Amount, "x/evmutil balance should be the same, hybrid bal: %v vanilla bal: %v", bal1, bal2)

	akavaAccs2 := suite.App.GetEvmutilKeeper().GetAllAccounts(suite.Ctx)
	suite.Assert().Equal(akavaAccs1, akavaAccs2, "x/evmutil accounts should have the same akava balances")
	suite.T().Logf("akavaAccs2: %v", akavaAccs2)

	suite.T().Logf("x/evmutil bal (before, after): %v -> %v", balInit, bal2)
	suite.T().Logf("user x/bank bal: %v -> %v", balUserInit, userBal2)
}

func (suite *keeperTestSuite) TestMintBurnOrder() {
	// Two different ways to result in the same balance, but with different bank
	// values in the process.
	// Mint 1ukava, burn 0.1ukava = 0.9ukava (1ukava evmutil x/bank balance - backed balance)
	// Mint 0.9ukava = 0.9ukava              (0ukava evmutil x/bank balance - not backed)
	//
	// Why can this happen?
	// The StateDB sets values by minting/burning the delta between the new and
	// old balance. This means the order of mint and burn operations is only
	// determined by which address balance is modified first and which direction
	// it goes. Mints are *not* always called first, even if balances are set
	// due to a zero sum transfer.
	//
	// Example: Want to transfer 1ukava from addr1 -> addr2
	// addr1: -1ukava
	// addr2: +1ukava
	// Balances are set in order of sorted addresses, in StateDB.Commit()
	// If addr1 is first -> burn 1ukava from addr1, then mint 1ukava for addr2
	// If addr2 is first -> mint 1ukava for addr2, then burn 1ukava from addr1
	//
	// # Unbacked akava
	// Mint 0.9ukava x10 -> 0ukava minted, 9ukava unbacked
	//
	// What is expected?
	// Mint and burns in any order should result in the same balance in both
	// accounts and backed module balance. akava balance should also be fully
	// backed.

	ek := suite.App.GetEvmutilKeeper()
	ebk := keeper.NewEvmBankKeeper(ek, suite.BankKeeper, suite.AccountKeeper)
	moduleAddr := suite.AccountKeeper.GetModuleAddress(types.ModuleName)

	// Mint 1ukava
	err := ebk.MintCoins(
		suite.Ctx,
		types.ModuleName,
		sdk.NewCoins(sdk.NewCoin(keeper.EvmDenom, keeper.ConversionMultiplier)),
	)
	suite.Require().NoError(err)

	// Burn 0.1ukava
	err = ebk.BurnCoins(
		suite.Ctx,
		types.ModuleName,
		sdk.NewCoins(sdk.NewCoin(keeper.EvmDenom, keeper.ConversionMultiplier.QuoRaw(10))),
	)
	suite.Require().NoError(err)

	// Check balance
	bal := suite.BankKeeper.GetBalance(suite.Ctx, moduleAddr, "ukava")
	akavaBal := suite.Keeper.GetBalance(suite.Ctx, moduleAddr)

	expAkava := "900000000000"
	suite.Require().Equal(sdk.NewInt(1), bal.Amount)
	suite.Require().Equal(expAkava, akavaBal.String())

	// ------------------------------------------------------------------------
	// Alternative way
	suite.SetupTest()
	ek = suite.App.GetEvmutilKeeper()
	ebk = keeper.NewEvmBankKeeper(ek, suite.BankKeeper, suite.AccountKeeper)

	// Mint 0.9ukava, no burn
	err = ebk.MintCoins(
		suite.Ctx,
		types.ModuleName,
		sdk.NewCoins(sdk.NewCoin(keeper.EvmDenom, keeper.ConversionMultiplier.QuoRaw(10).MulRaw(9))),
	)
	suite.Require().NoError(err)

	// Check balance
	bal = suite.BankKeeper.GetBalance(suite.Ctx, moduleAddr, "ukava")
	akavaBal = suite.Keeper.GetBalance(suite.Ctx, moduleAddr)
	suite.Require().Equal(sdk.NewInt(1), bal.Amount, "akava should be backed by 1ukava")
	suite.Require().Equal(expAkava, akavaBal.String())
}

func (suite *keeperTestSuite) TestTransferMintBurn() {
	addr1 := sdk.AccAddress{1}
	addr2 := sdk.AccAddress{2}

	ek := suite.App.GetEvmutilKeeper()
	ebk := keeper.NewEvmBankKeeper(ek, suite.BankKeeper, suite.AccountKeeper)
	moduleAddr := suite.AccountKeeper.GetModuleAddress(types.ModuleName)

	// Mint & send 10.1ukava for addr1 - 10ukava + 0.1ukava
	amt := keeper.ConversionMultiplier.MulRaw(10).
		Add(keeper.ConversionMultiplier.QuoRaw(10))
	err := ebk.MintCoins(
		suite.Ctx,
		types.ModuleName,
		sdk.NewCoins(sdk.NewCoin(keeper.EvmDenom, amt)),
	)
	suite.Require().NoError(err)

	// Send from module account to addr1
	err = ebk.SendCoinsFromModuleToAccount(
		suite.Ctx,
		types.ModuleName,
		addr1,
		sdk.NewCoins(sdk.NewCoin(keeper.EvmDenom, amt)),
	)
	suite.Require().NoError(err)

	// ------------------------
	// Transfer addr1 -> addr2
	// 1. Transfer addr1 -> module
	// 2. Burn 10.1ukava
	err = ebk.SendCoinsFromAccountToModule(
		suite.Ctx,
		addr1,
		types.ModuleName,
		sdk.NewCoins(sdk.NewCoin(keeper.EvmDenom, amt)),
	)
	suite.Require().NoError(err)

	err = ebk.BurnCoins(
		suite.Ctx,
		types.ModuleName,
		sdk.NewCoins(sdk.NewCoin(keeper.EvmDenom, amt)),
	)
	suite.Require().NoError(err)

	// 3. Mint 10.1ukava
	// 4. Transfer module -> addr2
	err = ebk.MintCoins(
		suite.Ctx,
		types.ModuleName,
		sdk.NewCoins(sdk.NewCoin(keeper.EvmDenom, amt)),
	)
	suite.Require().NoError(err)

	err = ebk.SendCoinsFromModuleToAccount(
		suite.Ctx,
		types.ModuleName,
		addr2,
		sdk.NewCoins(sdk.NewCoin(keeper.EvmDenom, amt)),
	)
	suite.Require().NoError(err)

	// Check balances
	bal1 := suite.BankKeeper.GetBalance(suite.Ctx, addr1, "ukava")
	bal2 := suite.BankKeeper.GetBalance(suite.Ctx, addr2, "ukava")

	akavaBal1 := suite.Keeper.GetBalance(suite.Ctx, addr1)
	akavaBal2 := suite.Keeper.GetBalance(suite.Ctx, addr2)

	kavaBalModule := suite.BankKeeper.GetBalance(suite.Ctx, moduleAddr, "ukava")
	akavaBalModule := suite.Keeper.GetBalance(suite.Ctx, moduleAddr)

	suite.Require().Equal("0", bal1.Amount.String())
	suite.Require().Equal("10", bal2.Amount.String())

	suite.Require().Equal("0", akavaBal1.String())
	suite.Require().Equal("100000000000", akavaBal2.String())

	suite.Require().Equal("0", kavaBalModule.Amount.String())
	suite.Require().Equal("0", akavaBalModule.String())
}

func (suite *keeperTestSuite) TestBurnUnbacked() {
	// burn an unbound amount of akava, as long as it is < 1ukava

	ek := suite.App.GetEvmutilKeeper()
	ebk := keeper.NewEvmBankKeeper(ek, suite.BankKeeper, suite.AccountKeeper)
	moduleAddr := suite.AccountKeeper.GetModuleAddress(types.ModuleName)

	// Start with module account balance of 10ukava
	err := ebk.MintCoins(
		suite.Ctx,
		types.ModuleName,
		sdk.NewCoins(sdk.NewCoin(
			keeper.EvmDenom, keeper.ConversionMultiplier.MulRaw(10),
		)),
	)
	suite.Require().NoError(err)

	initKavaBal := suite.BankKeeper.GetBalance(suite.Ctx, moduleAddr, "ukava")
	// initAkavaBal := suite.Keeper.GetBalance(suite.Ctx, moduleAddr)

	// Burn 0.1ukava 100 times (10ukava total equivalent)
	for i := 0; i < 100; i++ {
		err = ebk.BurnCoins(
			suite.Ctx,
			types.ModuleName,
			sdk.NewCoins(sdk.NewCoin(
				keeper.EvmDenom,
				keeper.ConversionMultiplier.QuoRaw(10),
			)),
		)
		suite.Require().NoError(err)
	}

	// Check balance
	bal := suite.BankKeeper.GetBalance(suite.Ctx, moduleAddr, "ukava")
	akavaBal := suite.Keeper.GetBalance(suite.Ctx, moduleAddr)

	suite.Require().Equal(
		initKavaBal.Amount.SubRaw(10).String(),
		bal.Amount.String(),
		"10 ukava should be burned",
	)
	suite.Require().Equal("0", akavaBal.String())
}

func (suite *keeperTestSuite) TestBurnLoss() {
	// x/evmutil ukava bank balance loss

	// ----------------------------------------
	// # Extra burn than it should?
	// Two separate transfers:
	// 0.6ukava addr1 -> addr2
	// 0.6ukava addr1 -> addr3
	// Total account sent is 1.2ukava
	//
	// Vanilla (transfer amounts merged)
	// -1.2ukava from addr1 (burn 1)
	// +0.6ukava to addr2 (mint 0)
	// +0.6ukava to addr3 (mint 0)
	// Burn and mints are unbalanced with a zero sum transfer
	//
	// Hybrid (transfer amounts kept separate)
	// -0.6ukava from addr1 (burn 0)
	// -0.6ukava from addr1 (burn 0)
	// +0.6ukava to addr2 (mint 0)
	// +0.6ukava to addr3 (mint 0)

	addr1 := sdk.AccAddress{1}
	addr2 := sdk.AccAddress{2}
	addr3 := sdk.AccAddress{3}
	moduleAddr := suite.AccountKeeper.GetModuleAddress(types.ModuleName)

	initialAmt := keeper.ConversionMultiplier.MulRaw(10)

	setupFn := func() {
		ek := suite.App.GetEvmutilKeeper()
		ebk := keeper.NewEvmBankKeeper(ek, suite.BankKeeper, suite.AccountKeeper)
		// Start with 10ukava for module, addr1, addr2
		suite.BankKeeper.MintCoins(suite.Ctx, types.ModuleName, sdk.NewCoins(sdk.NewInt64Coin("ukava", 10)))
		suite.addBalance(ebk, addr1, initialAmt)
		suite.addBalance(ebk, addr2, initialAmt)
		suite.addBalance(ebk, addr3, initialAmt)
	}

	type expectedBalances struct {
		addr1 string
		addr2 string
		addr3 string
	}

	amt0_6 := keeper.ConversionMultiplier.QuoRaw(10).MulRaw(6)

	tests := []struct {
		name         string
		run          func(ebk keeper.EvmBankKeeper) []int // return list of expected module balance changes
		expectedBals expectedBalances
	}{
		{
			"aggregate 1.2ukava transfer",
			func(ebk keeper.EvmBankKeeper) []int {
				// Vanilla - set the total net instead of each transfer separately
				return []int{
					suite.subBalance(ebk, addr1, amt0_6.MulRaw(2)), // -1.2ukava (burn 1, convert 1ukava, total module decrease 2ukava)
					suite.addBalance(ebk, addr2, amt0_6),           // +0.6ukava (mint 0)
					suite.addBalance(ebk, addr3, amt0_6),           // +0.6ukava (mint 0)
				}
			},
			expectedBalances{
				addr1: "8800000000000",
				addr2: "10600000000000",
				addr3: "10600000000000",
			},
		},
		{
			"split 1.2ukava transfer",
			func(ebk keeper.EvmBankKeeper) []int {

				return []int{
					suite.subBalance(ebk, addr1, amt0_6), // -0.6ukava (burn 0)
					suite.subBalance(ebk, addr1, amt0_6), // -0.6ukava (burn 0)
					suite.addBalance(ebk, addr2, amt0_6), // +0.6ukava (mint 0) - convert 1ukava -> akava
					suite.addBalance(ebk, addr3, amt0_6), // +0.6ukava (mint 0) - convert 1ukava -> akava
				}
			},
			expectedBalances{
				addr1: initialAmt.Sub(amt0_6.MulRaw(2)).String(),
				addr2: initialAmt.Add(amt0_6).String(),
				addr3: initialAmt.Add(amt0_6).String(),
			},
		},
		{
			"reverse after send",
			func(ebk keeper.EvmBankKeeper) []int {
				return []int{
					suite.addBalance(ebk, addr1, amt0_6), // 10.6 - mint 0ukava - no convert
					suite.addBalance(ebk, addr1, amt0_6), // 11.2 - mint 0ukava - convert akava -> ukava
					suite.subBalance(ebk, addr2, amt0_6), // 9.4 - burn 0ukava - convert ukava -> akava
					suite.subBalance(ebk, addr3, amt0_6), // 9.4 - burn 0ukava - convert ukava -> akava

					// Send the same values back
					suite.subBalance(ebk, addr1, amt0_6), // 10.6 - burn 0.6ukava - convert ukava -> akava
					suite.subBalance(ebk, addr1, amt0_6), // 10.0 - burn 0.6ukava - no convert
					suite.addBalance(ebk, addr2, amt0_6), // 10.0 - mint 0.6ukava - convert akava -> ukava
					suite.addBalance(ebk, addr3, amt0_6), // 10.0 - mint 0.6ukava - convert akava -> ukava
				}
			},
			expectedBalances{
				// Should end in the same initial balance
				addr1: "10000000000000",
				addr2: "10000000000000",
				addr3: "10000000000000",
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Rest test state
			suite.SetupTest()
			// Add initial balances
			setupFn()

			// Run test
			ebk := keeper.NewEvmBankKeeper(suite.App.GetEvmutilKeeper(), suite.BankKeeper, suite.AccountKeeper)

			// Return value is a slice of ints:
			//  +1 for expected ukava -> akava conversion (user -> module 1ukava for akava)
			//  -1 for expected akava -> ukava conversion (module -> user 1ukava for akava)
			// Use this instead of manually setting a hardcoded expected value
			// since it's quite complex and confusing to calculate.
			moduleDeltas := tt.run(ebk)

			moduleDeltaSum := 0
			for _, delta := range moduleDeltas {
				moduleDeltaSum += delta
			}

			// Check module balance
			moduleBal := ebk.GetBalance(suite.Ctx, moduleAddr, keeper.EvmDenom)
			suite.T().Logf("module bal: %v", moduleBal.Amount.String())

			expectedAmt := initialAmt.
				Add(keeper.ConversionMultiplier.MulRaw(int64(moduleDeltaSum)))

			suite.Assert().Equalf(
				expectedAmt.String(),
				moduleBal.Amount.String(),
				"incorrect module balance, expected %s, got %s - conversion deltas: %v",
				expectedAmt,
				moduleBal.Amount.String(),
				moduleDeltas,
			)

			// Check user address balances
			addrs := []struct {
				name        string
				addr        sdk.AccAddress
				expectedBal string
			}{
				{"addr1", addr1, tt.expectedBals.addr1},
				{"addr2", addr2, tt.expectedBals.addr2},
				{"addr3", addr3, tt.expectedBals.addr3},
			}

			for _, addr := range addrs {
				bal := ebk.GetBalance(suite.Ctx, addr.addr, keeper.EvmDenom)

				suite.T().Logf("%s bal: %v", addr.name, bal.Amount.String())
				suite.Assert().Equalf(
					addr.expectedBal,
					bal.Amount.String(),
					"incorrect balance for %s, expected %s, got %s",
					addr.name,
					addr.expectedBal,
					bal.Amount.String(),
				)
			}
		})
	}
}

func (suite *keeperTestSuite) subBalance(
	ebk keeper.EvmBankKeeper,
	from sdk.AccAddress,
	amt sdkmath.Int,
) int {
	ukava, akava, err := keeper.SplitAkavaCoins(sdk.NewCoins(sdk.NewCoin(keeper.EvmDenom, amt)))
	suite.Require().NoError(err)

	// Check if this balance sub will convert 1ukava -> akava
	prevAkava := suite.Keeper.GetBalance(suite.Ctx, from)
	// e.g. 9 > 8, 8-9 will need borrow
	willConvertToAkava := prevAkava.LT(akava)

	// Transfer by burning and minting
	err = ebk.SendCoinsFromAccountToModule(
		suite.Ctx,
		from,
		types.ModuleName,
		sdk.NewCoins(sdk.NewCoin(keeper.EvmDenom, amt)),
	)
	suite.Require().NoError(err)

	err = ebk.BurnCoins(
		suite.Ctx,
		types.ModuleName,
		sdk.NewCoins(sdk.NewCoin(keeper.EvmDenom, amt)),
	)
	suite.Require().NoError(err)

	// decrease by burn regular ukava amount
	// User converted 1ukava to akava, so module ukava bal also increases by 1
	if willConvertToAkava {
		return int(ukava.Amount.Int64()) + 1
	}

	return int(ukava.Amount.Int64())
}

func (suite *keeperTestSuite) addBalance(
	ebk keeper.EvmBankKeeper,
	to sdk.AccAddress,
	amt sdkmath.Int,
) int {
	ukava, akava, err := keeper.SplitAkavaCoins(sdk.NewCoins(sdk.NewCoin(keeper.EvmDenom, amt)))
	suite.Require().NoError(err)

	// Check if this balance add will convert akava -> 1ukava (rollover)
	prevAkava := suite.Keeper.GetBalance(suite.Ctx, to)
	// if current_akava + new_akava => 1ukava, akava will be converted back to 1ukava
	willConvertToUkava := akava.Add(prevAkava).GTE(keeper.ConversionMultiplier)

	// Mint and transfer
	err = ebk.MintCoins(
		suite.Ctx,
		types.ModuleName,
		sdk.NewCoins(sdk.NewCoin(keeper.EvmDenom, amt)),
	)
	suite.Require().NoError(err)

	err = ebk.SendCoinsFromModuleToAccount(
		suite.Ctx,
		types.ModuleName,
		to,
		sdk.NewCoins(sdk.NewCoin(keeper.EvmDenom, amt)),
	)
	suite.Require().NoError(err)

	// Whole ukava will be minted.
	// User converted akava to 1ukava, so module ukava bal decreases by 1
	if willConvertToUkava {
		return int(ukava.Amount.Int64()) - 1
	}

	return int(ukava.Amount.Int64())
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(keeperTestSuite))
}
