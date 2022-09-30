package keeper_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/savings/keeper"
	"github.com/kava-labs/kava/x/savings/types"
)

// Test suite used for all keeper tests
type KeeperTestSuite struct {
	suite.Suite
	keeper keeper.Keeper
	app    app.TestApp
	ctx    sdk.Context
	addrs  []sdk.AccAddress
}

// The default state used by each test
func (suite *KeeperTestSuite) SetupTest() {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)

	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})
	tApp.InitializeFromGenesisStates()
	_, addrs := app.GeneratePrivKeyAddressPairs(1)
	keeper := tApp.GetSavingsKeeper()
	suite.app = tApp
	suite.ctx = ctx
	suite.keeper = keeper
	suite.addrs = addrs

	stakingParams := stakingtypes.DefaultParams()
	stakingParams.BondDenom = "ukava"
	suite.app.GetStakingKeeper().SetParams(suite.ctx, stakingParams)
}

func (suite *KeeperTestSuite) TestGetSetDeleteDeposit() {
	dep := types.NewDeposit(sdk.AccAddress("test"), sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(100))))

	_, f := suite.keeper.GetDeposit(suite.ctx, sdk.AccAddress("test"))
	suite.Require().False(f)

	suite.keeper.SetDeposit(suite.ctx, dep)

	testDeposit, f := suite.keeper.GetDeposit(suite.ctx, sdk.AccAddress("test"))
	suite.Require().True(f)
	suite.Require().Equal(dep, testDeposit)

	suite.Require().NotPanics(func() { suite.keeper.DeleteDeposit(suite.ctx, dep) })
	_, f = suite.keeper.GetDeposit(suite.ctx, sdk.AccAddress("test"))
	suite.Require().False(f)
}

func (suite *KeeperTestSuite) TestIterateDeposits() {
	for i := 0; i < 5; i++ {
		dep := types.NewDeposit(sdk.AccAddress("test"+fmt.Sprint(i)), sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(100))))
		suite.Require().NotPanics(func() { suite.keeper.SetDeposit(suite.ctx, dep) })
	}
	var deposits []types.Deposit
	suite.keeper.IterateDeposits(suite.ctx, func(d types.Deposit) bool {
		deposits = append(deposits, d)
		return false
	})
	suite.Require().Equal(5, len(deposits))
}

func (suite *KeeperTestSuite) getAccountCoins(acc authtypes.AccountI) sdk.Coins {
	bk := suite.app.GetBankKeeper()
	return bk.GetAllBalances(suite.ctx, acc.GetAddress())
}

func (suite *KeeperTestSuite) getAccount(addr sdk.AccAddress) authtypes.AccountI {
	ak := suite.app.GetAccountKeeper()
	return ak.GetAccount(suite.ctx, addr)
}

func (suite *KeeperTestSuite) getAccountAtCtx(addr sdk.AccAddress, ctx sdk.Context) authtypes.AccountI {
	ak := suite.app.GetAccountKeeper()
	return ak.GetAccount(ctx, addr)
}

func (suite *KeeperTestSuite) getModuleAccount(name string) authtypes.ModuleAccountI {
	ak := suite.app.GetAccountKeeper()
	return ak.GetModuleAccount(suite.ctx, name)
}

func (suite *KeeperTestSuite) getModuleAccountAtCtx(name string, ctx sdk.Context) authtypes.ModuleAccountI {
	ak := suite.app.GetAccountKeeper()
	return ak.GetModuleAccount(ctx, name)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

// CreateAccount creates a new account from the provided balance and address
func (suite *KeeperTestSuite) CreateAccountWithAddress(addr sdk.AccAddress, initialBalance sdk.Coins) authtypes.AccountI {
	ak := suite.app.GetAccountKeeper()

	acc := ak.NewAccountWithAddress(suite.ctx, addr)
	ak.SetAccount(suite.ctx, acc)

	err := simapp.FundAccount(suite.app.GetBankKeeper(), suite.ctx, acc.GetAddress(), initialBalance)
	suite.Require().NoError(err)

	return acc
}

// CreateVestingAccount creates a new vesting account. `vestingBalance` should be a fraction of `initialBalance`.
func (suite *KeeperTestSuite) CreateVestingAccountWithAddress(addr sdk.AccAddress, initialBalance sdk.Coins, vestingBalance sdk.Coins) authtypes.AccountI {
	if vestingBalance.IsAnyGT(initialBalance) {
		panic("vesting balance must be less than initial balance")
	}
	acc := suite.CreateAccountWithAddress(addr, initialBalance)
	bacc := acc.(*authtypes.BaseAccount)

	periods := vestingtypes.Periods{
		vestingtypes.Period{
			Length: 31556952,
			Amount: vestingBalance,
		},
	}
	vacc := vestingtypes.NewPeriodicVestingAccount(bacc, vestingBalance, suite.ctx.BlockTime().Unix(), periods)
	suite.app.GetAccountKeeper().SetAccount(suite.ctx, vacc)
	return vacc
}

func (suite *KeeperTestSuite) deliverMsgCreateValidator(ctx sdk.Context, address sdk.ValAddress, selfDelegation sdk.Coin) error {
	msg, err := stakingtypes.NewMsgCreateValidator(
		address,
		ed25519.GenPrivKey().PubKey(),
		selfDelegation,
		stakingtypes.Description{},
		stakingtypes.NewCommissionRates(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec()),
		sdk.NewInt(1e6),
	)
	if err != nil {
		return err
	}

	msgServer := stakingkeeper.NewMsgServerImpl(suite.app.GetStakingKeeper())
	_, err = msgServer.CreateValidator(sdk.WrapSDKContext(suite.ctx), msg)
	return err
}

// CreateNewUnbondedValidator creates a new validator in the staking module.
// New validators are unbonded until the end blocker is run.
func (suite *KeeperTestSuite) CreateNewUnbondedValidator(addr sdk.ValAddress, selfDelegation sdk.Int) stakingtypes.Validator {
	// Create a validator
	err := suite.deliverMsgCreateValidator(suite.ctx, addr, sdk.NewCoin("ukava", selfDelegation))
	suite.Require().NoError(err)

	// New validators are created in an unbonded state. Note if the end blocker is run later this validator could become bonded.

	validator, found := suite.app.GetStakingKeeper().GetValidator(suite.ctx, addr)
	suite.Require().True(found)
	return validator
}

// CreateDelegation delegates tokens to a validator.
func (suite *KeeperTestSuite) CreateDelegation(valAddr sdk.ValAddress, delegator sdk.AccAddress, amount sdk.Int) sdk.Dec {
	sk := suite.app.GetStakingKeeper()

	stakingDenom := sk.BondDenom(suite.ctx)
	msg := stakingtypes.NewMsgDelegate(
		delegator,
		valAddr,
		sdk.NewCoin(stakingDenom, amount),
	)

	msgServer := stakingkeeper.NewMsgServerImpl(sk)
	_, err := msgServer.Delegate(sdk.WrapSDKContext(suite.ctx), msg)
	suite.Require().NoError(err)

	del, found := sk.GetDelegation(suite.ctx, delegator, valAddr)
	suite.Require().True(found)
	return del.Shares
}
