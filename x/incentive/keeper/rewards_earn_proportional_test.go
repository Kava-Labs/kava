package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/incentive/keeper"
	"github.com/kava-labs/kava/x/incentive/types"
	"github.com/stretchr/testify/require"
)

func TestGetProportionalRewardPeriod(t *testing.T) {
	tests := []struct {
		name                  string
		giveRewardPeriod      types.MultiRewardPeriod
		giveTotalBkavaSupply  sdk.Int
		giveSingleBkavaSupply sdk.Int
		wantRewardsPerSecond  sdk.Coins
	}{
		{
			"full amount",
			types.NewMultiRewardPeriod(
				true,
				"",
				time.Time{},
				time.Time{},
				cs(c("ukava", 100), c("hard", 200)),
			),
			i(100),
			i(100),
			cs(c("ukava", 100), c("hard", 200)),
		},
		{
			"half amount",
			types.NewMultiRewardPeriod(
				true,
				"",
				time.Time{},
				time.Time{},
				cs(c("ukava", 100), c("hard", 200)),
			),
			i(100),
			i(50),
			cs(c("ukava", 50), c("hard", 100)),
		},
		{
			"rounded down",
			types.NewMultiRewardPeriod(
				true,
				"",
				time.Time{},
				time.Time{},
				cs(c("ukava", 100), c("hard", 200)),
			),
			i(1000),
			i(1),
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newRewardPeriod := keeper.GetProportionalRewardPeriod(
				tt.giveRewardPeriod,
				tt.giveTotalBkavaSupply,
				tt.giveSingleBkavaSupply,
			)

			require.Equal(t, tt.wantRewardsPerSecond, newRewardPeriod.RewardsPerSecond)
		})
	}
}
