package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/liquid/keeper"
)

// Test suite used for all keeper tests
type KeeperTestSuite struct {
	suite.Suite
	App           app.TestApp
	Ctx           sdk.Context
	Keeper        keeper.Keeper
	BankKeeper    bankkeeper.Keeper
	StakingKeeper stakingkeeper.Keeper
}

// The default state used by each test
func (suite *KeeperTestSuite) SetupTest() {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})

	tApp.InitializeFromGenesisStates()

	suite.App = tApp
	suite.Ctx = ctx
	suite.Keeper = tApp.GetLiquidKeeper()
	suite.StakingKeeper = tApp.GetStakingKeeper()
	suite.BankKeeper = tApp.GetBankKeeper()
}

// CreateAccount creates a new account (with a fixed address) from the provided balance.
func (suite *KeeperTestSuite) CreateAccount(initialBalance sdk.Coins) authtypes.AccountI {
	_, addrs := app.GeneratePrivKeyAddressPairs(1)

	return suite.CreateAccountWithAddress(addrs[0], initialBalance)
}

// CreateAccount creates a new account from the provided balance and address
func (suite *KeeperTestSuite) CreateAccountWithAddress(addr sdk.AccAddress, initialBalance sdk.Coins) authtypes.AccountI {
	ak := suite.App.GetAccountKeeper()

	acc := ak.NewAccountWithAddress(suite.Ctx, addr)
	ak.SetAccount(suite.Ctx, acc)

	err := simapp.FundAccount(suite.BankKeeper, suite.Ctx, acc.GetAddress(), initialBalance)
	suite.Require().NoError(err)

	return acc
}

// CreateVestingAccount creates a new vesting account from the provided balance and vesting balance
func (suite *KeeperTestSuite) CreateVestingAccount(initialBalance sdk.Coins, vestingBalance sdk.Coins) authtypes.AccountI {
	acc := suite.CreateAccount(initialBalance)
	bacc := acc.(*authtypes.BaseAccount)

	periods := vestingtypes.Periods{
		vestingtypes.Period{
			Length: 31556952,
			Amount: vestingBalance,
		},
	}
	vacc := vestingtypes.NewPeriodicVestingAccount(bacc, initialBalance, time.Now().Unix(), periods) // TODO is initialBalance correct for originalVesting?

	return vacc
}

// AccountBalanceEqual checks if an account has the specified coins.
func (suite *KeeperTestSuite) AccountBalanceEqual(addr sdk.AccAddress, coins sdk.Coins) {
	balance := suite.BankKeeper.GetAllBalances(suite.Ctx, addr)
	suite.Equalf(coins, balance, "expected account balance to equal coins %s, but got %s", coins, balance)
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

	msgServer := stakingkeeper.NewMsgServerImpl(suite.StakingKeeper)
	_, err = msgServer.CreateValidator(sdk.WrapSDKContext(suite.Ctx), msg)
	return err
}

// NewBondCoin creates a Coin with the current staking denom.
func (suite *KeeperTestSuite) NewBondCoin(amount sdk.Int) sdk.Coin {
	stakingDenom := suite.StakingKeeper.BondDenom(suite.Ctx)
	return sdk.NewCoin(stakingDenom, amount)
}

// NewBondCoins creates Coins with the current staking denom.
func (suite *KeeperTestSuite) NewBondCoins(amount sdk.Int) sdk.Coins {
	return sdk.NewCoins(suite.NewBondCoin(amount))
}

// CreateNewUnbondedValidator creates a new validator in the staking module.
// New validators are unbonded until the end blocker is run.
func (suite *KeeperTestSuite) CreateNewUnbondedValidator(addr sdk.ValAddress, selfDelegation sdk.Int) stakingtypes.Validator {
	// Create a validator
	err := suite.deliverMsgCreateValidator(suite.Ctx, addr, suite.NewBondCoin(selfDelegation))
	suite.Require().NoError(err)

	// New validators are created in an unbonded state. Note if the end blocker is run later this validator could become bonded.

	validator, found := suite.StakingKeeper.GetValidator(suite.Ctx, addr)
	suite.Require().True(found)
	return validator
}

// SlashValidator burns tokens delegated to a validator.
func (suite *KeeperTestSuite) SlashValidator(addr sdk.ValAddress, slashFraction sdk.Dec) {
	validator, found := suite.StakingKeeper.GetValidator(suite.Ctx, addr)
	suite.Require().True(found)
	consAddr, err := validator.GetConsAddr()
	suite.Require().NoError(err)

	// Assume infraction was at current height. Note unbonding delegations and redelegations are only slashed if created after
	// the infraction height so none will be slashed.
	infractionHeight := suite.Ctx.BlockHeight()

	power := suite.StakingKeeper.TokensToConsensusPower(suite.Ctx, validator.GetTokens())

	suite.StakingKeeper.Slash(suite.Ctx, consAddr, infractionHeight, power, slashFraction)
}

// CreateDelegation delegates tokens to a validator.
func (suite *KeeperTestSuite) CreateDelegation(valAddr sdk.ValAddress, delegator sdk.AccAddress, amount sdk.Int) sdk.Dec {
	stakingDenom := suite.StakingKeeper.BondDenom(suite.Ctx)
	msg := stakingtypes.NewMsgDelegate(
		delegator,
		valAddr,
		sdk.NewCoin(stakingDenom, amount),
	)

	msgServer := stakingkeeper.NewMsgServerImpl(suite.StakingKeeper)
	_, err := msgServer.Delegate(sdk.WrapSDKContext(suite.Ctx), msg)
	suite.Require().NoError(err)

	del, found := suite.StakingKeeper.GetDelegation(suite.Ctx, delegator, valAddr)
	suite.Require().True(found)
	return del.Shares
}

// CreateRedelegation undelegates tokens from one validator and delegates to another.
func (suite *KeeperTestSuite) CreateRedelegation(delegator sdk.AccAddress, fromValidator, toValidator sdk.ValAddress, amount sdk.Int) {
	stakingDenom := suite.StakingKeeper.BondDenom(suite.Ctx)
	msg := stakingtypes.NewMsgBeginRedelegate(
		delegator,
		fromValidator,
		toValidator,
		sdk.NewCoin(stakingDenom, amount),
	)

	msgServer := stakingkeeper.NewMsgServerImpl(suite.StakingKeeper)
	_, err := msgServer.BeginRedelegate(sdk.WrapSDKContext(suite.Ctx), msg)
	suite.Require().NoError(err)
}

// DelegationSharesEqual checks if a delegation has the specified shares.
// It expects delegations with zero shares to not be stored in state.
func (suite *KeeperTestSuite) DelegationSharesEqual(valAddr sdk.ValAddress, delegator sdk.AccAddress, shares sdk.Dec) bool {
	del, found := suite.StakingKeeper.GetDelegation(suite.Ctx, delegator, valAddr)

	if shares.IsZero() {
		return suite.Falsef(found, "expected delegator to not be found, got %s shares", del.Shares)
	} else {
		res := suite.True(found, "expected delegator to be found")
		return res && suite.Truef(shares.Equal(del.Shares), "expected %s delegator shares but got %s", shares, del.Shares)
	}
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
