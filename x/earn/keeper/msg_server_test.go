package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/kava-labs/kava/x/earn/keeper"
	"github.com/kava-labs/kava/x/earn/testutil"
	"github.com/kava-labs/kava/x/earn/types"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto"
)

var moduleAccAddress = sdk.AccAddress(crypto.AddressHash([]byte(types.ModuleAccountName)))

type msgServerTestSuite struct {
	testutil.Suite

	msgServer types.MsgServer
}

func (suite *msgServerTestSuite) SetupTest() {
	suite.Suite.SetupTest()
	suite.Keeper.SetParams(suite.Ctx, types.DefaultParams())

	suite.msgServer = keeper.NewMsgServerImpl(suite.Keeper)
}

func TestMsgServerTestSuite(t *testing.T) {
	suite.Run(t, new(msgServerTestSuite))
}

func (suite *msgServerTestSuite) TestDeposit() {
	vaultDenom := "usdx"
	suite.CreateVault(vaultDenom, types.STRATEGY_TYPE_HARD)

	startBalance := sdk.NewInt64Coin(vaultDenom, 1000)
	depositAmount := sdk.NewInt64Coin(vaultDenom, 100)

	acc := suite.CreateAccount(sdk.NewCoins(startBalance), 0)

	msg := types.NewMsgDeposit(acc.GetAddress().String(), depositAmount)
	_, err := suite.msgServer.Deposit(sdk.WrapSDKContext(suite.Ctx), msg)
	suite.Require().NoError(err)

	suite.AccountBalanceEqual(
		acc.GetAddress(),
		sdk.NewCoins(startBalance.Sub(depositAmount)),
	)

	// Bank: Send deposit Account -> Module account
	suite.EventsContains(
		suite.GetEvents(),
		sdk.NewEvent(
			banktypes.EventTypeTransfer,
			sdk.NewAttribute(banktypes.AttributeKeyRecipient, moduleAccAddress.String()),
			sdk.NewAttribute(banktypes.AttributeKeySender, acc.GetAddress().String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, depositAmount.String()),
		),
	)

	// Keeper Deposit()
	suite.EventsContains(
		suite.GetEvents(),
		sdk.NewEvent(
			types.EventTypeVaultDeposit,
			sdk.NewAttribute(types.AttributeKeyVaultDenom, depositAmount.Denom),
			sdk.NewAttribute(types.AttributeKeyDepositor, acc.GetAddress().String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, depositAmount.Amount.String()),
		),
	)

	// Msg server module
	suite.EventsContains(
		suite.GetEvents(),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeySender, acc.GetAddress().String()),
		),
	)
}

func (suite *msgServerTestSuite) TestWithdraw() {
	vaultDenom := "usdx"
	suite.CreateVault(vaultDenom, types.STRATEGY_TYPE_HARD)

	startBalance := sdk.NewInt64Coin(vaultDenom, 1000)
	depositAmount := sdk.NewInt64Coin(vaultDenom, 100)

	acc := suite.CreateAccount(sdk.NewCoins(startBalance), 0)

	msgDeposit := types.NewMsgDeposit(acc.GetAddress().String(), depositAmount)
	_, err := suite.msgServer.Deposit(sdk.WrapSDKContext(suite.Ctx), msgDeposit)
	suite.Require().NoError(err)

	// Withdraw all
	msgWithdraw := types.NewMsgWithdraw(acc.GetAddress().String(), depositAmount)
	_, err = suite.msgServer.Withdraw(sdk.WrapSDKContext(suite.Ctx), msgWithdraw)
	suite.Require().NoError(err)

	// Bank: Send deposit Account -> Module account
	suite.EventsContains(
		suite.GetEvents(),
		sdk.NewEvent(
			banktypes.EventTypeTransfer,
			// Direction opposite from Deposit()
			sdk.NewAttribute(banktypes.AttributeKeyRecipient, acc.GetAddress().String()),
			sdk.NewAttribute(banktypes.AttributeKeySender, moduleAccAddress.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, depositAmount.String()),
		),
	)

	// Keeper Withdraw()
	suite.EventsContains(
		suite.GetEvents(),
		sdk.NewEvent(
			types.EventTypeVaultWithdraw,
			sdk.NewAttribute(types.AttributeKeyVaultDenom, depositAmount.Denom),
			sdk.NewAttribute(types.AttributeKeyOwner, acc.GetAddress().String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, depositAmount.Amount.String()),
		),
	)

	// Msg server module
	suite.EventsContains(
		suite.GetEvents(),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeySender, acc.GetAddress().String()),
		),
	)
}
