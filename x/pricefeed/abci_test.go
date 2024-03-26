package pricefeed_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/pricefeed"
	"github.com/kava-labs/kava/x/pricefeed/keeper"
	"github.com/kava-labs/kava/x/pricefeed/testutil"
)

func TestEndBlocker_UpdatesMultipleMarkets(t *testing.T) {
	testutil.SetCurrentPrices_PriceCalculations(t, func(ctx sdk.Context, keeper keeper.Keeper) {
		pricefeed.EndBlocker(ctx, keeper)
	})

	testutil.SetCurrentPrices_EventEmission(t, func(ctx sdk.Context, keeper keeper.Keeper) {
		pricefeed.EndBlocker(ctx, keeper)
	})
}
