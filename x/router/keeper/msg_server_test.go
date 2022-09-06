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
	"github.com/kava-labs/kava/x/router/types"
)

type msgServerTestSuite struct {
	KeeperTestSuite // TODO use testutil like swap/earn?

	msgServer types.MsgServer
}

func (suite *msgServerTestSuite) SetupTest() {
	suite.KeeperTestSuite.SetupTest()

	suite.msgServer = keeper.NewMsgServerImpl(suite.Keeper)
}

func TestMsgServerTestSuite(t *testing.T) {
	suite.Run(t, new(msgServerTestSuite))
}

func (suite *msgServerTestSuite) TestMintDepositAndWithdrawBurn() {
	_, addrs := app.GeneratePrivKeyAddressPairs(5)
	valAccAddr, user := addrs[0], addrs[1]
	valAddr := sdk.ValAddress(valAccAddr)

	suite.CreateAccountWithAddress(valAccAddr, suite.NewBondCoins(sdk.NewInt(1e9)))
	suite.CreateAccountWithAddress(user, suite.NewBondCoins(sdk.NewInt(1e9)))

	suite.CreateVault("bkava", earntypes.StrategyTypes{earntypes.STRATEGY_TYPE_SAVINGS}, false, nil)
	derivativeDenom := fmt.Sprintf("bkava-%s", valAddr)
	suite.SetSavingsSupportedDenoms([]string{derivativeDenom})

	suite.CreateNewUnbondedValidator(valAddr, sdk.NewInt(1e9))
	suite.CreateDelegation(valAddr, user, sdk.NewInt(1e9))
	staking.EndBlocker(suite.Ctx, suite.StakingKeeper)

	// run msg (full balance)
	msgDeposit := &types.MsgMintDeposit{
		user.String(),
		valAddr.String(),
		suite.NewBondCoin(sdk.NewInt(1e9)),
		earntypes.STRATEGY_TYPE_SAVINGS,
	}
	_, err := suite.msgServer.MintDeposit(sdk.WrapSDKContext(suite.Ctx), msgDeposit)
	suite.Require().NoError(err)

	// check no dust
	suite.AccountBalanceEqual(user, sdk.Coins{})
	suite.DelegationSharesEqual(valAddr, user, sdk.ZeroDec())

	suite.VaultAccountValueEqual(user, sdk.NewInt64Coin(derivativeDenom, 1e9))

	msgWithdraw := &types.MsgWithdrawBurn{
		user.String(),
		valAddr.String(),
		suite.NewBondCoin(sdk.NewInt(1e9)),
		earntypes.STRATEGY_TYPE_SAVINGS,
	}
	_, err = suite.msgServer.WithdrawBurn(sdk.WrapSDKContext(suite.Ctx), msgWithdraw)
	suite.Require().NoError(err)

	// check no dust
	suite.VaultAccountSharesEqual(user, nil)
	suite.AccountBalanceEqual(user, sdk.Coins{})

	suite.DelegationSharesEqual(valAddr, user, sdk.NewDec(1e9))
}

func (suite *msgServerTestSuite) TestDelegateMintDepositAndWithdrawBurnUndelegate() {
	_, addrs := app.GeneratePrivKeyAddressPairs(5)
	valAccAddr, user := addrs[0], addrs[1]
	valAddr := sdk.ValAddress(valAccAddr)

	suite.CreateAccountWithAddress(valAccAddr, suite.NewBondCoins(sdk.NewInt(1e9)))
	suite.CreateAccountWithAddress(user, suite.NewBondCoins(sdk.NewInt(1e9)))

	suite.CreateVault("bkava", earntypes.StrategyTypes{earntypes.STRATEGY_TYPE_SAVINGS}, false, nil)
	derivativeDenom := fmt.Sprintf("bkava-%s", valAddr)
	suite.SetSavingsSupportedDenoms([]string{derivativeDenom})

	suite.CreateNewUnbondedValidator(valAddr, sdk.NewInt(1e9))
	staking.EndBlocker(suite.Ctx, suite.StakingKeeper)

	// run msg (full balance)
	msg := &types.MsgDelegateMintDeposit{
		user.String(),
		valAddr.String(),
		suite.NewBondCoin(sdk.NewInt(1e9)),
		earntypes.STRATEGY_TYPE_SAVINGS,
	}
	_, err := suite.msgServer.DelegateMintDeposit(sdk.WrapSDKContext(suite.Ctx), msg)
	suite.Require().NoError(err)

	// check no dust
	suite.AccountBalanceEqual(user, sdk.Coins{})
	suite.DelegationSharesEqual(valAddr, user, sdk.ZeroDec())

	suite.VaultAccountValueEqual(user, sdk.NewInt64Coin(derivativeDenom, 1e9))

	msgWithdraw := &types.MsgWithdrawBurnUndelegate{
		user.String(),
		valAddr.String(),
		suite.NewBondCoin(sdk.NewInt(1e9)),
		earntypes.STRATEGY_TYPE_SAVINGS,
	}
	_, err = suite.msgServer.WithdrawBurnUndelegate(sdk.WrapSDKContext(suite.Ctx), msgWithdraw)
	suite.Require().NoError(err)

	suite.VaultAccountSharesEqual(user, nil)
	suite.AccountBalanceEqual(user, sdk.Coins{})

	suite.DelegationSharesEqual(valAddr, user, sdk.ZeroDec())
	ubd, found := suite.StakingKeeper.GetUnbondingDelegation(suite.Ctx, user, valAddr)
	suite.True(found, "expected unbonding delegation to exist")
	suite.Equal(ubd.Entries[0].Balance, sdk.NewInt(1e9))
}
