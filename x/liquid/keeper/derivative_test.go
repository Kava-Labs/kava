package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/liquid/types"
)

var (
	//d is an alias for sdk.MustNewDecFromStr
	d = sdk.MustNewDecFromStr
	// i is an alias for sdk.NewInt
	i = sdk.NewInt
)

func (suite *KeeperTestSuite) TestTransferDelegation_ValidatorStates() {
	_, addrs := app.GeneratePrivKeyAddressPairs(3)
	valAccAddr, fromDelegator, toDelegator := addrs[0], addrs[1], addrs[2]
	valAddr := sdk.ValAddress(valAccAddr)

	initialBalance := i(1e9)

	notBondedModAddr := authtypes.NewModuleAddress(stakingtypes.NotBondedPoolName)
	bondedModAddr := authtypes.NewModuleAddress(stakingtypes.BondedPoolName)

	testCases := []struct {
		name            string
		createValidator func() (delegatorShares sdk.Dec, err error)
	}{
		{
			name: "bonded validator",
			createValidator: func() (sdk.Dec, error) {
				suite.CreateNewUnbondedValidator(valAddr, initialBalance)
				delegatorShares := suite.CreateDelegation(valAddr, fromDelegator, i(1e9))

				// Run end blocker to update validator state to bonded.
				staking.EndBlocker(suite.Ctx, suite.StakingKeeper)

				return delegatorShares, nil
			},
		},
		{
			name: "unbonded validator",
			createValidator: func() (sdk.Dec, error) {
				suite.CreateNewUnbondedValidator(valAddr, initialBalance)
				delegatorShares := suite.CreateDelegation(valAddr, fromDelegator, i(1e9))

				// Don't run end blocker, new validators are by default unbonded.
				return delegatorShares, nil
			},
		},
		{
			name: "unbonding validator",
			createValidator: func() (sdk.Dec, error) {
				val := suite.CreateNewUnbondedValidator(valAddr, initialBalance)
				delegatorShares := suite.CreateDelegation(valAddr, fromDelegator, i(1e9))

				// Run end blocker to update validator state to bonded.
				staking.EndBlocker(suite.Ctx, suite.StakingKeeper)

				// Jail and run end blocker to transition validator to unbonding.
				consAddr, err := val.GetConsAddr()
				if err != nil {
					return sdk.Dec{}, err
				}
				suite.StakingKeeper.Jail(suite.Ctx, consAddr)
				staking.EndBlocker(suite.Ctx, suite.StakingKeeper)

				return delegatorShares, nil
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			suite.CreateAccountWithAddress(valAccAddr, suite.NewBondCoins(i(1e9)))
			suite.CreateAccountWithAddress(fromDelegator, suite.NewBondCoins(i(1e9)))

			fromDelegationShares, err := tc.createValidator()
			suite.Require().NoError(err)

			validator, found := suite.StakingKeeper.GetValidator(suite.Ctx, valAddr)
			suite.Require().True(found)
			notBondedBalance := suite.BankKeeper.GetAllBalances(suite.Ctx, notBondedModAddr)
			bondedBalance := suite.BankKeeper.GetAllBalances(suite.Ctx, bondedModAddr)

			shares := d("1000")

			_, err = suite.Keeper.TransferDelegation(suite.Ctx, valAddr, fromDelegator, toDelegator, shares)
			suite.Require().NoError(err)

			// Transferring a delegation should move shares, and leave the validator and pool balances the same.

			suite.DelegationSharesEqual(valAddr, fromDelegator, fromDelegationShares.Sub(shares))
			suite.DelegationSharesEqual(valAddr, toDelegator, shares) // also creates new delegation

			validatorAfter, found := suite.StakingKeeper.GetValidator(suite.Ctx, valAddr)
			suite.Require().True(found)
			suite.Equal(validator.GetTokens(), validatorAfter.GetTokens())
			suite.Equal(validator.GetDelegatorShares(), validatorAfter.GetDelegatorShares())
			suite.Equal(validator.GetStatus(), validatorAfter.GetStatus())

			suite.AccountBalanceEqual(notBondedModAddr, notBondedBalance)
			suite.AccountBalanceEqual(bondedModAddr, bondedBalance)
		})
	}
}

