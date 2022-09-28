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
			expectedErr:      types.ErrInvalidDenom,
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
		{
			name:       "error when transfer < 0",
			validator:  &validator{i(10), d("10")},
			delegation: d("10"),
			transfer:   i(-1),
			expected: returns{
				err: types.ErrUntransferableShares,
			},
		},
		{ // disallow zero transfers
			name:       "error when transfer = 0",
			validator:  &validator{i(10), d("10")},
			delegation: d("10"),
			transfer:   i(0),
			expected: returns{
				err: types.ErrUntransferableShares,
			},
		},
		{
			name:       "error when transfer > delegated shares",
			validator:  &validator{i(10), d("10")},
			delegation: d("10"),
			transfer:   i(11),
			expected: returns{
				err: sdkerrors.ErrInvalidRequest,
			},
		},
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
				suite.Equal(tc.expected.derivatives, derivatives, "expected '%s' got '%s'", tc.expected.derivatives, derivatives)
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

	initialBalance := i(1e9)
	vestedBalance := i(500e6)

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
			amount:                  suite.NewBondCoin(vestedBalance),
			expectedDerivatives:     i(500e6),
			expectedSharesRemaining: d("500000000.0"),
			expectedSharesAdded:     d("500000000.0"),
		},
		{
			name:        "error when the input denom isn't correct",
			amount:      sdk.NewCoin("invalid", i(1000)),
			expectedErr: types.ErrInvalidDenom,
		},
		{
			name:        "error when shares cannot be calculated",
			amount:      suite.NewBondCoin(initialBalance.Mul(i(100))),
			expectedErr: sdkerrors.ErrInvalidRequest,
		},
		{
			name:        "error when shares cannot be transferred",
			amount:      suite.NewBondCoin(initialBalance), // trying to move vesting coins will fail in `TransferShares`
			expectedErr: sdkerrors.ErrInsufficientFunds,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			suite.CreateAccountWithAddress(valAccAddr, suite.NewBondCoins(initialBalance))
			suite.CreateVestingAccountWithAddress(delegator, suite.NewBondCoins(initialBalance), suite.NewBondCoins(vestedBalance))

			suite.CreateNewUnbondedValidator(valAddr, initialBalance)
			suite.CreateDelegation(valAddr, delegator, initialBalance)
			staking.EndBlocker(suite.Ctx, suite.StakingKeeper)

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

			sharesTransferred := initialBalance.ToDec().Sub(tc.expectedSharesRemaining)
			suite.EventsContains(suite.Ctx.EventManager().Events(), sdk.NewEvent(
				types.EventTypeMintDerivative,
				sdk.NewAttribute(types.AttributeKeyDelegator, delegator.String()),
				sdk.NewAttribute(types.AttributeKeyValidator, valAddr.String()),
				sdk.NewAttribute(sdk.AttributeKeyAmount, derivative.String()),
				sdk.NewAttribute(types.AttributeKeySharesTransferred, sharesTransferred.String()),
			))
		})
	}
}

func (suite *KeeperTestSuite) TestIsDerivativeDenom() {
	_, addrs := app.GeneratePrivKeyAddressPairs(5)
	valAccAddr1, delegator, valAccAddr2 := addrs[0], addrs[1], addrs[2]
	valAddr1 := sdk.ValAddress(valAccAddr1)

	// Validator addr that has **not** delegated anything
	valAddr2 := sdk.ValAddress(valAccAddr2)

	initialBalance := i(1e9)
	vestedBalance := i(500e6)

	suite.CreateAccountWithAddress(valAccAddr1, suite.NewBondCoins(initialBalance))
	suite.CreateVestingAccountWithAddress(delegator, suite.NewBondCoins(initialBalance), suite.NewBondCoins(vestedBalance))

	suite.CreateNewUnbondedValidator(valAddr1, initialBalance)
	suite.CreateDelegation(valAddr1, delegator, initialBalance)
	staking.EndBlocker(suite.Ctx, suite.StakingKeeper)

	testCases := []struct {
		name        string
		denom       string
		wantIsDenom bool
	}{
		{
			name:        "valid derivative denom",
			denom:       suite.Keeper.GetLiquidStakingTokenDenom(valAddr1),
			wantIsDenom: true,
		},
		{
			name:        "invalid - undelegated validator addr",
			denom:       suite.Keeper.GetLiquidStakingTokenDenom(valAddr2),
			wantIsDenom: false,
		},
		{
			name:        "invalid - invalid val addr",
			denom:       "bkava-asdfasdf",
			wantIsDenom: false,
		},
		{
			name:        "invalid - ukava",
			denom:       "ukava",
			wantIsDenom: false,
		},
		{
			name:        "invalid - plain bkava",
			denom:       "bkava",
			wantIsDenom: false,
		},
		{
			name:        "invalid - bkava prefix",
			denom:       "bkava-",
			wantIsDenom: false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			isDenom := suite.Keeper.IsDerivativeDenom(suite.Ctx, tc.denom)

			suite.Require().Equal(tc.wantIsDenom, isDenom)
		})
	}
}

