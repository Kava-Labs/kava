package testutil

import (
	"fmt"
	"reflect"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	earnkeeper "github.com/kava-labs/kava/x/earn/keeper"
	earntypes "github.com/kava-labs/kava/x/earn/types"
	"github.com/kava-labs/kava/x/router/keeper"
	savingstypes "github.com/kava-labs/kava/x/savings/types"
)

// Test suite used for all keeper tests
type Suite struct {
	suite.Suite
	App           app.TestApp
	Ctx           sdk.Context
	Keeper        keeper.Keeper
	BankKeeper    bankkeeper.Keeper
	StakingKeeper stakingkeeper.Keeper
	EarnKeeper    earnkeeper.Keeper
}

// The default state used by each test
func (suite *Suite) SetupTest() {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})

	tApp.InitializeFromGenesisStates()

	suite.App = tApp
	suite.Ctx = ctx
	suite.Keeper = tApp.GetRouterKeeper()
	suite.StakingKeeper = tApp.GetStakingKeeper()
	suite.BankKeeper = tApp.GetBankKeeper()
	suite.EarnKeeper = tApp.GetEarnKeeper()
}

// CreateAccount creates a new account from the provided balance and address
func (suite *Suite) CreateAccountWithAddress(addr sdk.AccAddress, initialBalance sdk.Coins) authtypes.AccountI {
	ak := suite.App.GetAccountKeeper()

	acc := ak.NewAccountWithAddress(suite.Ctx, addr)
	ak.SetAccount(suite.Ctx, acc)

	err := simapp.FundAccount(suite.BankKeeper, suite.Ctx, acc.GetAddress(), initialBalance)
	suite.Require().NoError(err)

	return acc
}

