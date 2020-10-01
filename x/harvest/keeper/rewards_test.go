package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/harvest/types"
)

func (suite *KeeperTestSuite) TestApplyDepositRewards() {
	type args struct {
		depositor            sdk.AccAddress
		denom                string
		depositAmount        sdk.Coin
		totalDeposits        sdk.Coin
		rewardRate           sdk.Coin
		depositType          types.DepositType
		previousBlockTime    time.Time
		blockTime            time.Time
		expectedClaimBalance sdk.Coin
	}
	type errArgs struct {
		expectPanic bool
		contains    string
	}
	type testCase struct {
		name    string
		args    args
		errArgs errArgs
	}
	testCases := []testCase{
		{
			name: "distribute rewards",
			args: args{
				depositor:            sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				denom:                "bnb",
				rewardRate:           c("hard", 500),
				depositAmount:        c("bnb", 100),
				totalDeposits:        c("bnb", 1000),
				depositType:          types.LP,
				previousBlockTime:    time.Date(2020, 11, 1, 13, 59, 50, 0, time.UTC),
				blockTime:            time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC),
				expectedClaimBalance: c("hard", 500),
			},
			errArgs: errArgs{
				expectPanic: false,
				contains:    "",
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			config := sdk.GetConfig()
			app.SetBech32AddressPrefixes(config)
			// Initialize test app and set context
			tApp := app.NewTestApp()
			ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tc.args.blockTime})
			harvestGS := types.NewGenesisState(types.NewParams(
				true,
				types.DistributionSchedules{
					types.NewDistributionSchedule(true, "bnb", time.Date(2020, 10, 8, 14, 0, 0, 0, time.UTC), time.Date(2020, 11, 22, 14, 0, 0, 0, time.UTC), tc.args.rewardRate, time.Date(2021, 11, 22, 14, 0, 0, 0, time.UTC), types.Multipliers{types.NewMultiplier(types.Small, 0, sdk.MustNewDecFromStr("0.33")), types.NewMultiplier(types.Medium, 6, sdk.MustNewDecFromStr("0.5")), types.NewMultiplier(types.Large, 24, sdk.OneDec())}),
				},
				types.DelegatorDistributionSchedules{types.NewDelegatorDistributionSchedule(
					types.NewDistributionSchedule(true, "bnb", time.Date(2020, 10, 8, 14, 0, 0, 0, time.UTC), time.Date(2025, 10, 8, 14, 0, 0, 0, time.UTC), tc.args.rewardRate, time.Date(2026, 10, 8, 14, 0, 0, 0, time.UTC), types.Multipliers{types.NewMultiplier(types.Small, 0, sdk.MustNewDecFromStr("0.33")), types.NewMultiplier(types.Medium, 6, sdk.MustNewDecFromStr("0.5")), types.NewMultiplier(types.Large, 24, sdk.OneDec())}),
					time.Hour*24,
				),
				},
			), tc.args.previousBlockTime, types.DefaultDistributionTimes)
			tApp.InitializeFromGenesisStates(app.GenesisState{types.ModuleName: types.ModuleCdc.MustMarshalJSON(harvestGS)})
			supplyKeeper := tApp.GetSupplyKeeper()
			supplyKeeper.MintCoins(ctx, types.ModuleAccountName, cs(tc.args.totalDeposits))
			keeper := tApp.GetHarvestKeeper()
			deposit := types.NewDeposit(tc.args.depositor, tc.args.depositAmount, tc.args.depositType)
			keeper.SetDeposit(ctx, deposit)
			suite.app = tApp
			suite.ctx = ctx
			suite.keeper = keeper

			if tc.errArgs.expectPanic {
				suite.Require().Panics(func() { suite.keeper.ApplyDepositRewards(suite.ctx) })
			} else {
				suite.Require().NotPanics(func() { suite.keeper.ApplyDepositRewards(suite.ctx) })
				claim, f := suite.keeper.GetClaim(suite.ctx, tc.args.depositor, tc.args.denom, tc.args.depositType)
				suite.Require().True(f)
				suite.Require().Equal(tc.args.expectedClaimBalance, claim.Amount)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestApplyDelegatorRewards() {
	type args struct {
		delegator                sdk.AccAddress
		delegatorCoins           sdk.Coins
		delegationAmount         sdk.Coin
		totalBonded              sdk.Coin
		rewardRate               sdk.Coin
		depositType              types.DepositType
		previousDistributionTime time.Time
		blockTime                time.Time
		expectedClaimBalance     sdk.Coin
	}
	type errArgs struct {
		expectPanic bool
		contains    string
	}
	type testCase struct {
		name    string
		args    args
		errArgs errArgs
	}
	testCases := []testCase{
		{
			name: "distribute rewards",
			args: args{
				delegator:                sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				delegatorCoins:           cs(c("ukava", 1000)),
				rewardRate:               c("hard", 500),
				delegationAmount:         c("ukava", 100),
				totalBonded:              c("ukava", 900),
				depositType:              types.Stake,
				previousDistributionTime: time.Date(2020, 11, 1, 13, 59, 50, 0, time.UTC),
				blockTime:                time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC),
				expectedClaimBalance:     c("hard", 500),
			},
			errArgs: errArgs{
				expectPanic: false,
				contains:    "",
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			config := sdk.GetConfig()
			app.SetBech32AddressPrefixes(config)
			// Initialize test app and set context
			tApp := app.NewTestApp()
			ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tc.args.blockTime})
			authGS := app.NewAuthGenState([]sdk.AccAddress{tc.args.delegator, sdk.AccAddress(crypto.AddressHash([]byte("other_delegator")))}, []sdk.Coins{tc.args.delegatorCoins, cs(tc.args.totalBonded)})
			harvestGS := types.NewGenesisState(types.NewParams(
				true,
				types.DistributionSchedules{
					types.NewDistributionSchedule(true, "bnb", time.Date(2020, 10, 8, 14, 0, 0, 0, time.UTC), time.Date(2020, 11, 22, 14, 0, 0, 0, time.UTC), tc.args.rewardRate, time.Date(2021, 11, 22, 14, 0, 0, 0, time.UTC), types.Multipliers{types.NewMultiplier(types.Small, 0, sdk.MustNewDecFromStr("0.33")), types.NewMultiplier(types.Medium, 6, sdk.MustNewDecFromStr("0.5")), types.NewMultiplier(types.Large, 24, sdk.OneDec())}),
				},
				types.DelegatorDistributionSchedules{types.NewDelegatorDistributionSchedule(
					types.NewDistributionSchedule(true, "ukava", time.Date(2020, 10, 8, 14, 0, 0, 0, time.UTC), time.Date(2025, 10, 8, 14, 0, 0, 0, time.UTC), tc.args.rewardRate, time.Date(2026, 10, 8, 14, 0, 0, 0, time.UTC), types.Multipliers{types.NewMultiplier(types.Small, 0, sdk.MustNewDecFromStr("0.33")), types.NewMultiplier(types.Medium, 6, sdk.MustNewDecFromStr("0.5")), types.NewMultiplier(types.Large, 24, sdk.OneDec())}),
					time.Hour*24,
				),
				},
			), types.DefaultPreviousBlockTime, types.DefaultDistributionTimes)
			tApp.InitializeFromGenesisStates(authGS, app.GenesisState{types.ModuleName: types.ModuleCdc.MustMarshalJSON(harvestGS)})
			keeper := tApp.GetHarvestKeeper()
			keeper.SetPreviousDelegationDistribution(ctx, tc.args.previousDistributionTime, "ukava")
			stakingKeeper := tApp.GetStakingKeeper()
			stakingParams := stakingKeeper.GetParams(ctx)
			stakingParams.BondDenom = "ukava"
			stakingKeeper.SetParams(ctx, stakingParams)
			validatorPubKey := ed25519.GenPrivKey().PubKey()
			validator := stakingtypes.NewValidator(sdk.ValAddress(validatorPubKey.Address()), validatorPubKey, stakingtypes.Description{})
			validator.Status = sdk.Bonded
			stakingKeeper.SetValidator(ctx, validator)
			stakingKeeper.SetValidatorByConsAddr(ctx, validator)
			stakingKeeper.SetNewValidatorByPowerIndex(ctx, validator)
			// call the after-creation hook
			stakingKeeper.AfterValidatorCreated(ctx, validator.OperatorAddress)
			_, err := stakingKeeper.Delegate(ctx, tc.args.delegator, tc.args.delegationAmount.Amount, sdk.Unbonded, validator, true)
			suite.Require().NoError(err)
			stakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
			validator, f := stakingKeeper.GetValidator(ctx, validator.OperatorAddress)
			suite.Require().True(f)
			_, err = stakingKeeper.Delegate(ctx, sdk.AccAddress(crypto.AddressHash([]byte("other_delegator"))), tc.args.totalBonded.Amount, sdk.Unbonded, validator, true)
			suite.Require().NoError(err)
			stakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
			suite.app = tApp
			suite.ctx = ctx
			suite.keeper = keeper

			if tc.errArgs.expectPanic {
				suite.Require().Panics(func() { suite.keeper.ApplyDelegationRewards(suite.ctx, suite.keeper.BondDenom(suite.ctx)) })
			} else {
				suite.Require().NotPanics(func() { suite.keeper.ApplyDelegationRewards(suite.ctx, suite.keeper.BondDenom(suite.ctx)) })
				claim, f := suite.keeper.GetClaim(suite.ctx, tc.args.delegator, tc.args.delegationAmount.Denom, tc.args.depositType)
				suite.Require().True(f)
				suite.Require().Equal(tc.args.expectedClaimBalance, claim.Amount)
			}
		})
	}
}
