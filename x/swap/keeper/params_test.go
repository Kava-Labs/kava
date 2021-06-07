package keeper_test

import (
	"testing"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/swap/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"
)

func TestParams_SetterAndGetter(t *testing.T) {
	tApp := app.NewTestApp()
	keeper := tApp.GetSwapKeeper()

	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
	params := types.Params{
		Pairs: types.Pairs{
			types.NewPair("usdx", "ukava"),
		},
		SwapFee: sdk.MustNewDecFromStr("0.03"),
	}
	keeper.SetParams(ctx, params)
	assert.Equal(t, keeper.GetParams(ctx), params)

	oldParams := params
	params = types.Params{
		Pairs: types.Pairs{
			types.NewPair("hard", "ukava"),
		},
		SwapFee: sdk.MustNewDecFromStr("0.01"),
	}
	keeper.SetParams(ctx, params)
	assert.NotEqual(t, keeper.GetParams(ctx), oldParams)
	assert.Equal(t, keeper.GetParams(ctx), params)
}

func TestParams_GetPairsWithDenom(t *testing.T) {
	tApp := app.NewTestApp()
	keeper := tApp.GetSwapKeeper()

	testCases := []struct {
		name          string
		denom         string
		pairs         types.Pairs
		expectedPairs types.Pairs
	}{
		{
			name:          "single pair with no match",
			denom:         "bnb",
			pairs:         types.NewPairs(types.NewPair("ukava", "usdx")),
			expectedPairs: types.Pairs{},
		},
		{
			name:          "single pair with match token a",
			denom:         "ukava",
			pairs:         types.NewPairs(types.NewPair("ukava", "usdx")),
			expectedPairs: types.NewPairs(types.NewPair("ukava", "usdx")),
		},
		{
			name:          "single pair with match token b",
			denom:         "usdx",
			pairs:         types.NewPairs(types.NewPair("ukava", "usdx")),
			expectedPairs: types.NewPairs(types.NewPair("ukava", "usdx")),
		},
		{
			name:  "multiple pairs no match",
			denom: "bnb",
			pairs: types.NewPairs(
				types.NewPair("ukava", "usdx"),
				types.NewPair("hard", "usdx"),
			),
			expectedPairs: types.Pairs{},
		},
		{
			name:  "multiple pairs single match token a",
			denom: "ukava",
			pairs: types.NewPairs(
				types.NewPair("ukava", "usdx"),
				types.NewPair("hard", "usdx"),
			),
			expectedPairs: types.NewPairs(
				types.NewPair("ukava", "usdx"),
			),
		},
		{
			name:  "multiple pairs single match token b",
			denom: "usdx",
			pairs: types.NewPairs(
				types.NewPair("ukava", "usdx"),
				types.NewPair("hard", "ukava"),
			),
			expectedPairs: types.NewPairs(
				types.NewPair("ukava", "usdx"),
			),
		},
		{
			name:  "multiple pairs multiple match token a only",
			denom: "ukava",
			pairs: types.NewPairs(
				types.NewPair("ukava", "usdx"),
				types.NewPair("ukava", "hard"),
			),
			expectedPairs: types.NewPairs(
				types.NewPair("ukava", "usdx"),
				types.NewPair("ukava", "hard"),
			),
		},
		{
			name:  "multiple pairs multiple match token b only",
			denom: "usdx",
			pairs: types.NewPairs(
				types.NewPair("ukava", "usdx"),
				types.NewPair("hard", "usdx"),
			),
			expectedPairs: types.NewPairs(
				types.NewPair("ukava", "usdx"),
				types.NewPair("hard", "usdx"),
			),
		},
		{
			name:  "multiple pairs match both tokens",
			denom: "ukava",
			pairs: types.NewPairs(
				types.NewPair("ukava", "usdx"),
				types.NewPair("hard", "usdx"),
				types.NewPair("usdx", "bnb"),
				types.NewPair("btcb", "usdx"),
				types.NewPair("btcb", "ukava"),
			),
			expectedPairs: types.NewPairs(
				types.NewPair("ukava", "usdx"),
				types.NewPair("btcb", "ukava"),
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
			params := types.Params{
				Pairs:   tc.pairs,
				SwapFee: sdk.MustNewDecFromStr("0.03"),
			}
			keeper.SetParams(ctx, params)

			pairs := keeper.GetPairsWithDenom(ctx, tc.denom)
			assert.Equal(t, tc.expectedPairs, pairs)
		})
	}
}
