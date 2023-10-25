package app_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	"github.com/kava-labs/kava/app"
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
		operatorAddr   sdk.ValAddress
		consPriv       *ethsecp256k1.PrivKey
		commissionRate sdk.Dec
	}{
		{
			operatorAddr:   sdk.ValAddress("val0"),
			consPriv:       generateConsKey(t),
			commissionRate: sdk.ZeroDec(),
		},
		{
			operatorAddr:   sdk.ValAddress("val1"),
			consPriv:       generateConsKey(t),
			commissionRate: sdk.MustNewDecFromStr("0.01"),
		},
		{
			operatorAddr:   sdk.ValAddress("val2"),
			consPriv:       generateConsKey(t),
			commissionRate: sdk.MustNewDecFromStr("0.05"),
		},
		{
			operatorAddr:   sdk.ValAddress("val3"),
			consPriv:       generateConsKey(t),
			commissionRate: sdk.MustNewDecFromStr("0.06"),
		},
		{
			operatorAddr:   sdk.ValAddress("val4"),
			consPriv:       generateConsKey(t),
			commissionRate: sdk.MustNewDecFromStr("0.5"),
		},
	}

	for _, v := range vals {
		val, err := stakingtypes.NewValidator(
			v.operatorAddr,
			v.consPriv.PubKey(),
			stakingtypes.Description{},
		)
		require.NoError(t, err)
		val.Commission.Rate = v.commissionRate

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
	count := 0
	sk.IterateValidators(ctx, func(index int64, validator stakingtypes.ValidatorI) (stop bool) {
		require.True(
			t,
			validator.GetCommission().GTE(app.ValidatorMinimumCommission),
			"commission rate should be >= 5%",
		)

		count++
		return false
	})

	require.Equal(
		t,
		len(vals)+1, // InitializeFromGenesisStates adds a validator
		count,
		"validator count should match test validators",
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
