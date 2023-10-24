package e2e_test

import (
	"context"
	"fmt"
	"strconv"
	"time"

	sdkmath "cosmossdk.io/math"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	abci "github.com/tendermint/tendermint/abci/types"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/kava-labs/kava/tests/util"
	communitytypes "github.com/kava-labs/kava/x/community/types"
	kavadisttypes "github.com/kava-labs/kava/x/kavadist/types"
)

func (suite *IntegrationTestSuite) TestUpgradeInflation_Disable() {
	suite.SkipIfUpgradeDisabled()

	afterUpgradeCtx := util.CtxAtHeight(suite.UpgradeHeight)

	// Get x/community for switchover time
	params, err := suite.Kava.Community.Params(afterUpgradeCtx, &communitytypes.QueryParamsRequest{})
	suite.Require().NoError(err)

	// Sleep until switchover time + 6 seconds for extra block
	sleepDuration := time.Until(params.Params.UpgradeTimeDisableInflation.Add(6 * time.Second))
	time.Sleep(sleepDuration)

	suite.Require().Eventually(func() bool {
		communityParams, err := suite.Kava.Community.Params(afterUpgradeCtx, &communitytypes.QueryParamsRequest{})
		suite.Require().NoError(err)

		// After params are set in x/community -- non-zero switchover time
		return !communityParams.Params.UpgradeTimeDisableInflation.Equal(time.Time{})
	}, 20*time.Second, 3*time.Second)

	// Fetch exact block when inflation stop event emitted
	// This is run after the switchover, so we don't need to poll
	_, switchoverHeight, err := suite.Kava.GetBeginBlockEventsFromQuery(
		context.Background(),
		fmt.Sprintf(
			"%s.%s EXISTS",
			communitytypes.EventTypeInflationStop,
			communitytypes.AttributeKeyInflationDisableTime,
		),
	)
	suite.Require().NoError(err)
	suite.Require().NotZero(switchoverHeight)

	// 1 block before switchover
	beforeSwitchoverCtx := util.CtxAtHeight(switchoverHeight - 1)
	afterSwitchoverCtx := util.CtxAtHeight(switchoverHeight)

	suite.Run("x/mint, x/kavadist inflation before switchover", func() {
		mintParams, err := suite.Kava.Mint.Params(
			beforeSwitchoverCtx,
			&minttypes.QueryParamsRequest{},
		)
		suite.NoError(err)
		kavaDistParams, err := suite.Kava.Kavadist.Params(
			beforeSwitchoverCtx,
			&kavadisttypes.QueryParamsRequest{},
		)
		suite.NoError(err)

		// Use .String() to compare Decs since x/mint uses the deprecated one,
		// mismatch of types but same value.
		suite.Equal(
			sdkmath.LegacyMustNewDecFromStr("0.595000000000000000").String(),
			mintParams.Params.InflationMin.String(),
			"x/mint inflation min should be 59.5%% before switchover",
		)
		suite.Equal(
			sdkmath.LegacyMustNewDecFromStr("0.595000000000000000").String(),
			mintParams.Params.InflationMax.String(),
			"x/mint inflation max should be 59.5%% before switchover",
		)

		suite.True(
			kavaDistParams.Params.Active,
			"x/kavadist should be active before switchover",
		)
	})

	suite.Run("x/distribution community tax before switchover", func() {
		distrParams, err := suite.Kava.Distribution.Params(
			beforeSwitchoverCtx,
			&distributiontypes.QueryParamsRequest{},
		)
		suite.NoError(err)

		suite.Equal(
			sdkmath.LegacyMustNewDecFromStr("0.949500000000000000").String(),
			distrParams.Params.CommunityTax.String(),
			"x/distribution community tax should be 94.95%% before switchover",
		)
	})

	suite.Run("x/mint, x/kavadist inflation after switchover", func() {
		mintParams, err := suite.Kava.Mint.Params(
			afterSwitchoverCtx,
			&minttypes.QueryParamsRequest{},
		)
		suite.NoError(err)
		kavaDistParams, err := suite.Kava.Kavadist.Params(
			afterSwitchoverCtx,
			&kavadisttypes.QueryParamsRequest{},
		)
		suite.NoError(err)

		suite.Equal(
			sdkmath.LegacyZeroDec().String(),
			mintParams.Params.InflationMin.String(),
			"x/mint inflation min should be 0% after switchover",
		)
		suite.Equal(
			sdkmath.LegacyZeroDec().String(),
			mintParams.Params.InflationMax.String(),
			"x/mint inflation max should be 0% after switchover",
		)

		suite.False(
			kavaDistParams.Params.Active,
			"x/kavadist should be inactive after switchover",
		)
	})

	suite.Run("x/distribution community tax after switchover", func() {
		distrParams, err := suite.Kava.Distribution.Params(
			afterSwitchoverCtx,
			&distributiontypes.QueryParamsRequest{},
		)
		suite.NoError(err)

		suite.Equal(
			sdkmath.LegacyZeroDec().String(),
			distrParams.Params.CommunityTax.String(),
			"x/distribution community tax should be 0%% before switchover",
		)
	})

	// Ensure inflation was still active before switchover
	suite.Run("positive mint events before switchover", func() {
		// 1 block before switchover
		queryHeight := switchoverHeight - 1

		block, err := suite.Kava.TmSignClient.BlockResults(
			context.Background(),
			&queryHeight,
		)
		suite.Require().NoError(err)

		// Mint events should only occur in begin block
		mintEvents := util.FilterEventsByType(block.BeginBlockEvents, minttypes.EventTypeMint)

		suite.Require().NotEmpty(mintEvents, "mint events should be emitted")

		// Ensure mint amounts are non-zero
		found := false
		for _, event := range mintEvents {
			for _, attribute := range event.Attributes {
				// Bonded ratio and annual provisions unchecked

				if string(attribute.Key) == minttypes.AttributeKeyInflation {
					suite.Equal(
						sdkmath.LegacyMustNewDecFromStr("0.595000000000000000").String(),
						string(attribute.Value),
						"inflation should be 59.5%% before switchover",
					)
				}

				if string(attribute.Key) == sdk.AttributeKeyAmount {
					found = true
					// Parse as native go int, not necessary to use sdk.Int
					value, err := strconv.Atoi(string(attribute.Value))
					suite.Require().NoError(err)

					suite.NotZero(value, "mint amount should be non-zero")
					suite.Positive(value, "mint amount should be positive")
				}
			}
		}

		suite.True(found, "mint amount should be found")
	})

	suite.Run("staking denom supply increases before switchover", func() {
		queryHeight := switchoverHeight - 2

		supply1, err := suite.Kava.Bank.SupplyOf(
			util.CtxAtHeight(queryHeight),
			&types.QuerySupplyOfRequest{
				Denom: suite.Kava.StakingDenom,
			},
		)
		suite.Require().NoError(err)

		suite.NotZero(supply1.Amount, "ukava supply should be non-zero")

		// Next block
		queryHeight += 1
		supply2, err := suite.Kava.Bank.SupplyOf(
			util.CtxAtHeight(queryHeight),
			&types.QuerySupplyOfRequest{
				Denom: suite.Kava.StakingDenom,
			},
		)
		suite.Require().NoError(err)

		suite.NotZero(supply2.Amount, "ukava supply should be non-zero")

		suite.Truef(
			supply2.Amount.Amount.GT(supply1.Amount.Amount),
			"ukava supply before switchover should increase between blocks, %s > %s",
			supply2.Amount.Amount.String(),
		)
	})

	// Check if inflation is ACTUALLY disabled... check if any coins are being
	// minted in the blocks after switchover
	suite.Run("no minting after switchover", func() {
		kavaSupply := sdk.NewCoin(suite.Kava.StakingDenom, sdkmath.ZeroInt())

		// Next 5 blocks after switchover, ensure there's actually no more inflation
		for i := 0; i < 5; i++ {
			queryHeight := switchoverHeight + int64(i)

			suite.Run(
				fmt.Sprintf("x/mint events with 0 amount @ height=%d", queryHeight),
				func() {
					var block *coretypes.ResultBlockResults
					suite.Require().Eventually(func() bool {
						// Check begin block events
						block, err = suite.Kava.TmSignClient.BlockResults(
							context.Background(),
							&queryHeight,
						)

						return err == nil
					}, 20*time.Second, 3*time.Second)

					var mintEvents []abci.Event

					// Mint events should only occur in begin block, but we just include
					// everything else just in case anything changes in x/mint
					mintEventsBegin := util.FilterEventsByType(block.BeginBlockEvents, minttypes.EventTypeMint)
					mintEventsEnd := util.FilterEventsByType(block.EndBlockEvents, minttypes.EventTypeMint)
					mintEventsTx := util.FilterTxEventsByType(block.TxsResults, minttypes.EventTypeMint)

					mintEvents = append(mintEvents, mintEventsBegin...)
					mintEvents = append(mintEvents, mintEventsEnd...)
					mintEvents = append(mintEvents, mintEventsTx...)

					suite.Require().NotEmpty(mintEvents, "mint events should still be emitted")

					// Ensure mint amounts are 0
					found := false
					for _, event := range mintEvents {
						for _, attribute := range event.Attributes {
							// Bonded ratio and annual provisions unchecked

							if string(attribute.Key) == minttypes.AttributeKeyInflation {
								suite.Equal(sdkmath.LegacyZeroDec().String(), string(attribute.Value))
							}

							if string(attribute.Key) == sdk.AttributeKeyAmount {
								found = true
								suite.Equal(sdkmath.ZeroInt().String(), string(attribute.Value))
							}
						}
					}

					suite.True(found, "mint amount should be found")
				},
			)

			// Run this after the events check, since that one waits for the
			// new block if necessary
			suite.Run(
				fmt.Sprintf("total staking denom supply should not change @ height=%d", queryHeight),
				func() {
					supplyRes, err := suite.Kava.Bank.SupplyOf(
						util.CtxAtHeight(queryHeight),
						&types.QuerySupplyOfRequest{
							Denom: suite.Kava.StakingDenom,
						},
					)
					suite.Require().NoError(err)

					if kavaSupply.IsZero() {
						// First iteration, set supply
						kavaSupply = supplyRes.Amount
					} else {
						suite.Require().Equal(
							kavaSupply,
							supplyRes.Amount,
							"ukava supply should not change",
						)
					}
				},
			)
		}
	})

	suite.Run("no staking rewards from x/community before switchover", func() {
		// 1 block before switchover
		queryHeight := switchoverHeight - 1

		block, err := suite.Kava.TmSignClient.BlockResults(
			context.Background(),
			&queryHeight,
		)
		suite.Require().NoError(err)

		// Events are not emitted if amount is 0
		stakingRewardEvents := util.FilterEventsByType(block.BeginBlockEvents, communitytypes.EventTypeStakingRewardsPaid)
		suite.Require().Empty(stakingRewardEvents, "staking reward events should not be emitted")
	})

	suite.Run("staking rewards pay out from x/community after switchover", func() {
		for i := 0; i < 5; i++ {
			// after switchover
			queryHeight := switchoverHeight + int64(i)

			block, err := suite.Kava.TmSignClient.BlockResults(
				context.Background(),
				&queryHeight,
			)
			suite.Require().NoError(err)

			stakingRewardEvents := util.FilterEventsByType(
				block.BeginBlockEvents,
				communitytypes.EventTypeStakingRewardsPaid,
			)
			suite.Require().NotEmptyf(
				stakingRewardEvents,
				"staking reward events should be emitted at height=%d",
				queryHeight,
			)

			// Ensure amounts are non-zero
			found := false
			for _, attr := range stakingRewardEvents[0].Attributes {
				if string(attr.Key) == communitytypes.AttributeKeyStakingRewardAmount {
					coins, err := sdk.ParseCoinNormalized(string(attr.Value))
					suite.Require().NoError(err, "staking reward amount should be parsable coins")

					suite.Truef(
						coins.Amount.IsPositive(),
						"staking reward amount should be a positive amount at height=%d",
						queryHeight,
					)
					found = true
				}
			}

			suite.Truef(
				found,
				"staking reward amount should be found in events at height=%d",
				queryHeight,
			)
		}
	})

	// Staking rewards can still be claimed
	suite.Run("staking rewards claimable after switchover", func() {
		suite.SkipIfKvtoolDisabled()

		// Get the delegator of the only validator
		validators, err := suite.Kava.Staking.Validators(
			context.Background(),
			&stakingtypes.QueryValidatorsRequest{},
		)
		suite.Require().NoError(err)
		suite.Require().Positive(len(validators.Validators), "should only be at least 1 validator")

		valAddr, err := sdk.ValAddressFromBech32(validators.Validators[0].OperatorAddress)
		suite.Require().NoError(err)

		accAddr := sdk.AccAddress(valAddr.Bytes())

		balBefore, err := suite.Kava.Bank.Balance(
			context.Background(),
			&types.QueryBalanceRequest{
				Address: accAddr.String(),
				Denom:   suite.Kava.StakingDenom,
			},
		)
		suite.Require().NoError(err)
		suite.Require().False(balBefore.Balance.IsZero(), "val staking denom balance should be non-zero")

		delegationRewards, err := suite.Kava.Distribution.DelegationRewards(
			context.Background(),
			&distributiontypes.QueryDelegationRewardsRequest{
				ValidatorAddress: valAddr.String(),
				DelegatorAddress: accAddr.String(),
			},
		)
		suite.Require().NoError(err)

		suite.False(delegationRewards.Rewards.Empty())
		suite.True(delegationRewards.Rewards.IsAllPositive(), "queried rewards should be positive")

		withdrawRewardsMsg := distributiontypes.NewMsgWithdrawDelegatorReward(
			accAddr,
			valAddr,
		)

		// Get the validator private key from kava keyring
		key, err := suite.Kava.Keyring.(unsafeExporter).ExportPrivateKeyObject(
			"validator",
		)
		suite.Require().NoError(err)

		acc := suite.Kava.AddNewSigningAccountFromPrivKey(
			"validator",
			key,
			"",
			suite.Kava.ChainID,
		)

		gasLimit := int64(2e5)
		fee := ukava(200)
		req := util.KavaMsgRequest{
			Msgs:      []sdk.Msg{withdrawRewardsMsg},
			GasLimit:  uint64(gasLimit),
			FeeAmount: sdk.NewCoins(fee),
			Memo:      "give me my money",
		}
		res := acc.SignAndBroadcastKavaTx(req)

		_, err = util.WaitForSdkTxCommit(suite.Kava.Tx, res.Result.TxHash, 6*time.Second)
		suite.Require().NoError(err)

		balAfter, err := suite.Kava.Bank.Balance(
			context.Background(),
			&types.QueryBalanceRequest{
				Address: accAddr.String(),
				Denom:   suite.Kava.StakingDenom,
			},
		)
		suite.Require().NoError(err)
		suite.Require().False(balAfter.Balance.IsZero(), "val staking denom balance should be non-zero")

		balIncrease := balAfter.Balance.
			Sub(*balBefore.Balance).
			Add(res.Tx.GetFee()[0]) // Add the fee back to balance to compare actual balances

		queriedRewardsCoins, _ := delegationRewards.Rewards.TruncateDecimal()

		suite.Require().Truef(
			queriedRewardsCoins.AmountOf(suite.Kava.StakingDenom).
				LTE(balIncrease.Amount),
			"claimed rewards should be >= queried delegation rewards, got claimed %s vs queried %s",
			balIncrease.Amount.String(),
			queriedRewardsCoins.AmountOf(suite.Kava.StakingDenom).String(),
		)
	})
}

// unsafeExporter is implemented by key stores that support unsafe export
// of private keys' material.
type unsafeExporter interface {
	// ExportPrivateKeyObject returns a private key in unarmored format.
	ExportPrivateKeyObject(uid string) (cryptotypes.PrivKey, error)
}
