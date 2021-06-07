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
		AllowedPools: types.AllowedPools{
			types.NewAllowedPool("usdx", "ukava"),
		},
		SwapFee: sdk.MustNewDecFromStr("0.03"),
	}
	keeper.SetParams(ctx, params)
	assert.Equal(t, keeper.GetParams(ctx), params)

	oldParams := params
	params = types.Params{
		AllowedPools: types.AllowedPools{
			types.NewAllowedPool("hard", "ukava"),
		},
		SwapFee: sdk.MustNewDecFromStr("0.01"),
	}
	keeper.SetParams(ctx, params)
	assert.NotEqual(t, keeper.GetParams(ctx), oldParams)
	assert.Equal(t, keeper.GetParams(ctx), params)
}