func (suite *KeeperTestSuite) TestTransferDelegation_Shares() {
	_, addrs := app.GeneratePrivKeyAddressPairs(3)
	valAccAddr, fromDelegator, toDelegator := addrs[0], addrs[1], addrs[2]
	valAddr := sdk.ValAddress(valAccAddr)

	initialBalance := i(1e12)

	notBondedModAddr := authtypes.NewModuleAddress(stakingtypes.NotBondedPoolName)
	bondedModAddr := authtypes.NewModuleAddress(stakingtypes.BondedPoolName)

	testCases := []struct {
		name            string
		createValidator func() (delegatorShares sdk.Dec, err error)
		shares          sdk.Dec
		expectedShares  sdk.Dec
	}{
		{
			name: "half shares burned",
			createValidator: func() (sdk.Dec, error) {
				suite.CreateNewUnbondedValidator(valAddr, i(1))
				fromDelegationShares := suite.CreateDelegation(valAddr, fromDelegator, i(1e9))
				// Run end blocker to update validator state to bonded.
				staking.EndBlocker(suite.Ctx, suite.StakingKeeper)

				return fromDelegationShares, nil
			},
			shares:         d("999999999.999999999999999999"),
			expectedShares: d("499999999.500000000499999999"),
		},
		{
			name: "slashed",
			createValidator: func() (sdk.Dec, error) {
				suite.CreateNewUnbondedValidator(valAddr, i(1e9))
				staking.EndBlocker(suite.Ctx, suite.StakingKeeper)
				// slash validator to change shares to tokens rate
				suite.SlashValidator(valAddr, d("0.05"))
				fromDelegationShares := suite.CreateDelegation(valAddr, fromDelegator, i(1e9))
				staking.EndBlocker(suite.Ctx, suite.StakingKeeper)

				return fromDelegationShares, nil
			},
			shares:         d("1000.0000000000000000"),
			expectedShares: d("999.999999999999999999"),
		},
		{
			name: "0",
			createValidator: func() (sdk.Dec, error) {
				suite.CreateNewUnbondedValidator(valAddr, i(1e9))
				fromDelegationShares := suite.CreateDelegation(valAddr, fromDelegator, i(1e9))
				staking.EndBlocker(suite.Ctx, suite.StakingKeeper)

				return fromDelegationShares, nil
			},
			shares:         d("0"),
			expectedShares: d("0"),
		},
		{
			name: "half",
			createValidator: func() (sdk.Dec, error) {
				suite.CreateNewUnbondedValidator(valAddr, i(1e9))
				fromDelegationShares := suite.CreateDelegation(valAddr, fromDelegator, i(1e9))
				staking.EndBlocker(suite.Ctx, suite.StakingKeeper)

				return fromDelegationShares, nil
			},
			shares:         d("500000000"),
			expectedShares: d("500000000"),
		},
		{
			name: "all",
			createValidator: func() (sdk.Dec, error) {
				suite.CreateNewUnbondedValidator(valAddr, i(1e9))
				fromDelegationShares := suite.CreateDelegation(valAddr, fromDelegator, i(1e9))
				// Run end blocker to update validator state to bonded.
				staking.EndBlocker(suite.Ctx, suite.StakingKeeper)

				return fromDelegationShares, nil
			},
			shares:         d("1000000000"),
			expectedShares: d("1000000000"),
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			suite.CreateAccountWithAddress(valAccAddr, suite.NewBondCoins(initialBalance))
			suite.CreateAccountWithAddress(fromDelegator, suite.NewBondCoins(initialBalance))

			fromDelegationShares, err := tc.createValidator()
			suite.Require().NoError(err)

			validator, found := suite.StakingKeeper.GetValidator(suite.Ctx, valAddr)
			suite.Require().True(found)
			notBondedBalance := suite.BankKeeper.GetAllBalances(suite.Ctx, notBondedModAddr)
			bondedBalance := suite.BankKeeper.GetAllBalances(suite.Ctx, bondedModAddr)

			newShares, err := suite.Keeper.TransferDelegation(suite.Ctx, valAddr, fromDelegator, toDelegator, tc.shares)
			suite.Require().NoError(err)

			suite.DelegationSharesEqual(valAddr, fromDelegator, fromDelegationShares.Sub(tc.shares))
			suite.DelegationSharesEqual(valAddr, toDelegator, newShares) // also creates new delegation
			suite.Equal(tc.expectedShares, newShares)

			validatorAfter, found := suite.StakingKeeper.GetValidator(suite.Ctx, valAddr)
			suite.Require().True(found)
			suite.Equal(validator.GetTokens(), validatorAfter.GetTokens())
			suite.True(validator.GetDelegatorShares().GTE(validatorAfter.GetDelegatorShares()))
			suite.Equal(validator.GetDelegatorShares().Sub(tc.shares).Add(newShares), validatorAfter.GetDelegatorShares())
			suite.Equal(validator.GetStatus(), validatorAfter.GetStatus())

			// Even though shares can change, the underlying tokens they correspond to should not not differ by more than one.
			_, previousTokens := validator.RemoveDelShares(tc.shares)
			_, newTokens := validatorAfter.RemoveDelShares(newShares)
			diff := previousTokens.Sub(newTokens)
			suite.True(diff.GTE(sdk.ZeroInt()))
			suite.True(diff.LTE(sdk.OneInt()))

			suite.AccountBalanceEqual(notBondedModAddr, notBondedBalance)
			suite.AccountBalanceEqual(bondedModAddr, bondedBalance)
		})
	}
}

