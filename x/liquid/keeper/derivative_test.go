package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/liquid/types"
)

func (suite *KeeperTestSuite) TestMintDerivative_Old() {
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


func (suite *KeeperTestSuite) TestCalculateShares() {
	_, addrs := app.GeneratePrivKeyAddressPairs(5)
	valAccAddr, delegator := addrs[0], addrs[1]
	valAddr := sdk.ValAddress(valAccAddr)

	type returns struct {
		derivatives sdk.Int // truncated shares
		shares      sdk.Dec
		err         error
	}
	type validator struct {
		tokens          sdk.Int
		delegatorShares sdk.Dec
	}
	testCases := []struct {
		name       string
		validator  *validator
		delegation sdk.Dec
		transfer   sdk.Int
		expected   returns
	}{
		// Ignoring delegation amounts
		{
			name:       "error when transfer > tokens",
			validator:  &validator{i(10), d("10")},
			delegation: d("10"),
			transfer:   i(11),
			expected: returns{
				err: types.ErrInvalidMint,
			},
		},
		// {
		// 	name:       "error when transfer = 0",
		// 	validator:  &validator{i(10), d("10")},
		// 	delegation: d("10"),
		// 	transfer:   i(0),
		// 	expected: returns{
		// 		err: types.ErrInvalidMint,
		// 	},
		// },
		// {
		// 	name:       "error when transfer < 0",
		// 	validator:  &validator{i(10), d("10")},
		// 	delegation: d("10"),
		// 	transfer:   i(-1),
		// 	expected: returns{
		// 		err: types.ErrInvalidMint,
		// 	},
		// },
		{
			name:       "error when validator has no tokens",
			validator:  &validator{i(0), d("10")},
			delegation: d("10"),
			transfer:   i(5),
			expected: returns{
				err: types.ErrInvalidMint,
			},
		},
		{
			name:       "shares and derivatives are truncated",
			validator:  &validator{i(3), d("4")},
			delegation: d("4"),
			transfer:   i(2),
			expected: returns{
				derivatives: i(2),                      // not rounded to 3
				shares:      d("2.666666666666666666"), // not rounded to ...667
			},
		},
		// delegation limits
		// TODO
		// invalid state
		// {
		// 	name:       "error when validator not found",
		// 	validator:  nil,
		// 	delegation: d("1000000000"),
		// 	transfer:   i(500e6),
		// 	expected: returns{
		// 		err: types.ErrNoValidatorFound,
		// 	},
		// },
		// {
		// 	name:       "error when delegation not found",
		// 	validator:  &validator{i(1e9), d("1000000000")},
		// 	delegation: sdk.Dec{},
		// 	transfer:   i(500e6),
		// 	expected: returns{
		// 		err: types.ErrNoDelegatorForAddress,
		// 	},
		// },
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			if tc.validator != nil {
				suite.StakingKeeper.SetValidator(suite.Ctx, stakingtypes.Validator{
					OperatorAddress: valAddr.String(),
					Tokens:          tc.validator.tokens,
					DelegatorShares: tc.validator.delegatorShares,
				})
			}
			if !tc.delegation.IsNil() {
				suite.StakingKeeper.SetDelegation(suite.Ctx, stakingtypes.Delegation{
					DelegatorAddress: delegator.String(),
					ValidatorAddress: valAddr.String(),
					Shares:           tc.delegation,
				})
			}

			derivatives, shares, err := suite.Keeper.CalculateDerivativeSharesFromTokens(suite.Ctx, delegator, valAddr, tc.transfer)
			if tc.expected.err != nil {
				suite.ErrorIs(err, tc.expected.err)
			} else {
				suite.NoError(err)
				suite.Equal(tc.expected.derivatives, derivatives)
				suite.Equal(tc.expected.shares, shares)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestMintDerivative() {
	_, addrs := app.GeneratePrivKeyAddressPairs(5)
	valAccAddr, delegator := addrs[0], addrs[1]
	valAddr := sdk.ValAddress(valAccAddr)
	moduleAccAddress := authtypes.NewModuleAddress(types.ModuleAccountName)

	testCases := []struct {
		name                    string
		amount                  sdk.Coin
		expectedDerivatives     sdk.Int
		expectedSharesRemaining sdk.Dec
		expectedErr             error
	}{
		// {
		// 	name:                    "derivative is minted",
		// 	amount:                  suite.NewBondCoin(i(333_333_333)),
		// 	expectedDerivatives:     i(1e9 - 2),
		// 	expectedSharesRemaining: d("1.5"),
		// },
		// {
		// 	name:                    "when the input denom isn't correct derivatives are not minted",
		// 	amount:                  sdk.NewCoin("invalid", i(1000)),
		// 	expectedDerivatives:     i(0),
		// 	expectedSharesRemaining: d("1000000000"),
		// 	expectedErr:             types.ErrOnlyBondDenomAllowedForTokenize,
		// },
		// // TODO vesting {name: "derivative cannot be minted for vesting account"}
		// {
		// 	name:                    "when shares cannot be calculated derivatives are not minted",
		// 	amount:                  sdk.Coin{Denom: suite.StakingKeeper.BondDenom(suite.Ctx), Amount: i(-1)},
		// 	expectedDerivatives:     i(0),
		// 	expectedSharesRemaining: d("1000000000"),
		// 	expectedErr:             types.ErrInvalidMint,
		// },
		// {
		// 	name:                    "when shares cannot be transferred derivatives are not minted",
		// 	amount:                  suite.NewBondCoin(i(1e9)),
		// 	expectedDerivatives:     i(0),
		// 	expectedSharesRemaining: d("1000000000"),
		// 	expectedErr:             types.ErrNotEnoughDelegationShares,
		// },
		// TODO check for zero transfer? {
		// 	name:                    "when input is 0 derivative is not minted",
		// 	amount:                  suite.NewBondCoin(i(0)),
		// 	expectedDerivatives:     i(0),
		// 	expectedSharesRemaining: d("1000000000"),
		// 	expectedErr:             types.ErrInvalidRequest,
		// },
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			initialBalance := i(1e9)

			suite.CreateAccountWithAddress(valAccAddr, suite.NewBondCoins(initialBalance))
			suite.CreateAccountWithAddress(delegator, suite.NewBondCoins(initialBalance))

			suite.CreateNewUnbondedValidator(valAddr, initialBalance)
			suite.CreateDelegation(valAddr, delegator, initialBalance)
			staking.EndBlocker(suite.Ctx, suite.StakingKeeper)

			// slash validator so that the user's delegation owns fractional tokens to allow more complex tests
			suite.SlashValidator(valAddr, d("0.666666666666666667"))
			val, found := suite.StakingKeeper.GetValidator(suite.Ctx, valAddr)
			suite.Require().True(found)
			suite.Equal(i(666666667), val.GetTokens()) // note the slash amount is truncated to an int before being removed from the validator

			err := suite.Keeper.MintDerivative(suite.Ctx, delegator, valAddr, tc.amount)
			suite.ErrorIs(err, tc.expectedErr)

			derivative := sdk.NewCoins(sdk.NewCoin(fmt.Sprintf("bkava-%s", valAddr), tc.expectedDerivatives))
			suite.AccountBalanceEqual(delegator, derivative)

			suite.DelegationSharesEqual(valAddr, delegator, tc.expectedSharesRemaining)
			sharesTransferred := initialBalance.ToDec().Sub(tc.expectedSharesRemaining)
			suite.DelegationSharesEqual(valAddr, moduleAccAddress, sharesTransferred)

			if tc.expectedErr != nil {
				suite.EventsDoNotContainType(suite.Ctx.EventManager().Events(), types.EventTypeMintDerivative)
			} else {
				suite.EventsContains(suite.Ctx.EventManager().Events(), sdk.NewEvent(
					types.EventTypeMintDerivative,
					sdk.NewAttribute(types.AttributeKeyDelegator, delegator.String()),
					sdk.NewAttribute(types.AttributeKeyValidator, valAddr.String()),
					sdk.NewAttribute(sdk.AttributeKeyAmount, derivative.String()),
					sdk.NewAttribute(types.AttributeKeySharesTransferred, sharesTransferred.String()),
				))
			}
		})
	}
}