func (suite *KeeperTestSuite) TestGetStakedTokensForDerivatives() {
	_, addrs := app.GeneratePrivKeyAddressPairs(5)
	valAccAddr1, delegator, valAccAddr2, valAccAddr3 := addrs[0], addrs[1], addrs[2], addrs[3]
	valAddr1 := sdk.ValAddress(valAccAddr1)

	// Validator addr that has **not** delegated anything
	valAddr2 := sdk.ValAddress(valAccAddr2)

	valAddr3 := sdk.ValAddress(valAccAddr3)

	initialBalance := i(1e9)
	vestedBalance := i(500e6)
	delegateAmount := i(100e6)

	suite.CreateAccountWithAddress(valAccAddr1, suite.NewBondCoins(initialBalance))
	suite.CreateVestingAccountWithAddress(delegator, suite.NewBondCoins(initialBalance), suite.NewBondCoins(vestedBalance))

	suite.CreateNewUnbondedValidator(valAddr1, initialBalance)
	suite.CreateDelegation(valAddr1, delegator, delegateAmount)

	suite.CreateAccountWithAddress(valAccAddr3, suite.NewBondCoins(initialBalance))

	suite.CreateNewUnbondedValidator(valAddr3, initialBalance)
	suite.CreateDelegation(valAddr3, delegator, delegateAmount)
	staking.EndBlocker(suite.Ctx, suite.StakingKeeper)

	suite.SlashValidator(valAddr3, d("0.05"))

	_, err := suite.Keeper.MintDerivative(suite.Ctx, delegator, valAddr1, suite.NewBondCoin(delegateAmount))
	suite.Require().NoError(err)

	testCases := []struct {
		name           string
		derivatives    sdk.Coins
		wantKavaAmount sdk.Int
		err            error
	}{
		{
			name: "valid derivative denom",
			derivatives: sdk.NewCoins(
				sdk.NewCoin(suite.Keeper.GetLiquidStakingTokenDenom(valAddr1), vestedBalance),
			),
			wantKavaAmount: vestedBalance,
		},
		{
			name: "valid - slashed validator",
			derivatives: sdk.NewCoins(
				sdk.NewCoin(suite.Keeper.GetLiquidStakingTokenDenom(valAddr3), vestedBalance),
			),
			// vestedBalance * 95%
			wantKavaAmount: vestedBalance.Mul(sdk.NewInt(95)).Quo(sdk.NewInt(100)),
		},
		{
			name: "valid - sum",
			derivatives: sdk.NewCoins(
				sdk.NewCoin(suite.Keeper.GetLiquidStakingTokenDenom(valAddr3), vestedBalance),
				sdk.NewCoin(suite.Keeper.GetLiquidStakingTokenDenom(valAddr1), vestedBalance),
			),
			// vestedBalance + (vestedBalance * 95%)
			wantKavaAmount: vestedBalance.Mul(sdk.NewInt(95)).Quo(sdk.NewInt(100)).Add(vestedBalance),
		},
		{
			name: "invalid - undelegated validator address denom",
			derivatives: sdk.NewCoins(
				sdk.NewCoin(suite.Keeper.GetLiquidStakingTokenDenom(valAddr2), vestedBalance),
			),
			err: fmt.Errorf("invalid derivative denom %s: validator not found", suite.Keeper.GetLiquidStakingTokenDenom(valAddr2)),
		},
		{
			name: "invalid - denom",
			derivatives: sdk.NewCoins(
				sdk.NewCoin("kava", vestedBalance),
			),
			err: fmt.Errorf("invalid derivative denom: cannot parse denom kava"),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			kavaAmount, err := suite.Keeper.GetStakedTokensForDerivatives(suite.Ctx, tc.derivatives)

			if tc.err != nil {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				suite.Require().Equal(suite.NewBondCoin(tc.wantKavaAmount), kavaAmount)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestDerivativeFromTokens() {
	_, addrs := app.GeneratePrivKeyAddressPairs(1)
	valAccAddr := addrs[0]
	valAddr := sdk.ValAddress(valAccAddr)
	moduleAccAddress := authtypes.NewModuleAddress(types.ModuleAccountName)

	initialBalance := i(1e9)

	suite.CreateAccountWithAddress(valAccAddr, suite.NewBondCoins(initialBalance))
	suite.AddCoinsToModule(types.ModuleAccountName, suite.NewBondCoins(initialBalance))

	suite.CreateNewUnbondedValidator(valAddr, initialBalance)
	suite.CreateDelegation(valAddr, moduleAccAddress, initialBalance)
	staking.EndBlocker(suite.Ctx, suite.StakingKeeper)

	_, err := suite.Keeper.DerivativeFromTokens(suite.Ctx, valAddr, sdk.NewCoin("invalid", initialBalance))
	suite.ErrorIs(err, types.ErrInvalidDenom)

	derivatives, err := suite.Keeper.DerivativeFromTokens(suite.Ctx, valAddr, suite.NewBondCoin(initialBalance))
	suite.NoError(err)
	expected := sdk.NewCoin(fmt.Sprintf("bkava-%s", valAddr), initialBalance)
	suite.Equal(expected, derivatives)
}