func (suite *KeeperTestSuite) TestTransferDelegation_RedelegationsForbidden() {
	_, addrs := app.GeneratePrivKeyAddressPairs(4)
	val1AccAddr, val2AccAddr, fromDelegator, toDelegator := addrs[0], addrs[1], addrs[2], addrs[3]
	val1Addr := sdk.ValAddress(val1AccAddr)
	val2Addr := sdk.ValAddress(val2AccAddr)

	initialBalance := i(1e12)

	suite.CreateAccountWithAddress(val1AccAddr, suite.NewBondCoins(initialBalance))
	suite.CreateAccountWithAddress(val2AccAddr, suite.NewBondCoins(initialBalance))
	suite.CreateAccountWithAddress(fromDelegator, suite.NewBondCoins(initialBalance))

	suite.CreateNewUnbondedValidator(val1Addr, i(1e9))
	fromDelegationShares := suite.CreateDelegation(val1Addr, fromDelegator, i(1e9))
	staking.EndBlocker(suite.Ctx, suite.StakingKeeper)

	suite.CreateNewUnbondedValidator(val2Addr, i(1e9))
	suite.CreateRedelegation(fromDelegator, val1Addr, val2Addr, i(1e9))
	staking.EndBlocker(suite.Ctx, suite.StakingKeeper)

	_, err := suite.Keeper.TransferDelegation(suite.Ctx, val2Addr, fromDelegator, toDelegator, fromDelegationShares)
	suite.Require().ErrorIs(err, types.ErrRedelegationsNotCompleted)
}