// CreateVestingAccount creates a new vesting account. `vestingBalance` should be a fraction of `initialBalance`.
func (suite *Suite) CreateVestingAccountWithAddress(addr sdk.AccAddress, initialBalance sdk.Coins, vestingBalance sdk.Coins) authtypes.AccountI {
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
func (suite *Suite) AddCoinsToModule(module string, amount sdk.Coins) {
	err := simapp.FundModuleAccount(suite.BankKeeper, suite.Ctx, module, amount)
	suite.Require().NoError(err)
}

// AccountBalanceEqual checks if an account has the specified coins.
func (suite *Suite) AccountBalanceEqual(addr sdk.AccAddress, coins sdk.Coins) {
	balance := suite.BankKeeper.GetAllBalances(suite.Ctx, addr)
	suite.Equalf(coins, balance, "expected account balance to equal coins %s, but got %s", coins, balance)
}

// AccountBalanceOfEqual checks if an account has the specified amount of one denom.
func (suite *Suite) AccountBalanceOfEqual(addr sdk.AccAddress, denom string, amount sdk.Int) {
	balance := suite.BankKeeper.GetBalance(suite.Ctx, addr, denom).Amount
	suite.Equalf(amount, balance, "expected account balance to have %[1]s%[2]s, but got %[3]s%[2]s", amount, denom, balance)
}

// AccountSpendableBalanceEqual checks if an account has the specified coins unlocked.
func (suite *Suite) AccountSpendableBalanceEqual(addr sdk.AccAddress, amount sdk.Coins) {
	balance := suite.BankKeeper.SpendableCoins(suite.Ctx, addr)
	suite.Equalf(amount, balance, "expected account spendable balance to equal coins %s, but got %s", amount, balance)
}

func (suite *Suite) QueryBank_SpendableBalance(user sdk.AccAddress) sdk.Coins {
	res, err := suite.BankKeeper.SpendableBalances(
		sdk.WrapSDKContext(suite.Ctx),
		&banktypes.QuerySpendableBalancesRequest{
			Address: user.String(),
		},
	)
	suite.Require().NoError(err)
	return *&res.Balances
}

func (suite *Suite) deliverMsgCreateValidator(ctx sdk.Context, address sdk.ValAddress, selfDelegation sdk.Coin) error {
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
func (suite *Suite) NewBondCoin(amount sdk.Int) sdk.Coin {
	stakingDenom := suite.StakingKeeper.BondDenom(suite.Ctx)
	return sdk.NewCoin(stakingDenom, amount)
}

// NewBondCoins creates Coins with the current staking denom.
func (suite *Suite) NewBondCoins(amount sdk.Int) sdk.Coins {
	return sdk.NewCoins(suite.NewBondCoin(amount))
}

// CreateNewUnbondedValidator creates a new validator in the staking module.
// New validators are unbonded until the end blocker is run.
func (suite *Suite) CreateNewUnbondedValidator(addr sdk.ValAddress, selfDelegation sdk.Int) stakingtypes.Validator {
	// Create a validator
	err := suite.deliverMsgCreateValidator(suite.Ctx, addr, suite.NewBondCoin(selfDelegation))
	suite.Require().NoError(err)

	// New validators are created in an unbonded state. Note if the end blocker is run later this validator could become bonded.

	validator, found := suite.StakingKeeper.GetValidator(suite.Ctx, addr)
	suite.Require().True(found)
	return validator
}

// SlashValidator burns tokens staked in a validator. new_tokens = old_tokens * (1-slashFraction)
func (suite *Suite) SlashValidator(addr sdk.ValAddress, slashFraction sdk.Dec) {
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
func (suite *Suite) CreateDelegation(valAddr sdk.ValAddress, delegator sdk.AccAddress, amount sdk.Int) sdk.Dec {
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

// DelegationSharesEqual checks if a delegation has the specified shares.
// It expects delegations with zero shares to not be stored in state.
func (suite *Suite) DelegationSharesEqual(valAddr sdk.ValAddress, delegator sdk.AccAddress, shares sdk.Dec) bool {
	del, found := suite.StakingKeeper.GetDelegation(suite.Ctx, delegator, valAddr)

	if shares.IsZero() {
		return suite.Falsef(found, "expected delegator to not be found, got %s shares", del.Shares)
	} else {
		res := suite.True(found, "expected delegator to be found")
		return res && suite.Truef(shares.Equal(del.Shares), "expected %s delegator shares but got %s", shares, del.Shares)
	}
}

// DelegationBalanceLessThan checks if a delegation's staked token balance is less the specified amount.
// It treats not found delegations as having zero shares.
func (suite *Suite) DelegationBalanceLessThan(valAddr sdk.ValAddress, delegator sdk.AccAddress, max sdk.Int) bool {
	shares := sdk.ZeroDec()
	del, found := suite.StakingKeeper.GetDelegation(suite.Ctx, delegator, valAddr)
	if found {
		shares = del.Shares
	}

	val, found := suite.StakingKeeper.GetValidator(suite.Ctx, valAddr)
	suite.Require().Truef(found, "expected validator to be found")

	tokens := val.TokensFromShares(shares).TruncateInt()

	return suite.Truef(tokens.LT(max), "expected delegation balance to be less than %s, got %s", max, tokens)
}

// DelegationBalanceInDeltaBelow checks if a delegation's staked token balance is between `expected` and `expected - delta` inclusive.
// It treats not found delegations as having zero shares.
func (suite *Suite) DelegationBalanceInDeltaBelow(valAddr sdk.ValAddress, delegator sdk.AccAddress, expected, delta sdk.Int) bool {
	shares := sdk.ZeroDec()
	del, found := suite.StakingKeeper.GetDelegation(suite.Ctx, delegator, valAddr)
	if found {
		shares = del.Shares
	}

	val, found := suite.StakingKeeper.GetValidator(suite.Ctx, valAddr)
	suite.Require().Truef(found, "expected validator to be found")

	tokens := val.TokensFromShares(shares).TruncateInt()

	lte := suite.Truef(tokens.LTE(expected), "expected delegation balance to be less than or equal to %s, got %s", expected, tokens)
	gte := suite.Truef(tokens.GTE(expected.Sub(delta)), "expected delegation balance to be greater than or equal to %s, got %s", expected.Sub(delta), tokens)
	return lte && gte
}

// UnbondingDelegationInDeltaBelow checks if the total balance in an unbonding delegation is between `expected` and `expected - delta` inclusive.
func (suite *Suite) UnbondingDelegationInDeltaBelow(valAddr sdk.ValAddress, delegator sdk.AccAddress, expected, delta sdk.Int) bool {
	tokens := sdk.ZeroInt()
	ubd, found := suite.StakingKeeper.GetUnbondingDelegation(suite.Ctx, delegator, valAddr)
	if found {
		for _, entry := range ubd.Entries {
			tokens = tokens.Add(entry.Balance)
		}
	}

	lte := suite.Truef(tokens.LTE(expected), "expected unbonding delegation balance to be less than or equal to %s, got %s", expected, tokens)
	gte := suite.Truef(tokens.GTE(expected.Sub(delta)), "expected unbonding delegation balance to be greater than or equal to %s, got %s", expected.Sub(delta), tokens)
	return lte && gte
}

func (suite *Suite) QueryStaking_Delegation(valAddr sdk.ValAddress, delegator sdk.AccAddress) stakingtypes.DelegationResponse {
	stakingQuery := stakingkeeper.Querier{Keeper: suite.StakingKeeper}
	res, err := stakingQuery.Delegation(
		sdk.WrapSDKContext(suite.Ctx),
		&stakingtypes.QueryDelegationRequest{
			DelegatorAddr: delegator.String(),
			ValidatorAddr: valAddr.String(),
		},
	)
	suite.Require().NoError(err)
	return *res.DelegationResponse
}

// EventsContains asserts that the expected event is in the provided events
func (suite *Suite) EventsContains(events sdk.Events, expectedEvent sdk.Event) {
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

func attrsToMap(attrs []abci.EventAttribute) []sdk.Attribute {
	out := []sdk.Attribute{}

	for _, attr := range attrs {
		out = append(out, sdk.NewAttribute(string(attr.Key), string(attr.Value)))
	}

	return out
}

// CreateVault adds a new earn vault to the earn keeper parameters
func (suite *Suite) CreateVault(
	vaultDenom string,
	vaultStrategies earntypes.StrategyTypes,
	isPrivateVault bool,
	allowedDepositors []sdk.AccAddress,
) {
	vault := earntypes.NewAllowedVault(vaultDenom, vaultStrategies, isPrivateVault, allowedDepositors)

	allowedVaults := suite.EarnKeeper.GetAllowedVaults(suite.Ctx)
	allowedVaults = append(allowedVaults, vault)

	params := earntypes.NewParams(allowedVaults)

	suite.EarnKeeper.SetParams(
		suite.Ctx,
		params,
	)
}

// SetSavingsSupportedDenoms overwrites the list of supported denoms in the savings module params.
func (suite *Suite) SetSavingsSupportedDenoms(denoms []string) {
	sk := suite.App.GetSavingsKeeper()
	sk.SetParams(suite.Ctx, savingstypes.NewParams(denoms))
}

// VaultAccountValueEqual asserts that the vault account value matches the provided coin amount.
func (suite *Suite) VaultAccountValueEqual(acc sdk.AccAddress, coin sdk.Coin) {

	accVaultBal, err := suite.EarnKeeper.GetVaultAccountValue(suite.Ctx, coin.Denom, acc)
	suite.Require().NoError(err)

	suite.Require().Truef(
		coin.Equal(accVaultBal),
		"expected account vault balance to equal %s, but got %s",
		coin, accVaultBal,
	)
}

// VaultAccountSharesEqual asserts that the vault account shares match the provided values.
func (suite *Suite) VaultAccountSharesEqual(acc sdk.AccAddress, shares earntypes.VaultShares) { // TODO

	accVaultShares, found := suite.EarnKeeper.GetVaultAccountShares(suite.Ctx, acc)
	if !found {
		suite.Empty(shares)
	} else {
		suite.Equal(shares, accVaultShares)
	}
}

func (suite *Suite) QueryEarn_VaultValue(depositor sdk.AccAddress, vaultDenom string) earntypes.DepositResponse {
	earnQuery := earnkeeper.NewQueryServerImpl(suite.EarnKeeper)
	res, err := earnQuery.Deposits(
		sdk.WrapSDKContext(suite.Ctx),
		&earntypes.QueryDepositsRequest{
			Depositor: depositor.String(),
			Denom:     vaultDenom,
		},
	)
	suite.Require().NoError(err)
	suite.Require().Equalf(1, len(res.Deposits), "while earn supports one vault per denom, deposits response should be length 1")
	return res.Deposits[0]
}
