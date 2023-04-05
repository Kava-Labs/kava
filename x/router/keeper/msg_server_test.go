package keeper_test

import (
	"fmt"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
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

func (suite *msgServerTestSuite) TestDelegateMintDeposit_Events() {
	user, valAddr, balance := suite.setupValidator()
	suite.setupEarnForDeposits(valAddr)

	msg := types.NewMsgDelegateMintDeposit(
		user,
		valAddr,
		suite.NewBondCoin(balance),
	)
	_, err := suite.msgServer.DelegateMintDeposit(sdk.WrapSDKContext(suite.Ctx), msg)
	suite.Require().NoError(err)

	suite.EventsContains(suite.Ctx.EventManager().Events(),
		sdk.NewEvent(sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeySender, user.String()),
		),
	)
	expectedShares := sdk.NewDecFromInt(msg.Amount.Amount) // no slashes so shares equal staked tokens
	suite.EventsContains(suite.Ctx.EventManager().Events(),
		sdk.NewEvent(
			stakingtypes.EventTypeDelegate,
			sdk.NewAttribute(stakingtypes.AttributeKeyValidator, msg.Validator),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.String()),
			sdk.NewAttribute(stakingtypes.AttributeKeyNewShares, expectedShares.String()),
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

func (suite *msgServerTestSuite) TestWithdrawBurnUndelegate_Events() {
	user, valAddr, delegated := suite.setupDerivatives()
	// clear events from setup
	suite.Ctx = suite.Ctx.WithEventManager(sdk.NewEventManager())

	msg := types.NewMsgWithdrawBurnUndelegate(
		user,
		valAddr,
		// in this setup where the validator is not slashed, the derivative amount is equal to the staked balance
		suite.NewBondCoin(delegated.Amount),
	)
	_, err := suite.msgServer.WithdrawBurnUndelegate(sdk.WrapSDKContext(suite.Ctx), msg)
	suite.Require().NoError(err)

	suite.EventsContains(suite.Ctx.EventManager().Events(),
		sdk.NewEvent(sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeySender, user.String()),
		),
	)
	unbondingTime := suite.StakingKeeper.UnbondingTime(suite.Ctx)
	completionTime := suite.Ctx.BlockTime().Add(unbondingTime)
	suite.EventsContains(suite.Ctx.EventManager().Events(),
		sdk.NewEvent(
			stakingtypes.EventTypeUnbond,
			sdk.NewAttribute(stakingtypes.AttributeKeyValidator, msg.Validator),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.String()),
			sdk.NewAttribute(stakingtypes.AttributeKeyCompletionTime, completionTime.Format(time.RFC3339)),
		),
	)
}

func (suite *msgServerTestSuite) TestMintDepositAndWithdrawBurn_TransferEntireBalance() {
	_, addrs := app.GeneratePrivKeyAddressPairs(5)
	valAccAddr, user := addrs[0], addrs[1]
	valAddr := sdk.ValAddress(valAccAddr)

	derivativeDenom := suite.setupEarnForDeposits(valAddr)

	suite.CreateAccountWithAddress(valAccAddr, suite.NewBondCoins(sdkmath.NewInt(1e9)))
	vesting := sdkmath.NewInt(1000)
	suite.CreateVestingAccountWithAddress(user,
		suite.NewBondCoins(sdkmath.NewInt(1e9).Add(vesting)),
		suite.NewBondCoins(vesting),
	)

	// Create a slashed validator, where the delegator owns fractional tokens.
	suite.CreateNewUnbondedValidator(valAddr, sdkmath.NewInt(1e9))
	suite.CreateDelegation(valAddr, user, sdkmath.NewInt(1e9))
	staking.EndBlocker(suite.Ctx, suite.StakingKeeper)
	suite.SlashValidator(valAddr, sdk.MustNewDecFromStr("0.666666666666666667"))

	// Query the full staked balance and convert it all to derivatives
	// user technically 333_333_333.333333333333333333 staked tokens without rounding
	delegation := suite.QueryStaking_Delegation(valAddr, user)
	suite.Equal(sdkmath.NewInt(333_333_333), delegation.Balance.Amount)

	msgDeposit := types.NewMsgMintDeposit(
		user,
		valAddr,
		delegation.Balance,
	)
	_, err := suite.msgServer.MintDeposit(sdk.WrapSDKContext(suite.Ctx), msgDeposit)
	suite.Require().NoError(err)

	// There should be no extractable balance left in delegation
	suite.DelegationBalanceLessThan(valAddr, user, sdkmath.NewInt(1))
	// All derivative coins should be deposited to earn
	suite.AccountBalanceOfEqual(user, derivativeDenom, sdk.ZeroInt())
	// Earn vault has all minted derivatives
	suite.VaultAccountValueEqual(user, sdk.NewInt64Coin(derivativeDenom, 999_999_998)) // 2 lost in conversion

	// Query the full kava balance of the earn deposit and convert all to a delegation
	deposit := suite.QueryEarn_VaultValue(user, "bkava")
	suite.Equal(suite.NewBondCoins(sdkmath.NewInt(333_333_332)), deposit.Value) // 1 lost due to lost shares

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
	// The user should get back most of their original deposited balance
	suite.DelegationBalanceInDeltaBelow(valAddr, user, sdkmath.NewInt(333_333_332), sdkmath.NewInt(2))
}