func (suite *KeeperTestSuite) TestMintDerivative() {
	testCases := []struct {
		name           string
		balance        sdk.Coin
		amount         sdk.Coin
		bondValidator  bool
		vestingAccount bool
	}{
		{
			name:    "validator unbonded",
			balance: sdk.NewInt64Coin("stake", 1e9),
			amount:  sdk.NewCoin("stake", sdk.NewInt(1e6)),
		},
		{
			name:          "validator bonded",
			balance:       sdk.NewInt64Coin("stake", 1e9),
			amount:        sdk.NewCoin("stake", sdk.NewInt(1e6)),
			bondValidator: true,
		},
		{
			name:           "delegator is vesting account",
			balance:        sdk.NewInt64Coin("stake", 1e9),
			amount:         sdk.NewCoin("stake", sdk.NewInt(1e6)),
			vestingAccount: true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			// Set up delegation to validator
			var delegatorAcc authtypes.AccountI
			if !tc.vestingAccount {
				delegatorAcc = suite.CreateAccount(sdk.NewCoins(tc.balance))
			} else {
				delegatorAcc = suite.CreateVestingAccount(
					sdk.NewCoins(tc.balance),
					sdk.NewCoins(sdk.NewCoin(tc.balance.Denom, tc.balance.Amount.Mul(sdk.NewInt(10)))),
				)
			}

			delegatorAddr := delegatorAcc.GetAddress()
			validatorAddr := sdk.ValAddress(delegatorAddr)
			err := suite.deliverMsgCreateValidator(suite.Ctx, validatorAddr, tc.balance)
			suite.NoError(err)
			validator, _ := suite.StakingKeeper.GetValidator(suite.Ctx, validatorAddr)
			suite.Equal("BOND_STATUS_UNBONDED", validator.Status.String())

			// Run the EndBlocker to bond validator
			if tc.bondValidator {
				_ = suite.App.EndBlocker(suite.Ctx, abci.RequestEndBlock{})
				validator, _ = suite.StakingKeeper.GetValidator(suite.Ctx, validatorAddr)
				suite.Equal("BOND_STATUS_BONDED", validator.Status.String())
			}

			// Check delegation from delegator to validator
			delegation, found := suite.StakingKeeper.GetDelegation(suite.Ctx, delegatorAddr, validatorAddr)
			suite.True(found)
			suite.Equal(delegation.GetShares(), tc.balance.Amount.ToDec())

			// Check delegation for module account to validator
			moduleAccAddress := authtypes.NewModuleAddress(types.ModuleAccountName)
			_, found = suite.StakingKeeper.GetDelegation(suite.Ctx, moduleAccAddress, validatorAddr)
			suite.False(found)

			// Get pre-mint total supply and delegator balance of the validator-specific liquid token
			liquidTokenDenom := suite.Keeper.GetLiquidStakingTokenDenom(suite.Ctx, validatorAddr)
			delegatorBalancePre := suite.BankKeeper.GetBalance(suite.Ctx, delegatorAddr, liquidTokenDenom)
			totalSupplyPre := suite.BankKeeper.GetSupply(suite.Ctx, liquidTokenDenom)

			// Create and deliver MintDerivative msg
			msgMint := types.NewMsgMintDerivative(
				delegatorAddr,
				validatorAddr,
				tc.amount,
			)
			_, err = suite.App.MsgServiceRouter().Handler(&msgMint)(suite.Ctx, &msgMint)
			suite.NoError(err)

			// Confirm that delegator's delegated amount has decreased by correct amount of shares
			shares, err := validator.SharesFromTokens(tc.amount.Amount)
			suite.NoError(err)
			delegation, found = suite.StakingKeeper.GetDelegation(suite.Ctx, delegatorAddr, validatorAddr)
			suite.True(found)
			suite.Equal(delegation.GetShares(), tc.balance.Amount.ToDec().Sub(shares))

			// Confirm that module account's delegation holds correct amount of shares
			delegation, found = suite.StakingKeeper.GetDelegation(suite.Ctx, moduleAccAddress, validatorAddr)
			suite.True(found)
			suite.Equal(delegation.GetShares(), shares)

			// Confirm that the total supply of the validator-specific liquid tokens has increased as expected
			totalSupplyPost := suite.BankKeeper.GetSupply(suite.Ctx, liquidTokenDenom)
			suite.Equal(totalSupplyPost.Sub(totalSupplyPre).Amount, tc.amount.Amount)

			// Confirm that the delegator's balances the validator-specific liquid token has increased as expected
			delegatorBalancePost := suite.BankKeeper.GetBalance(suite.Ctx, delegatorAddr, liquidTokenDenom)
			suite.Equal(delegatorBalancePost.Sub(delegatorBalancePre).Amount, tc.amount.Amount)
		})
	}
}

