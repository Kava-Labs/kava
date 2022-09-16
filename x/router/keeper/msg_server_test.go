package keeper_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/app"
	earntypes "github.com/kava-labs/kava/x/earn/types"
	"github.com/kava-labs/kava/x/router/keeper"
	"github.com/kava-labs/kava/x/router/testutil"
	"github.com/kava-labs/kava/x/router/types"
)

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

func (suite *msgServerTestSuite) TestMintDeposit_Events() {
	user, valAddr, delegation := suite.setupValidatorAndDelegation()
	suite.setupEarnForDeposits(valAddr)

	msg := types.NewMsgMintDeposit(
		user,
		valAddr,
		suite.NewBondCoin(delegation),
	)
	_, err := suite.msgServer.MintDeposit(sdk.WrapSDKContext(suite.Ctx), msg)
	suite.Require().NoError(err)

	suite.EventsContains(suite.Ctx.EventManager().Events(),
		sdk.NewEvent(sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeySender, user.String()),
		),
	)
}


func (suite *msgServerTestSuite) TestWithdrawBurn_Events() {
	user, valAddr, delegated := suite.setupDerivatives()
	// clear events from setup
	suite.Ctx = suite.Ctx.WithEventManager(sdk.NewEventManager())

	msg := types.NewMsgWithdrawBurn(
		user,
		valAddr,
		// in this setup where the validator is not slashed, the derivative amount is equal to the staked balance
		suite.NewBondCoin(delegated.Amount),
	)
	_, err := suite.msgServer.WithdrawBurn(sdk.WrapSDKContext(suite.Ctx), msg)
	suite.Require().NoError(err)

	suite.EventsContains(suite.Ctx.EventManager().Events(),
		sdk.NewEvent(sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeySender, user.String()),
		),
	)
}


func (suite *msgServerTestSuite) TestMintDepositAndWithdrawBurn_TransferEntireBalance() {
	_, addrs := app.GeneratePrivKeyAddressPairs(5)
	valAccAddr, user := addrs[0], addrs[1]
	valAddr := sdk.ValAddress(valAccAddr)

	suite.CreateAccountWithAddress(valAccAddr, suite.NewBondCoins(sdk.NewInt(1e9)))
	suite.CreateAccountWithAddress(user, suite.NewBondCoins(sdk.NewInt(1e9)))

	derivativeDenom := suite.setupEarnForDeposits(valAddr)

	// Create a slashed validator, where the delegator owns fractional tokens.
	suite.CreateNewUnbondedValidator(valAddr, sdk.NewInt(1e9))
	suite.CreateDelegation(valAddr, user, sdk.NewInt(1e9))
	staking.EndBlocker(suite.Ctx, suite.StakingKeeper)
	suite.SlashValidator(valAddr, sdk.MustNewDecFromStr("0.666666666666666667"))

	// Query the full staked balance and convert it all to derivatives
	delegation := suite.QueryStaking_Delegation(valAddr, user)
	suite.Equal(sdk.NewInt(333_333_333), delegation.Balance.Amount)

	msgDeposit := types.NewMsgMintDeposit(
		user,
		valAddr,
		delegation.Balance,
	)
	_, err := suite.msgServer.MintDeposit(sdk.WrapSDKContext(suite.Ctx), msgDeposit)
	suite.Require().NoError(err)

	// There should be no extractable balance left in delegation
	suite.DelegationBalanceLessThan(valAddr, user, sdk.NewInt(1))
	// All derivative coins should be deposited to earn
	suite.AccountBalanceOfEqual(user, derivativeDenom, sdk.ZeroInt())

	suite.VaultAccountValueEqual(user, sdk.NewInt64Coin(derivativeDenom, 999_999_998))

	// Query the full kava balance of the earn deposit and convert all to a delegation
	// TODO query kava amount
	// deposit := suite.QueryEarn_VaultValue(user, "bkava")
	deposit := earntypes.DepositResponse{Value: sdk.NewCoins(sdk.NewInt64Coin("stake", 333_333_332))}
	suite.Equal(suite.NewBondCoins(sdk.NewInt(333_333_332)), deposit.Value)

	msgWithdraw := types.NewMsgWithdrawBurn(
		user,
		valAddr,
		deposit.Value[0],
	)
	_, err = suite.msgServer.WithdrawBurn(sdk.WrapSDKContext(suite.Ctx), msgWithdraw)
	suite.Require().NoError(err)

	// There should be no earn deposit left (earn removes dust amounts)
	suite.VaultAccountSharesEqual(user, nil)
	// All derivative coins should be converted to a delegation
	suite.AccountBalanceOfEqual(user, derivativeDenom, sdk.ZeroInt())
	// The user should get back most of their original delegation
	suite.DelegationBalanceInDeltaBelow(valAddr, user, sdk.NewInt(333_333_332), sdk.NewInt(2))
}


func (suite *msgServerTestSuite) setupValidator() (sdk.AccAddress, sdk.ValAddress, sdk.Int) {
	_, addrs := app.GeneratePrivKeyAddressPairs(5)
	valAccAddr, user := addrs[0], addrs[1]
	valAddr := sdk.ValAddress(valAccAddr)

	balance := sdk.NewInt(1e9)

	suite.CreateAccountWithAddress(valAccAddr, suite.NewBondCoins(balance))
	suite.CreateAccountWithAddress(user, suite.NewBondCoins(balance))

	suite.CreateNewUnbondedValidator(valAddr, balance)
	staking.EndBlocker(suite.Ctx, suite.StakingKeeper)
	return user, valAddr, balance
}

func (suite *msgServerTestSuite) setupValidatorAndDelegation() (sdk.AccAddress, sdk.ValAddress, sdk.Int) {
	_, addrs := app.GeneratePrivKeyAddressPairs(5)
	valAccAddr, user := addrs[0], addrs[1]
	valAddr := sdk.ValAddress(valAccAddr)

	balance := sdk.NewInt(1e9)

	suite.CreateAccountWithAddress(valAccAddr, suite.NewBondCoins(balance))
	suite.CreateAccountWithAddress(user, suite.NewBondCoins(balance))

	suite.CreateNewUnbondedValidator(valAddr, balance)
	suite.CreateDelegation(valAddr, user, balance)
	staking.EndBlocker(suite.Ctx, suite.StakingKeeper)
	return user, valAddr, balance
}

func (suite *msgServerTestSuite) setupEarnForDeposits(valAddr sdk.ValAddress) string {
	suite.CreateVault("bkava", earntypes.StrategyTypes{earntypes.STRATEGY_TYPE_SAVINGS}, false, nil)
	derivativeDenom := fmt.Sprintf("bkava-%s", valAddr)
	suite.SetSavingsSupportedDenoms([]string{derivativeDenom})
	return derivativeDenom
}

func (suite *msgServerTestSuite) setupDerivatives() (sdk.AccAddress, sdk.ValAddress, sdk.Coin) {
	user, valAddr, delegation := suite.setupValidatorAndDelegation()
	suite.setupEarnForDeposits(valAddr)

	msg := types.NewMsgMintDeposit(
		user,
		valAddr,
		suite.NewBondCoin(delegation),
	)
	_, err := suite.msgServer.MintDeposit(sdk.WrapSDKContext(suite.Ctx), msg)
	suite.Require().NoError(err)

	derivativeDenom := fmt.Sprintf("bkava-%s", valAddr)
	derivatives, err := suite.EarnKeeper.GetVaultAccountValue(suite.Ctx, derivativeDenom, user)
	suite.Require().NoError(err)

	return user, valAddr, derivatives
}
