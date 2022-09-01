package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/liquid/types"
)

func (suite *KeeperTestSuite) TestBurnDerivative() {
	_, addrs := app.GeneratePrivKeyAddressPairs(5)
	valAccAddr, user := addrs[0], addrs[1]
	valAddr := sdk.ValAddress(valAccAddr)

	liquidDenom := suite.Keeper.GetLiquidStakingTokenDenom(valAddr)

	testCases := []struct {
		name             string
		balance          sdk.Coin
		moduleDelegation sdk.Int
		burnAmount       sdk.Coin
		expectedErr      error
	}{
		{
			name:             "user can burn their entire balance",
			balance:          c(liquidDenom, 1e9),
			moduleDelegation: i(1e9),
			burnAmount:       c(liquidDenom, 1e9),
		},
		{
			name:             "user can burn minimum derivative unit",
			balance:          c(liquidDenom, 1e9),
			moduleDelegation: i(1e9),
			burnAmount:       c(liquidDenom, 1),
		},
		{
			name:             "error when denom cannot be parsed",
			balance:          c(liquidDenom, 1e9),
			moduleDelegation: i(1e9),
			burnAmount:       c(fmt.Sprintf("ckava-%s", valAddr), 1e6),
			expectedErr:      types.ErrInvalidDerivativeDenom,
		},
		{
			name:             "error when burn amount is 0",
			balance:          c(liquidDenom, 1e9),
			moduleDelegation: i(1e9),
			burnAmount:       c(liquidDenom, 0),
			expectedErr:      types.ErrUntransferableShares,
		},
		{
			name:             "error when user doesn't have enough funds",
			balance:          c("ukava", 10),
			moduleDelegation: i(1e9),
			burnAmount:       c(liquidDenom, 1e9),
			expectedErr:      sdkerrors.ErrInsufficientFunds,
		},
		{
			name:             "error when backing delegation isn't large enough",
			balance:          c(liquidDenom, 1e9),
			moduleDelegation: i(999_999_999),
			burnAmount:       c(liquidDenom, 1e9),
			expectedErr:      stakingtypes.ErrNotEnoughDelegationShares,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			suite.CreateAccountWithAddress(valAccAddr, suite.NewBondCoins(i(1e6)))
			suite.CreateAccountWithAddress(user, sdk.NewCoins(tc.balance))
			suite.AddCoinsToModule(types.ModuleAccountName, suite.NewBondCoins(tc.moduleDelegation))

			// create delegation from module account to back the derivatives
			moduleAccAddress := authtypes.NewModuleAddress(types.ModuleAccountName)
			suite.CreateNewUnbondedValidator(valAddr, i(1e6))
			suite.CreateDelegation(valAddr, moduleAccAddress, tc.moduleDelegation)
			staking.EndBlocker(suite.Ctx, suite.StakingKeeper)
			modBalance := suite.BankKeeper.GetAllBalances(suite.Ctx, moduleAccAddress)

			_, err := suite.Keeper.BurnDerivative(suite.Ctx, user, valAddr, tc.burnAmount)

			suite.Require().ErrorIs(err, tc.expectedErr)
			if tc.expectedErr != nil {
				// if an error is expected, state should be reverted so don't need to test state is unchanged
				return
			}

			suite.AccountBalanceEqual(user, sdk.NewCoins(tc.balance.Sub(tc.burnAmount)))
			suite.AccountBalanceEqual(moduleAccAddress, modBalance) // ensure derivatives are burned, and not in module account

			sharesTransferred := tc.burnAmount.Amount.ToDec()
			suite.DelegationSharesEqual(valAddr, user, sharesTransferred)
			suite.DelegationSharesEqual(valAddr, moduleAccAddress, tc.moduleDelegation.ToDec().Sub(sharesTransferred))

			suite.EventsContains(suite.Ctx.EventManager().Events(), sdk.NewEvent(
				types.EventTypeBurnDerivative,
				sdk.NewAttribute(types.AttributeKeyDelegator, user.String()),
				sdk.NewAttribute(types.AttributeKeyValidator, valAddr.String()),
				sdk.NewAttribute(sdk.AttributeKeyAmount, tc.burnAmount.String()),
				sdk.NewAttribute(types.AttributeKeySharesTransferred, sharesTransferred.String()),
			))
		})
	}
}

