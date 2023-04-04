package ante_test

import (
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/app/ante"
)

func mustParseDecCoins(value string) sdk.DecCoins {
	coins, err := sdk.ParseDecCoins(strings.ReplaceAll(value, ";", ","))
	if err != nil {
		panic(err)
	}

	return coins
}

func TestEvmMinGasFilter(t *testing.T) {
	tApp := app.NewTestApp()
	handler := ante.NewEvmMinGasFilter(tApp.GetEvmKeeper())

	ctx := tApp.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})
	tApp.GetEvmKeeper().SetParams(ctx, evmtypes.Params{
		EvmDenom: "akava",
	})

	testCases := []struct {
		name                 string
		minGasPrices         sdk.DecCoins
		expectedMinGasPrices sdk.DecCoins
	}{
		{
			"no min gas prices",
			mustParseDecCoins(""),
			mustParseDecCoins(""),
		},
		{
			"zero ukava gas price",
			mustParseDecCoins("0ukava"),
			mustParseDecCoins("0ukava"),
		},
		{
			"non-zero ukava gas price",
			mustParseDecCoins("0.001ukava"),
			mustParseDecCoins("0.001ukava"),
		},
		{
			"zero ukava gas price, min akava price",
			mustParseDecCoins("0ukava;100000akava"),
			mustParseDecCoins("0ukava"), // akava is removed
		},
		{
			"zero ukava gas price, min akava price, other token",
			mustParseDecCoins("0ukava;100000akava;0.001other"),
			mustParseDecCoins("0ukava;0.001other"), // akava is removed
		},
		{
			"non-zero ukava gas price, min akava price",
			mustParseDecCoins("0.25ukava;100000akava;0.001other"),
			mustParseDecCoins("0.25ukava;0.001other"), // akava is removed
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := tApp.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})

			ctx = ctx.WithMinGasPrices(tc.minGasPrices)
			mmd := MockAnteHandler{}

			_, err := handler.AnteHandle(ctx, nil, false, mmd.AnteHandle)
			require.NoError(t, err)
			require.True(t, mmd.WasCalled)

			assert.NoError(t, mmd.CalledCtx.MinGasPrices().Validate())
			assert.Equal(t, tc.expectedMinGasPrices, mmd.CalledCtx.MinGasPrices())
		})
	}
}
