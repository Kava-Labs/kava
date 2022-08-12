package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/kava/x/liquid/types"
)

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
