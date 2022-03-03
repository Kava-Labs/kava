package evmutil_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

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
	gs := types.NewGenesisState([]types.Account{
		{Address: s.Addrs[0], Balance: sdk.NewInt(100)},
	})
	accounts := s.Keeper.GetAllAccounts(s.Ctx)
	s.Require().Len(accounts, 0)
	evmutil.InitGenesis(s.Ctx, s.Keeper, gs)
	accounts = s.Keeper.GetAllAccounts(s.Ctx)
	s.Require().Len(accounts, 1)
	account := s.Keeper.GetAccount(s.Ctx, s.Addrs[0])
	s.Require().Equal(account.Address, s.Addrs[0])
	s.Require().Equal(account.Balance, sdk.NewInt(100))
}

func (s *genesisTestSuite) TestInitGenesis_ValidateFail() {
	gs := types.NewGenesisState([]types.Account{
		{Address: s.Addrs[0], Balance: sdk.NewInt(-100)},
	})
	s.Require().Panics(func() {
		evmutil.InitGenesis(s.Ctx, s.Keeper, gs)
	})
}

func (s *genesisTestSuite) TestExportGenesis() {
	accounts := []types.Account{
		{Address: s.Addrs[0], Balance: sdk.NewInt(10)},
		{Address: s.Addrs[1], Balance: sdk.NewInt(20)},
	}
	for _, account := range accounts {
		s.Keeper.SetAccount(s.Ctx, account)
	}
	gs := evmutil.ExportGenesis(s.Ctx, s.Keeper)
	s.Require().Equal(gs.Accounts, accounts)
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(genesisTestSuite))
}
