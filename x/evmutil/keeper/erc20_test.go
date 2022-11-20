package keeper_test

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/x/evmutil/testutil"
	"github.com/kava-labs/kava/x/evmutil/types"
)

type ERC20TestSuite struct {
	testutil.Suite

	contractAddr types.InternalEVMAddress
}

func TestERC20TestSuite(t *testing.T) {
	suite.Run(t, new(ERC20TestSuite))
}

func (suite *ERC20TestSuite) SetupTest() {
	suite.Suite.SetupTest()
	suite.contractAddr = suite.DeployERC20()
}

func (suite *ERC20TestSuite) TestERC20QueryBalanceOf_Empty() {
	bal, err := suite.Keeper.QueryERC20BalanceOf(
		suite.Ctx,
		suite.contractAddr,
		suite.Key1Addr,
	)
	suite.Require().NoError(err)
	suite.Require().True(bal.Cmp(big.NewInt(0)) == 0, "balance should be 0")
}

func (suite *ERC20TestSuite) TestERC20QueryBalanceOf_NonEmpty() {
	// Mint some tokens for the address
	err := suite.Keeper.MintERC20(
		suite.Ctx,
		suite.contractAddr,
		suite.Key1Addr,
		big.NewInt(10),
	)
	suite.Require().NoError(err)

	bal, err := suite.Keeper.QueryERC20BalanceOf(
		suite.Ctx,
		suite.contractAddr,
		suite.Key1Addr,
	)
	suite.Require().NoError(err)
	suite.Require().Equal(big.NewInt(10), bal, "balance should be 10")
}

func (suite *ERC20TestSuite) TestERC20Mint() {
	contractAddr := suite.DeployERC20()

	// We can't test mint by module account like the Unauthorized test as we
	// cannot sign as the module account. Instead, we test the keeper method for
	// minting.

	receiver := common.BytesToAddress(suite.Key2.PubKey().Address())
	amount := big.NewInt(1234)
	err := suite.App.GetEvmutilKeeper().MintERC20(suite.Ctx, contractAddr, types.NewInternalEVMAddress(receiver), amount)
	suite.Require().NoError(err)

	// Query ERC20.balanceOf()
	addr := common.BytesToAddress(suite.Key1.PubKey().Address())
	res, err := suite.QueryContract(
		types.CustomERC20Contract.ABI,
		addr,
		suite.Key1,
		contractAddr,
		"balanceOf",
		receiver,
	)
	suite.Require().NoError(err)
	suite.Require().Len(res, 1)

	balance, ok := res[0].(*big.Int)
	suite.Require().True(ok, "balanceOf should respond with *big.Int")
	suite.Require().Equal(big.NewInt(1234), balance)
}