func (suite *msgServerTestSuite) TestDelegateMintDepositAndWithdrawBurnUndelegate_TransferEntireBalance() {
	_, addrs := app.GeneratePrivKeyAddressPairs(5)
	valAccAddr, user := addrs[0], addrs[1]
	valAddr := sdk.ValAddress(valAccAddr)

	derivativeDenom := suite.setupEarnForDeposits(valAddr)

	valBalance := sdkmath.NewInt(1e9)
	suite.CreateAccountWithAddress(valAccAddr, suite.NewBondCoins(valBalance))

	// Create a slashed validator, where a future delegator will own fractional tokens.
	suite.CreateNewUnbondedValidator(valAddr, valBalance)
	staking.EndBlocker(suite.Ctx, suite.StakingKeeper)
	suite.SlashValidator(valAddr, sdk.MustNewDecFromStr("0.4")) // tokens remaining 600_000_000

	userBalance := sdkmath.NewInt(100e6)
	vesting := sdkmath.NewInt(1000)
	suite.CreateVestingAccountWithAddress(user,
		suite.NewBondCoins(userBalance.Add(vesting)),
		suite.NewBondCoins(vesting),
	)

	// Query the full vested balance and convert it all to derivatives
	balance := suite.QueryBank_SpendableBalance(user)
	suite.Equal(suite.NewBondCoins(userBalance), balance)

	// When delegation is created it will have 166_666_666.666666666666666666 shares
	// newShares = validatorShares * newTokens/validatorTokens, truncated to 18 decimals
	msgDeposit := types.NewMsgDelegateMintDeposit(
		user,
		valAddr,
		balance[0],
	)
	_, err := suite.msgServer.DelegateMintDeposit(sdk.WrapSDKContext(suite.Ctx), msgDeposit)
	suite.Require().NoError(err)

	// All spendable balance should be withdrawn
	suite.AccountSpendableBalanceEqual(user, nil)
	// Since shares are newly created, the exact amount should be converted to derivatives, leaving none behind
	suite.DelegationSharesEqual(valAddr, user, sdk.ZeroDec())
	// All derivative coins should be deposited to earn
	suite.AccountBalanceOfEqual(user, derivativeDenom, sdk.ZeroInt())

	suite.VaultAccountValueEqual(user, sdk.NewInt64Coin(derivativeDenom, 166_666_666))

	// Query the full kava balance of the earn deposit and convert all to a delegation
	deposit := suite.QueryEarn_VaultValue(user, "bkava")
	suite.Equal(suite.NewBondCoins(sdkmath.NewInt(99_999_999)), deposit.Value) // 1 lost due to truncating shares to derivatives

	msgWithdraw := types.NewMsgWithdrawBurnUndelegate(
		user,
		valAddr,
		deposit.Value[0],
	)
	_, err = suite.msgServer.WithdrawBurnUndelegate(sdk.WrapSDKContext(suite.Ctx), msgWithdraw)
	suite.Require().NoError(err)

	// There should be no earn deposit left (earn removes dust amounts)
	suite.VaultAccountSharesEqual(user, nil)
	// All derivative coins should be converted to a delegation
	suite.AccountBalanceOfEqual(user, derivativeDenom, sdk.ZeroInt())
	// There should be zero shares left because undelegate removes all created by burn.
	suite.AccountBalanceOfEqual(user, derivativeDenom, sdk.ZeroInt())
	// User should have most of their original balance back in an unbonding delegation
	suite.UnbondingDelegationInDeltaBelow(valAddr, user, userBalance, sdkmath.NewInt(2))
}

func (suite *msgServerTestSuite) setupValidator() (sdk.AccAddress, sdk.ValAddress, sdkmath.Int) {
	_, addrs := app.GeneratePrivKeyAddressPairs(5)
	valAccAddr, user := addrs[0], addrs[1]
	valAddr := sdk.ValAddress(valAccAddr)

	balance := sdkmath.NewInt(1e9)

	suite.CreateAccountWithAddress(valAccAddr, suite.NewBondCoins(balance))
	suite.CreateAccountWithAddress(user, suite.NewBondCoins(balance))

	suite.CreateNewUnbondedValidator(valAddr, balance)
	staking.EndBlocker(suite.Ctx, suite.StakingKeeper)
	return user, valAddr, balance
}

func (suite *msgServerTestSuite) setupValidatorAndDelegation() (sdk.AccAddress, sdk.ValAddress, sdkmath.Int) {
	_, addrs := app.GeneratePrivKeyAddressPairs(5)
	valAccAddr, user := addrs[0], addrs[1]
	valAddr := sdk.ValAddress(valAccAddr)

	balance := sdkmath.NewInt(1e9)

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
