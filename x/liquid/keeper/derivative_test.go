package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/kava-labs/kava/x/liquid/types"
)

var d = sdk.MustNewDecFromStr

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
