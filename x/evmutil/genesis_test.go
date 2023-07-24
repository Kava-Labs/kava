package evmutil_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/kava-labs/kava/x/evmutil"
	"github.com/kava-labs/kava/x/evmutil/testutil"
	"github.com/kava-labs/kava/x/evmutil/types"
)

type genesisTestSuite struct {
	testutil.Suite
}

func (suite *genesisTestSuite) SetupTest() {
	suite.Suite.SetupTest()
}

func (s *genesisTestSuite) TestInitGenesis_SetAccounts() {
	gs := types.NewGenesisState(
		[]types.Account{
			{Address: s.Addrs[0], Balance: sdkmath.NewInt(100)},
		},
		types.DefaultParams(),
	)
	accounts := s.Keeper.GetAllAccounts(s.Ctx)
	s.Require().Len(accounts, 0)
	evmutil.InitGenesis(s.Ctx, s.Keeper, gs, s.AccountKeeper)
	accounts = s.Keeper.GetAllAccounts(s.Ctx)
	s.Require().Len(accounts, 1)
	account := s.Keeper.GetAccount(s.Ctx, s.Addrs[0])
	s.Require().Equal(account.Address, s.Addrs[0])
	s.Require().Equal(account.Balance, sdkmath.NewInt(100))
}

func (s *genesisTestSuite) TestInitGenesis_SetParams() {
	params := types.DefaultParams()
	conversionPair := types.ConversionPair{
		KavaERC20Address: testutil.MustNewInternalEVMAddressFromString("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2").Bytes(),
		Denom:            "weth",
	}
	params.EnabledConversionPairs = []types.ConversionPair{conversionPair}
	gs := types.NewGenesisState(
		[]types.Account{},
		params,
	)
	evmutil.InitGenesis(s.Ctx, s.Keeper, gs, s.AccountKeeper)
	params = s.Keeper.GetParams(s.Ctx)
	s.Require().Len(params.EnabledConversionPairs, 1)
	s.Require().Equal(conversionPair, params.EnabledConversionPairs[0])
}

func (s *genesisTestSuite) TestInitGenesis_ValidateFail() {
	gs := types.NewGenesisState(
		[]types.Account{
			{Address: s.Addrs[0], Balance: sdkmath.NewInt(-100)},
		},
		types.DefaultParams(),
	)
	s.Require().Panics(func() {
		evmutil.InitGenesis(s.Ctx, s.Keeper, gs, s.AccountKeeper)
	})
}

func (s *genesisTestSuite) TestInitGenesis_ModuleAccount() {
	gs := types.NewGenesisState(
		[]types.Account{},
		types.DefaultParams(),
	)
	s.Require().NotPanics(func() {
		evmutil.InitGenesis(s.Ctx, s.Keeper, gs, s.AccountKeeper)
	})
	// check for module account this way b/c GetModuleAccount creates if not existing.
	acc := s.AccountKeeper.GetAccount(s.Ctx, s.AccountKeeper.GetModuleAddress(types.ModuleName))
	s.Require().NotNil(acc)
	_, ok := acc.(authtypes.ModuleAccountI)
	s.Require().True(ok)
}

func (s *genesisTestSuite) TestExportGenesis() {
	accounts := []types.Account{
		{Address: s.Addrs[0], Balance: sdkmath.NewInt(10)},
		{Address: s.Addrs[1], Balance: sdkmath.NewInt(20)},
	}
	for _, account := range accounts {
		s.Keeper.SetAccount(s.Ctx, account)
	}
	params := types.DefaultParams()
	params.EnabledConversionPairs = []types.ConversionPair{
		{
			KavaERC20Address: testutil.MustNewInternalEVMAddressFromString("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2").Bytes(),
			Denom:            "weth"},
	}
	params.AllowedCosmosDenoms = []types.AllowedCosmosCoinERC20Token{
		{
			CosmosDenom: "hard",
			Name:        "Kava EVM HARD",
			Symbol:      "HARD",
			Decimals:    6,
		},
	}
	s.Keeper.SetParams(s.Ctx, params)
	gs := evmutil.ExportGenesis(s.Ctx, s.Keeper)
	s.Require().Equal(gs.Accounts, accounts)
	s.Require().Equal(params, gs.Params)
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(genesisTestSuite))
}
