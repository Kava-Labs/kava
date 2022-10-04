package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/incentive/keeper"
	kavadisttypes "github.com/kava-labs/kava/x/kavadist/types"
	"github.com/stretchr/testify/require"
)

func TestGetTotalInfrastructureInflation(t *testing.T) {
	blockTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		periods       kavadisttypes.Periods
		wantInflation sdk.Dec
	}{
		{
			"one period",
			kavadisttypes.Periods{
				kavadisttypes.NewPeriod(
					blockTime.Add(-time.Hour),
					blockTime.Add(time.Hour),
					sdk.MustNewDecFromStr("1.000000003022265980"),
				),
			},
			sdk.MustNewDecFromStr("0.000000003022265980").MulInt64(keeper.SecondsPerYear),
		},
		{
			"one period expired",
			kavadisttypes.Periods{
				kavadisttypes.NewPeriod(
					blockTime.Add(-2*time.Hour),
					blockTime.Add(-time.Hour),
					sdk.MustNewDecFromStr("1.000000003022265980"),
				),
			},
			sdk.ZeroDec(),
		},
		{
			"one period in future",
			kavadisttypes.Periods{
				kavadisttypes.NewPeriod(
					blockTime.Add(time.Hour),
					blockTime.Add(2*time.Hour),
					sdk.MustNewDecFromStr("1.000000003022265980"),
				),
			},
			sdk.ZeroDec(),
		},
		{
			"two periods active",
			kavadisttypes.Periods{
				// Additional expired period
				kavadisttypes.NewPeriod(
					blockTime.Add(-2*time.Hour),
					blockTime.Add(-time.Hour),
					sdk.MustNewDecFromStr("1.000000003022265980"),
				),
				// Two active periods
				kavadisttypes.NewPeriod(
					blockTime.Add(-time.Hour),
					blockTime.Add(time.Hour),
					sdk.MustNewDecFromStr("1.000000003022265980"),
				),
				kavadisttypes.NewPeriod(
					blockTime.Add(-2*time.Hour),
					blockTime.Add(2*time.Hour),
					sdk.MustNewDecFromStr("1.0000000095129375"),
				),
				// An additional future period
				kavadisttypes.NewPeriod(
					blockTime.Add(2*time.Hour),
					blockTime.Add(3*time.Hour),
					sdk.MustNewDecFromStr("1.0000000095129375"),
				),
			},
			sdk.MustNewDecFromStr("0.000000003022265980").
				Add(sdk.MustNewDecFromStr("0.0000000095129375")).
				MulInt64(keeper.SecondsPerYear),
		},
	}

	ctx := NewTestContext().WithBlockTime(blockTime)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			inflation := keeper.GetTotalInfrastructureInflation(ctx, tc.periods)
			t.Logf("inflation per year: %s (~%v%%)",
				inflation, inflation.MulInt64(100).RoundInt64())

			require.Equal(t, tc.wantInflation, inflation)
		})
	}
}
