package evmutil_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"
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
	evmutil.InitGenesis(s.Ctx, s.Keeper, gs)
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
	evmutil.InitGenesis(s.Ctx, s.Keeper, gs)
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
		evmutil.InitGenesis(s.Ctx, s.Keeper, gs)
	})
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
	s.Keeper.SetParams(s.Ctx, params)
	gs := evmutil.ExportGenesis(s.Ctx, s.Keeper)
	s.Require().Equal(gs.Accounts, accounts)
	s.Require().Equal(params, gs.Params)
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(genesisTestSuite))
}