func (suite *KeeperTestSuite) TestBurnDerivative() {
	testCases := []struct {
		name           string
		balance        sdk.Coin
		mintAmount     sdk.Coin
		burnAmountInt  sdk.Int // sdk.Int instead of sdk.Coin because we don't know the validator-specific token denom yet
		bondValidator  bool
		vestingAccount bool
	}{
		{
			name:          "validator unbonded",
			balance:       sdk.NewInt64Coin("stake", 1e9),
			mintAmount:    sdk.NewCoin("stake", sdk.NewInt(1e6)),
			burnAmountInt: sdk.NewInt(1e6),
		},
		{
			name:          "validator unbonded; burn less than full amount",
			balance:       sdk.NewInt64Coin("stake", 1e9),
			mintAmount:    sdk.NewCoin("stake", sdk.NewInt(1e6)),
			burnAmountInt: sdk.NewInt(999999),
		},
		{
			name:          "validator bonded",
			balance:       sdk.NewInt64Coin("stake", 1e9),
			mintAmount:    sdk.NewCoin("stake", sdk.NewInt(1e6)),
			burnAmountInt: sdk.NewInt(1e6),
			bondValidator: true,
		},
		{
			name:           "delegator is vesting account",
			balance:        sdk.NewInt64Coin("stake", 1e9),
			mintAmount:     sdk.NewCoin("stake", sdk.NewInt(1e6)),
			burnAmountInt:  sdk.NewInt(1e6),
			vestingAccount: true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			// Set up delegation to validator
			var delegatorAcc authtypes.AccountI
			if !tc.vestingAccount {
				delegatorAcc = suite.CreateAccount(sdk.NewCoins(tc.balance))
			} else {
				delegatorAcc = suite.CreateVestingAccount(
					sdk.NewCoins(tc.balance),
					sdk.NewCoins(sdk.NewCoin(tc.balance.Denom, tc.balance.Amount.Mul(sdk.NewInt(10)))),
				)
			}

			delegatorAddr := delegatorAcc.GetAddress()
			validatorAddr := sdk.ValAddress(delegatorAddr)
			err := suite.deliverMsgCreateValidator(suite.Ctx, validatorAddr, tc.balance)
			suite.NoError(err)

			validator, _ := suite.StakingKeeper.GetValidator(suite.Ctx, validatorAddr)
			suite.Equal("BOND_STATUS_UNBONDED", validator.Status.String())
			// Run the EndBlocker to bond validator
			if tc.bondValidator {
				_ = suite.App.EndBlocker(suite.Ctx, abci.RequestEndBlock{})
				validator, _ = suite.StakingKeeper.GetValidator(suite.Ctx, validatorAddr)
				suite.Equal("BOND_STATUS_BONDED", validator.Status.String())
			}

			msgMint := types.NewMsgMintDerivative(
				delegatorAddr,
				validatorAddr,
				tc.mintAmount,
			)
			_, err = suite.App.MsgServiceRouter().Handler(&msgMint)(suite.Ctx, &msgMint)
			suite.NoError(err)

			initialDelegation, _ := suite.StakingKeeper.GetDelegation(suite.Ctx, delegatorAddr, validatorAddr)

			// Confirm that the delegator has available stKava to burn
			liquidTokenDenom := suite.Keeper.GetLiquidStakingTokenDenom(suite.Ctx, validatorAddr)
			preBurnBalance := suite.BankKeeper.GetBalance(suite.Ctx, delegatorAddr, liquidTokenDenom)
			suite.Equal(sdk.NewCoin(liquidTokenDenom, tc.mintAmount.Amount), preBurnBalance) // delegated 'stake' converted to 'stStake'

			shares, err := validator.SharesFromTokens(tc.burnAmountInt)
			suite.NoError(err)

			burnAmount := sdk.NewCoin(liquidTokenDenom, tc.burnAmountInt)
			msgBurn := types.NewMsgBurnDerivative(
				delegatorAddr,
				validatorAddr,
				burnAmount,
			)
			_, err = suite.App.MsgServiceRouter().Handler(&msgBurn)(suite.Ctx, &msgBurn)
			suite.NoError(err)

			// Confirm that coins were burned from the delegator's address
			postBurnBalance := suite.BankKeeper.GetBalance(suite.Ctx, delegatorAddr, liquidTokenDenom)
			// Hacky way to compare values without sdk.Coin/sdk.Int nil throwing an error
			suite.Equal(postBurnBalance.Amount.Uint64(), preBurnBalance.Sub(burnAmount).Amount.Uint64())

			// Confirm that the delegation has been successfully returned to delegator
			delegation, found := suite.StakingKeeper.GetDelegation(suite.Ctx, delegatorAddr, validatorAddr)
			suite.True(found)
			suite.Equal(initialDelegation.Shares.Add(shares), delegation.Shares)
		})
	}
}
