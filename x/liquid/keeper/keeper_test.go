package keeper_test

import (
	"fmt"
	"reflect"
	"testing"

	sdkmath "cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtime "github.com/cometbft/cometbft/types/time"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/suite"

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
	StakingKeeper *stakingkeeper.Keeper
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
func (suite *KeeperTestSuite) CreateAccount(initialBalance sdk.Coins, index int) authtypes.AccountI {
	_, addrs := app.GeneratePrivKeyAddressPairs(index + 1)

	return suite.CreateAccountWithAddress(addrs[index], initialBalance)
}

// CreateAccount creates a new account from the provided balance and address
func (suite *KeeperTestSuite) CreateAccountWithAddress(addr sdk.AccAddress, initialBalance sdk.Coins) authtypes.AccountI {
	ak := suite.App.GetAccountKeeper()

	acc := ak.NewAccountWithAddress(suite.Ctx, addr)
	ak.SetAccount(suite.Ctx, acc)

	err := suite.App.FundAccount(suite.Ctx, acc.GetAddress(), initialBalance)
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
	vacc := vestingtypes.NewPeriodicVestingAccount(bacc, vestingBalance, suite.Ctx.BlockTime().Unix(), periods)
	suite.App.GetAccountKeeper().SetAccount(suite.Ctx, vacc)
	return vacc
}

// AddCoinsToModule adds coins to the a module account, creating it if it doesn't exist.
func (suite *KeeperTestSuite) AddCoinsToModule(module string, amount sdk.Coins) {
	err := suite.App.FundModuleAccount(suite.Ctx, module, amount)
	suite.Require().NoError(err)
}

// AccountBalanceEqual checks if an account has the specified coins.
func (suite *KeeperTestSuite) AccountBalanceEqual(addr sdk.AccAddress, coins sdk.Coins) {
	balance := suite.BankKeeper.GetAllBalances(suite.Ctx, addr)
	suite.Truef(coins.IsEqual(balance), "expected account balance to equal coins %s, but got %s", coins, balance)
}

func (suite *KeeperTestSuite) deliverMsgCreateValidator(ctx sdk.Context, address sdk.ValAddress, selfDelegation sdk.Coin) error {
	msg, err := stakingtypes.NewMsgCreateValidator(
		address,
		ed25519.GenPrivKey().PubKey(),
		selfDelegation,
		stakingtypes.Description{},
		stakingtypes.NewCommissionRates(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec()),
		sdkmath.NewInt(1e6),
	)
	if err != nil {
		return err
	}

	msgServer := stakingkeeper.NewMsgServerImpl(suite.StakingKeeper)
	_, err = msgServer.CreateValidator(sdk.WrapSDKContext(suite.Ctx), msg)
	return err
}

// NewBondCoin creates a Coin with the current staking denom.
func (suite *KeeperTestSuite) NewBondCoin(amount sdkmath.Int) sdk.Coin {
	stakingDenom := suite.StakingKeeper.BondDenom(suite.Ctx)
	return sdk.NewCoin(stakingDenom, amount)
}

// NewBondCoins creates Coins with the current staking denom.
func (suite *KeeperTestSuite) NewBondCoins(amount sdkmath.Int) sdk.Coins {
	return sdk.NewCoins(suite.NewBondCoin(amount))
}

// CreateNewUnbondedValidator creates a new validator in the staking module.
// New validators are unbonded until the end blocker is run.
func (suite *KeeperTestSuite) CreateNewUnbondedValidator(addr sdk.ValAddress, selfDelegation sdkmath.Int) stakingtypes.Validator {
	// Create a validator
	err := suite.deliverMsgCreateValidator(suite.Ctx, addr, suite.NewBondCoin(selfDelegation))
	suite.Require().NoError(err)

	// New validators are created in an unbonded state. Note if the end blocker is run later this validator could become bonded.

	validator, found := suite.StakingKeeper.GetValidator(suite.Ctx, addr)
	suite.Require().True(found)
	return validator
}

// SlashValidator burns tokens staked in a validator. new_tokens = old_tokens * (1-slashFraction)
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
func (suite *KeeperTestSuite) CreateDelegation(valAddr sdk.ValAddress, delegator sdk.AccAddress, amount sdkmath.Int) sdk.Dec {
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
func (suite *KeeperTestSuite) CreateRedelegation(delegator sdk.AccAddress, fromValidator, toValidator sdk.ValAddress, amount sdkmath.Int) {
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

// CreateUnbondingDelegation undelegates tokens from a validator.
func (suite *KeeperTestSuite) CreateUnbondingDelegation(delegator sdk.AccAddress, validator sdk.ValAddress, amount sdkmath.Int) {
	stakingDenom := suite.StakingKeeper.BondDenom(suite.Ctx)
	msg := stakingtypes.NewMsgUndelegate(
		delegator,
		validator,
		sdk.NewCoin(stakingDenom, amount),
	)
	msgServer := stakingkeeper.NewMsgServerImpl(suite.StakingKeeper)
	_, err := msgServer.Undelegate(sdk.WrapSDKContext(suite.Ctx), msg)
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

// EventsContains asserts that the expected event is in the provided events
func (suite *KeeperTestSuite) EventsContains(events sdk.Events, expectedEvent sdk.Event) {
	foundMatch := false
	for _, event := range events {
		if event.Type == expectedEvent.Type {
			if reflect.DeepEqual(attrsToMap(expectedEvent.Attributes), attrsToMap(event.Attributes)) {
				foundMatch = true
			}
		}
	}

	suite.True(foundMatch, fmt.Sprintf("event of type %s not found or did not match", expectedEvent.Type))
}

// EventsDoNotContainType asserts that the provided events do contain an event of a certain type.
func (suite *KeeperTestSuite) EventsDoNotContainType(events sdk.Events, eventType string) {
	for _, event := range events {
		suite.Falsef(event.Type == eventType, "found unexpected event %s", eventType)
	}
}

func attrsToMap(attrs []abci.EventAttribute) []sdk.Attribute {
	out := []sdk.Attribute{}

	for _, attr := range attrs {
		out = append(out, sdk.NewAttribute(string(attr.Key), string(attr.Value)))
	}

	return out
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
