package metrics

import (
	"github.com/kava-labs/kava/x/metrics/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlocker publishes metrics at the start of each block.
func BeginBlocker(ctx sdk.Context, metrics *types.Metrics) {
	metrics.LatestBlockHeight.Set(float64(ctx.BlockHeight()))
}
