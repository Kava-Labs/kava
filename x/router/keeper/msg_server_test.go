package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto"

	"github.com/kava-labs/kava/x/earn/keeper"
	"github.com/kava-labs/kava/x/earn/testutil"
	"github.com/kava-labs/kava/x/earn/types"
)

var moduleAccAddress = sdk.AccAddress(crypto.AddressHash([]byte(types.ModuleAccountName)))

type msgServerTestSuite struct {
	testutil.Suite

	msgServer types.MsgServer
}

func (suite *msgServerTestSuite) SetupTest() {
	suite.Suite.SetupTest()

	suite.msgServer = keeper.NewMsgServerImpl(suite.Keeper)
}

func TestMsgServerTestSuite(t *testing.T) {
	suite.Run(t, new(msgServerTestSuite))
}

/*
run tests for full balances
*/

func (suite *msgServerTestSuite) TestMintDeposit() {
	// setup validator, user
	// setup earn module (params, vault)
	// setup validator

	// run msg (full balance)

	// check no dust

	// vaultDenom := "bkava"
	// suite.CreateVault(vaultDenom, types.StrategyTypes{types.STRATEGY_TYPE_SAVINGS}, false, nil)

	// startBalance := sdk.NewInt64Coin(vaultDenom, 1000)
	// depositAmount := sdk.NewInt64Coin(vaultDenom, 100)

	// acc := suite.CreateAccount(sdk.NewCoins(startBalance), 0)

	// msgMintDeposit := types.MsgMintDeposit{acc.GetAddress().String(), depositAmount, types.STRATEGY_TYPE_SAVINGS}

	// _, err := suite.msgServer.MintDeposit(sdk.WrapSDKContext(suite.Ctx), &msgMintDeposit)
	// suite.Require().NoError(err)
}