func (suite *KeeperTestSuite) TestCalculateShares() {
	_, addrs := app.GeneratePrivKeyAddressPairs(5)
	valAccAddr, delegator := addrs[0], addrs[1]
	valAddr := sdk.ValAddress(valAccAddr)

	type returns struct {
		derivatives sdk.Int
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
		{
			name:       "error when transfer > tokens",
			validator:  &validator{i(10), d("10")},
			delegation: d("10"),
			transfer:   i(11),
			expected: returns{
				err: sdkerrors.ErrInvalidRequest,
			},
		},
		// { // TODO catch these cases?
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
				err: stakingtypes.ErrInsufficientShares,
			},
		},
		{
			name:       "shares and derivatives are truncated",
			validator:  &validator{i(3), d("4")},
			delegation: d("4"),
			transfer:   i(2),
			expected: returns{
				derivatives: i(2),                      // truncated down
				shares:      d("2.666666666666666666"), // 2/3 * 4 not rounded to ...667
			},
		},
		{
			name:       "error if calculated shares > shares in delegation",
			validator:  &validator{i(3), d("4")},
			delegation: d("2.666666666666666665"), // one less than 2/3 * 4
			transfer:   i(2),
			expected: returns{
				err: sdkerrors.ErrInvalidRequest,
			},
		},
		{
			name:       "error when validator not found",
			validator:  nil,
			delegation: d("1000000000"),
			transfer:   i(500e6),
			expected: returns{
				err: stakingtypes.ErrNoValidatorFound,
			},
		},
		{
			name:       "error when delegation not found",
			validator:  &validator{i(1e9), d("1000000000")},
			delegation: sdk.Dec{},
			transfer:   i(500e6),
			expected: returns{
				err: stakingtypes.ErrNoDelegation,
			},
		},
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
		expectedSharesAdded     sdk.Dec
		expectedErr             error
	}{
		{
			name:                    "derivative is minted",
			amount:                  suite.NewBondCoin(i(333_333_333)),
			expectedDerivatives:     i(1e9 - 2),
			expectedSharesRemaining: d("1.499999999250000001"), // not 1.5
			expectedSharesAdded:     d("999999998.500000000750000000"),
		},
		{
			name:        "error when the input denom isn't correct",
			amount:      sdk.NewCoin("invalid", i(1000)),
			expectedErr: types.ErrOnlyBondDenomAllowedForTokenize,
		},
		{
			name:        "error when shares cannot be calculated",
			amount:      suite.NewBondCoin(i(1e15)),
			expectedErr: sdkerrors.ErrInvalidRequest,
		},
		{
			name:        "error when shares cannot be transferred",
			amount:      sdk.Coin{Denom: suite.StakingKeeper.BondDenom(suite.Ctx), Amount: i(-1)}, // TODO find better way to trigger this
			expectedErr: types.ErrUntransferableShares,
		},
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

			_, err := suite.Keeper.MintDerivative(suite.Ctx, delegator, valAddr, tc.amount)

			suite.Require().ErrorIs(err, tc.expectedErr)
			if tc.expectedErr != nil {
				// if an error is expected, state should be reverted so don't need to test state is unchanged
				return
			}

			derivative := sdk.NewCoins(sdk.NewCoin(fmt.Sprintf("bkava-%s", valAddr), tc.expectedDerivatives))
			suite.AccountBalanceEqual(delegator, derivative)

			suite.DelegationSharesEqual(valAddr, delegator, tc.expectedSharesRemaining)
			suite.DelegationSharesEqual(valAddr, moduleAccAddress, tc.expectedSharesAdded)

			if tc.expectedErr == nil {
				sharesTransferred := initialBalance.ToDec().Sub(tc.expectedSharesRemaining)
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

/*
BurnDerivative
inputs: user, coin amount, (validator)
state: module delegation, user delegation, (validator)

effect:
- user balance decrease, coins burned
- user delegation increase (maybe created)
- module delegation decrease
- update vesting tracking in account??
- events


err - user has no balance
err - invalid coin denom
err - module has no delegation (should be impossible)

not enough balance
not enough delegation (should be impossible)

user delegation doesn't exist, is created

create user with bkava
create validator (with self delegation)
create delegation from module account





shares, tokens
us, ut (up to 1 token less than correct amount)
newShares = (ut / (tokens-ut)) * (shares-s)
= shares-s / (tokens/ut - 1)
what is shares-s / (tokens/ut - 1) - shares-s / (tokens/ut` - 1)

shares-s * (1 / (tokens/ut - 1) - 1 / (tokens/ut` - 1))

shares-s * (1 / (tokens/ut - 1) - 1 / (tokens/(ut-∂) - 1))

either reach in to staking more to just move shares around
or do some calculations to bound the error more (given small slash fractions it's probably fine)

= shares-s * 1 / (tokens/(ut-∂) - 1) where 0 < ∂ < 1
1 / (tokens/(ut-∂) - 1) is the fraction by which the new shares are too small

what would the values be such that it was < 0.999 (ie 0.1% loss) with the worst ∂ of 1
0.999 > 1 / (tokens/(ut-1) - 1)
0.999 * (tokens/(ut-1) - 1) > 1
(tokens/(ut-1) - 1) > 1/0.999
tokens/(ut-1) > 1 + 1/0.999
tokens > (ut-1)(1 + 1/0.999)
tokens > ut(1 + 1/0.999) - (1 + 1/0.999)

so as long as roughly ut < 0.5*tokens is seems ok

new shares, as a fraction of original shares moved

ut = floor(tokens * sIn/shares)
newShares = (ut / (tokens-ut)) * (shares-sIn)


newShares = ((tokens * sIn/shares) / (tokens-(tokens * sIn/shares))) * (shares-sIn)
newShares = ((sIn/shares) / (1-sIn/shares)) * (shares-sIn)
newShares = ((sIn/shares) / ((shares-sIn)/shares)) * (shares-sIn)
newShares = (sIn / (shares-sIn)) * (shares-sIn)
newShares = sIn, as expected

newShares = (1 / (tokens/ut-1)) * (shares-sIn)






newShare/sIn = (shares/sIn - 1) * 1 / sIn(tokens/ut - 1)

newShare/sIn = (shares/sIn - 1) * 1 / (shares-sIn)
newShare/sIn = ((shares-sIn)/sIn) * 1 / (shares-sIn)


There's an edge case where a validator with super low self delegation, and delegator with larger amount, can burn lots of shares.
But in that case the user effectively just gives a single unit to the validator.
Burning shares is ok for incentive - it collects the total every block, and individual amounts can change however.
*/
/*
	validator _/un/ing
		bonded - deliver msg, run end blocker
		unbonded - create validator but don't run end blocker (could also create 101th validator then end blocker can run (it won't do anything))
		unbonding - create validator, jail (or slash until it's out of the set), run end blocker
		slashed&bonded - create validator, run end blocker, call slash
	shares - 0, max, outside range

	from delegator, validator don't exist - trivial
	to delegator doesn't exist - required
	delegation has a previous redelegation
	from delegation is self-delegation (check validator is jailed (or block this case))

	check exact shares moved, check no tokens moved, check validator no change in status, no change in tokens or total shares
	or error
*/

/*
Mint
Convert ukava amt into shares to transfer and rounded shares to mint.
Deal with case of user wanting to move entire position.


Burn
burn tokens, transfer equivalent shares (possibly losing 1 ukava in the process)


MintDerivative tests
- errors
	- no delegation
	- incorrect denom
	- 0 amount
	- vesting account ?
	- invalid coin?
- convert 0, below shares, exact shares, 1 above shares, much above shares

- check mint amount, delegation transferred, event

BurnDerivative tests
- errors
- pay out remaining shares to last withdrawer?

TestRoundTripLoss
mint then burn, check amount is less than 1

*/
