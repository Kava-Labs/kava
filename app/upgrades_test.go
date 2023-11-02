package app_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	"github.com/kava-labs/kava/app"
	incentivetypes "github.com/kava-labs/kava/x/incentive/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"
)

func TestUpgradeCommunityParams_Mainnet(t *testing.T) {
	require.Equal(
		t,
		sdkmath.LegacyZeroDec().String(),
		app.CommunityParams_Mainnet.StakingRewardsPerSecond.String(),
	)

	require.Equal(
		t,
		// Manually confirmed
		"317097.919837645865043125",
		app.CommunityParams_Mainnet.UpgradeTimeSetStakingRewardsPerSecond.String(),
		"mainnet kava per second should be correct",
	)
}

func TestUpgradeCommunityParams_Testnet(t *testing.T) {
	require.Equal(
		t,
		sdkmath.LegacyZeroDec().String(),
		app.CommunityParams_Testnet.StakingRewardsPerSecond.String(),
	)

	require.Equal(
		t,
		// Manually confirmed
		"475646879756.468797564687975646",
		app.CommunityParams_Testnet.UpgradeTimeSetStakingRewardsPerSecond.String(),
		"testnet kava per second should be correct",
	)
}

func TestUpdateValidatorMinimumCommission(t *testing.T) {
	tApp := app.NewTestApp()
	tApp.InitializeFromGenesisStates()
	ctx := tApp.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})

	sk := tApp.GetStakingKeeper()
	stakingParams := sk.GetParams(ctx)
	stakingParams.MinCommissionRate = sdk.ZeroDec()
	sk.SetParams(ctx, stakingParams)

	// Set some validators with varying commission rates

	vals := []struct {
		name              string
		operatorAddr      sdk.ValAddress
		consPriv          *ethsecp256k1.PrivKey
		commissionRateMin sdk.Dec
		commissionRateMax sdk.Dec
		shouldBeUpdated   bool
	}{
		{
			name:              "zero commission rate",
			operatorAddr:      sdk.ValAddress("val0"),
			consPriv:          generateConsKey(t),
			commissionRateMin: sdk.ZeroDec(),
			commissionRateMax: sdk.ZeroDec(),
			shouldBeUpdated:   true,
		},
		{
			name:              "0.01 commission rate",
			operatorAddr:      sdk.ValAddress("val1"),
			consPriv:          generateConsKey(t),
			commissionRateMin: sdk.MustNewDecFromStr("0.01"),
			commissionRateMax: sdk.MustNewDecFromStr("0.01"),
			shouldBeUpdated:   true,
		},
		{
			name:              "0.05 commission rate",
			operatorAddr:      sdk.ValAddress("val2"),
			consPriv:          generateConsKey(t),
			commissionRateMin: sdk.MustNewDecFromStr("0.05"),
			commissionRateMax: sdk.MustNewDecFromStr("0.05"),
			shouldBeUpdated:   false,
		},
		{
			name:              "0.06 commission rate",
			operatorAddr:      sdk.ValAddress("val3"),
			consPriv:          generateConsKey(t),
			commissionRateMin: sdk.MustNewDecFromStr("0.06"),
			commissionRateMax: sdk.MustNewDecFromStr("0.06"),
			shouldBeUpdated:   false,
		},
		{
			name:              "0.5 commission rate",
			operatorAddr:      sdk.ValAddress("val4"),
			consPriv:          generateConsKey(t),
			commissionRateMin: sdk.MustNewDecFromStr("0.5"),
			commissionRateMax: sdk.MustNewDecFromStr("0.5"),
			shouldBeUpdated:   false,
		},
	}

	for _, v := range vals {
		val, err := stakingtypes.NewValidator(
			v.operatorAddr,
			v.consPriv.PubKey(),
			stakingtypes.Description{},
		)
		require.NoError(t, err)
		val.Commission.Rate = v.commissionRateMin
		val.Commission.MaxRate = v.commissionRateMax

		err = sk.SetValidatorByConsAddr(ctx, val)
		require.NoError(t, err)
		sk.SetValidator(ctx, val)
	}

	require.NotPanics(
		t, func() {
			app.UpdateValidatorMinimumCommission(ctx, tApp.App)
		},
	)

	stakingParamsAfter := sk.GetParams(ctx)
	require.Equal(t, stakingParamsAfter.MinCommissionRate, app.ValidatorMinimumCommission)

	// Check that all validators have a commission rate >= 5%
	for _, val := range vals {
		t.Run(val.name, func(t *testing.T) {
			validator, found := sk.GetValidator(ctx, val.operatorAddr)
			require.True(t, found, "validator should be found")

			require.True(
				t,
				validator.GetCommission().GTE(app.ValidatorMinimumCommission),
				"commission rate should be >= 5%",
			)

			require.True(
				t,
				validator.Commission.MaxRate.GTE(app.ValidatorMinimumCommission),
				"commission rate max should be >= 5%, got %s",
				validator.Commission.MaxRate,
			)

			if val.shouldBeUpdated {
				require.Equal(
					t,
					ctx.BlockTime(),
					validator.Commission.UpdateTime,
					"commission update time should be set to block time",
				)
			} else {
				require.Equal(
					t,
					time.Unix(0, 0).UTC(),
					validator.Commission.UpdateTime,
					"commission update time should not be changed -- default value is 0",
				)
			}
		})
	}
}

func TestUpdateIncentiveParams(t *testing.T) {
	tApp := app.NewTestApp()
	tApp.InitializeFromGenesisStates()
	ctx := tApp.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})

	ik := tApp.GetIncentiveKeeper()
	params := ik.GetParams(ctx)

	startPeriod := time.Date(2021, 10, 26, 15, 0, 0, 0, time.UTC)
	endPeriod := time.Date(2022, 10, 26, 15, 0, 0, 0, time.UTC)

	params.EarnRewardPeriods = incentivetypes.MultiRewardPeriods{
		incentivetypes.NewMultiRewardPeriod(
			true,
			"bkava",
			startPeriod,
			endPeriod,
			sdk.NewCoins(
				sdk.NewCoin("ukava", sdk.NewInt(159459)),
			),
		),
	}
	ik.SetParams(ctx, params)

	beforeParams := ik.GetParams(ctx)
	require.Equal(t, params, beforeParams, "initial incentive params should be set")

	// -- UPGRADE
	app.UpdateIncentiveParams(ctx, tApp.App)

	// -- After
	afterParams := ik.GetParams(ctx)

	require.Len(
		t,
		afterParams.EarnRewardPeriods[0].RewardsPerSecond,
		1,
		"bkava earn reward period should only contain 1 coin",
	)
	require.Equal(
		t,
		// Manual calculation of
		// 600,000 * 1000,000 / (365 * 24 * 60 * 60)
		sdk.NewCoin("ukava", sdkmath.NewInt(19025)),
		afterParams.EarnRewardPeriods[0].RewardsPerSecond[0],
		"bkava earn reward period should be updated",
	)

	// Check that other params are not changed
	afterParams.EarnRewardPeriods[0].RewardsPerSecond[0] = beforeParams.EarnRewardPeriods[0].RewardsPerSecond[0]
	require.Equal(
		t,
		beforeParams,
		afterParams,
		"other param values should not be changed",
	)
}

func generateConsKey(
	t *testing.T,
) *ethsecp256k1.PrivKey {
	t.Helper()

	key, err := ethsecp256k1.GenerateKey()
	require.NoError(t, err)

	return key
}
