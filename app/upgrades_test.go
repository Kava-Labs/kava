package app_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/kava-labs/kava/app"
	"github.com/stretchr/testify/require"
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
