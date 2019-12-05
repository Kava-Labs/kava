package keeper_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/cdp/types"
	"github.com/kava-labs/kava/x/pricefeed"
)

// How could one reduce the number of params in the test cases. Create a table driven test for each of the 4 add/withdraw collateral/debt?

// These are more like app level tests - I think this is a symptom of having 'ModifyCDP' do a lot. Could be easier for testing purposes to break it down.
func TestKeeper_ModifyCDP(t *testing.T) {
	_, addrs := app.GeneratePrivKeyAddressPairs(1)
	ownerAddr := addrs[0]

	type state struct {
		CDP             types.CDP
		OwnerCoins      sdk.Coins
		GlobalDebt      sdk.Int
		CollateralState types.CollateralState
	}
	type args struct {
		owner              sdk.AccAddress
		collateralDenom    string
		changeInCollateral sdk.Int
		changeInDebt       sdk.Int
	}

	tests := []struct {
		name       string
		priorState state
		price      string
		// also missing CDPModuleParams
		args          args
		expectPass    bool
		expectedState state
	}{
		{
			"addCollateralAndDecreaseDebt",
			state{types.CDP{ownerAddr, "xrp", i(100), i(2)}, cs(c("xrp", 10), c("usdx", 2)), i(2), types.CollateralState{"xrp", i(2)}},
			"10.345",
			args{ownerAddr, "xrp", i(10), i(-1)},
			true,
			state{types.CDP{ownerAddr, "xrp", i(110), i(1)}, cs( /*  0xrp  */ c("usdx", 1)), i(1), types.CollateralState{"xrp", i(1)}},
		},
		{
			"removeTooMuchCollateral",
			state{types.CDP{ownerAddr, "xrp", i(1000), i(200)}, cs(c("xrp", 10), c("usdx", 10)), i(200), types.CollateralState{"xrp", i(200)}},
			"1.00",
			args{ownerAddr, "xrp", i(-601), i(0)},
			false,
			state{types.CDP{ownerAddr, "xrp", i(1000), i(200)}, cs(c("xrp", 10), c("usdx", 10)), i(200), types.CollateralState{"xrp", i(200)}},
		},
		{
			"withdrawTooMuchStableCoin",
			state{types.CDP{ownerAddr, "xrp", i(1000), i(200)}, cs(c("xrp", 10), c("usdx", 10)), i(200), types.CollateralState{"xrp", i(200)}},
			"1.00",
			args{ownerAddr, "xrp", i(0), i(301)},
			false,
			state{types.CDP{ownerAddr, "xrp", i(1000), i(200)}, cs(c("xrp", 10), c("usdx", 10)), i(200), types.CollateralState{"xrp", i(200)}},
		},
		{
			"createCDPAndWithdrawStable",
			state{types.CDP{}, cs(c("xrp", 10), c("usdx", 10)), i(0), types.CollateralState{"xrp", i(0)}},
			"1.00",
			args{ownerAddr, "xrp", i(5), i(2)},
			true,
			state{types.CDP{ownerAddr, "xrp", i(5), i(2)}, cs(c("xrp", 5), c("usdx", 12)), i(2), types.CollateralState{"xrp", i(2)}},
		},
		{
			"emptyCDP",
			state{types.CDP{ownerAddr, "xrp", i(1000), i(200)}, cs(c("xrp", 10), c("usdx", 201)), i(200), types.CollateralState{"xrp", i(200)}},
			"1.00",
			args{ownerAddr, "xrp", i(-1000), i(-200)},
			true,
			state{types.CDP{}, cs(c("xrp", 1010), c("usdx", 1)), i(0), types.CollateralState{"xrp", i(0)}},
		},
		{
			"invalidCollateralType",
			state{types.CDP{}, cs(c("shitcoin", 5000000)), i(0), types.CollateralState{}},
			"0.000001",
			args{ownerAddr, "shitcoin", i(5000000), i(1)}, // ratio of 5:1
			false,
			state{types.CDP{}, cs(c("shitcoin", 5000000)), i(0), types.CollateralState{}},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// setup keeper
			tApp := app.NewTestApp()
			// initialize cdp owner account with coins
			authGen := tApp.NewAuthGenStateFromAccounts([]sdk.AccAddress{ownerAddr}, []sdk.Coins{tc.priorState.OwnerCoins})
			tApp.InitializeFromGenesisStates(authGen)
			// create a context for db access
			ctx := tApp.NewContext(false, abci.Header{})

			// setup store state
			keeper := tApp.GetCDPKeeper()
			keeper.SetParams(ctx, defaultParamsSingle())
			pricefeedKeeper := tApp.GetPriceFeedKeeper()
			ap := pricefeed.AssetParams{
				Assets: []pricefeed.Asset{
					pricefeed.Asset{AssetCode: "xrp", Description: ""},
				},
			}
			pricefeedKeeper.SetAssetParams(ctx, ap)
			_, err := pricefeedKeeper.SetPrice(
				ctx, ownerAddr, "xrp",
				sdk.MustNewDecFromStr(tc.price),
				sdk.NewInt(ctx.BlockHeight()+10))
			if err != nil {
				t.Log("test context height", ctx.BlockHeight())
				t.Log(err)
				t.Log(tc.name)
			}
			err = pricefeedKeeper.SetCurrentPrices(ctx)
			if err != nil {
				t.Log("test context height", ctx.BlockHeight())
				t.Log(err)
				t.Log(tc.name)
			}
			if tc.priorState.CDP.CollateralDenom != "" { // check if the prior CDP should be created or not (see if an empty one was specified)
				keeper.SetCDP(ctx, tc.priorState.CDP)
			}
			keeper.SetGlobalDebt(ctx, tc.priorState.GlobalDebt)
			if tc.priorState.CollateralState.Denom != "" {
				keeper.SetCollateralState(ctx, tc.priorState.CollateralState)
			}

			// call func under test
			err = keeper.ModifyCDP(ctx, tc.args.owner, tc.args.collateralDenom, tc.args.changeInCollateral, tc.args.changeInDebt)

			// get new state for verification
			actualCDP, found := keeper.GetCDP(ctx, tc.args.owner, tc.args.collateralDenom)
			if tc.name == "removeTooMuchCollateral" {
				t.Log(actualCDP.String())
			}
			// check for err
			if tc.expectPass {
				require.NoError(t, err, fmt.Sprint(err))
			} else {
				require.Error(t, err)
			}
			actualGDebt := keeper.GetGlobalDebt(ctx)
			actualCstate, _ := keeper.GetCollateralState(ctx, tc.args.collateralDenom)
			// check state
			require.Equal(t, tc.expectedState.CDP, actualCDP)
			if tc.expectedState.CDP.CollateralDenom == "" { // if the expected CDP is blank, then expect the CDP to have been deleted (hence not found)
				require.False(t, found)
			} else {
				require.True(t, found)
			}
			require.Equal(t, tc.expectedState.GlobalDebt, actualGDebt)
			require.Equal(t, tc.expectedState.CollateralState, actualCstate)
			// check owner balance
			tApp.CheckBalance(t, ctx, ownerAddr, tc.expectedState.OwnerCoins)
		})
	}
}

func TestKeeper_PartialSeizeCDP(t *testing.T) {
	// Setup
	const collateral = "xrp"
	_, addrs := app.GeneratePrivKeyAddressPairs(1)
	testAddr := addrs[0]

	tApp := app.NewTestApp()
	authGenState := tApp.NewAuthGenStateFromAccounts(addrs, []sdk.Coins{cs(c(collateral, 100))})
	tApp.InitializeFromGenesisStates(authGenState)

	ctx := tApp.NewContext(false, abci.Header{})
	keeper := tApp.GetCDPKeeper()

	// setup pricefeed
	pricefeedKeeper := tApp.GetPriceFeedKeeper()
	ap := pricefeed.AssetParams{
		Assets: []pricefeed.Asset{
			pricefeed.Asset{AssetCode: collateral, Description: ""},
		},
	}
	pricefeedKeeper.SetAssetParams(ctx, ap)
	pricefeedKeeper.SetPrice(
		ctx, sdk.AccAddress{}, collateral,
		sdk.MustNewDecFromStr("1.00"),
		i(10))
	require.NoError(t, pricefeedKeeper.SetCurrentPrices(ctx))

	// Create CDP
	keeper.SetParams(ctx, defaultParamsSingle())
	err := keeper.ModifyCDP(ctx, testAddr, collateral, i(10), i(5))
	require.NoError(t, err)
	// Reduce price
	pricefeedKeeper.SetPrice(
		ctx, sdk.AccAddress{}, collateral,
		sdk.MustNewDecFromStr("0.90"),
		i(10))
	require.NoError(t, pricefeedKeeper.SetCurrentPrices(ctx))

	// Seize entire CDP
	err = keeper.PartialSeizeCDP(ctx, testAddr, collateral, i(10), i(5))

	// Check
	require.NoError(t, err)
	_, found := keeper.GetCDP(ctx, testAddr, collateral)
	require.False(t, found)
	collateralState, found := keeper.GetCollateralState(ctx, collateral)
	require.True(t, found)
	require.Equal(t, sdk.ZeroInt(), collateralState.TotalDebt)
}

func TestKeeper_GetCDPs(t *testing.T) {
	// setup keeper
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{})
	keeper := tApp.GetCDPKeeper()
	keeper.SetParams(ctx, defaultParamsMulti())
	// setup CDPs
	_, addrs := app.GeneratePrivKeyAddressPairs(2)
	cdps := types.CDPs{
		{addrs[0], "xrp", i(4000), i(5)},
		{addrs[1], "xrp", i(4000), i(2000)},
		{addrs[0], "btc", i(10), i(20)},
	}
	for _, cdp := range cdps {
		keeper.SetCDP(ctx, cdp)
	}

	// Check nil params returns all CDPs
	returnedCdps, err := keeper.GetCDPs(ctx, "", sdk.Dec{})
	require.NoError(t, err)
	require.Equal(t,
		types.CDPs{
			{addrs[0], "btc", i(10), i(20)},
			{addrs[1], "xrp", i(4000), i(2000)},
			{addrs[0], "xrp", i(4000), i(5)}},
		returnedCdps,
	)
	// Check correct CDPs filtered by collateral and sorted
	returnedCdps, err = keeper.GetCDPs(ctx, "xrp", d("0.00000001"))
	require.NoError(t, err)
	require.Equal(t,
		types.CDPs{
			{addrs[1], "xrp", i(4000), i(2000)},
			{addrs[0], "xrp", i(4000), i(5)}},
		returnedCdps,
	)
	returnedCdps, err = keeper.GetCDPs(ctx, "xrp", sdk.Dec{})
	require.NoError(t, err)
	require.Equal(t,
		types.CDPs{
			{addrs[1], "xrp", i(4000), i(2000)},
			{addrs[0], "xrp", i(4000), i(5)}},
		returnedCdps,
	)
	returnedCdps, err = keeper.GetCDPs(ctx, "xrp", d("0.9"))
	require.NoError(t, err)
	require.Equal(t,
		types.CDPs{
			{addrs[1], "xrp", i(4000), i(2000)}},
		returnedCdps,
	)
	// Check high price returns no CDPs
	returnedCdps, err = keeper.GetCDPs(ctx, "xrp", d("999999999.99"))
	require.NoError(t, err)
	require.Equal(t,
		types.CDPs(nil),
		returnedCdps,
	)
	// Check unauthorized collateral denom returns error
	_, err = keeper.GetCDPs(ctx, "a non existent coin", d("0.34023"))
	require.Error(t, err)
	// Check price without collateral returns error
	_, err = keeper.GetCDPs(ctx, "", d("0.34023"))
	require.Error(t, err)
	// Check deleting a CDP removes it
	keeper.DeleteCDP(ctx, cdps[0])
	returnedCdps, err = keeper.GetCDPs(ctx, "", sdk.Dec{})
	require.NoError(t, err)
	require.Equal(t,
		types.CDPs{
			{addrs[0], "btc", i(10), i(20)},
			{addrs[1], "xrp", i(4000), i(2000)}},
		returnedCdps,
	)
}

func TestKeeper_GetSetDeleteCDP(t *testing.T) {
	// setup keeper, create CDP
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{})
	keeper := tApp.GetCDPKeeper()

	keeper.SetParams(ctx, defaultParamsSingle())
	_, addrs := app.GeneratePrivKeyAddressPairs(1)
	cdp := types.CDP{addrs[0], "xrp", i(412), i(56)}

	// write and read from store
	keeper.SetCDP(ctx, cdp)
	readCDP, found := keeper.GetCDP(ctx, cdp.Owner, cdp.CollateralDenom)

	// check before and after match
	require.True(t, found)
	require.Equal(t, cdp, readCDP)

	// delete auction
	keeper.DeleteCDP(ctx, cdp)

	// check auction does not exist
	_, found = keeper.GetCDP(ctx, cdp.Owner, cdp.CollateralDenom)
	require.False(t, found)
}
func TestKeeper_GetSetGDebt(t *testing.T) {
	// setup keeper, create GDebt
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{})
	keeper := tApp.GetCDPKeeper()
	keeper.SetParams(ctx, defaultParamsSingle())
	gDebt := i(4120000)

	// write and read from store
	keeper.SetGlobalDebt(ctx, gDebt)
	readGDebt := keeper.GetGlobalDebt(ctx)

	// check before and after match
	require.Equal(t, gDebt, readGDebt)
}

func TestKeeper_GetSetCollateralState(t *testing.T) {
	// setup keeper, create CState
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{})
	keeper := tApp.GetCDPKeeper()
	keeper.SetParams(ctx, defaultParamsSingle())
	collateralState := types.CollateralState{"xrp", i(15400)}

	// write and read from store
	keeper.SetCollateralState(ctx, collateralState)
	readCState, found := keeper.GetCollateralState(ctx, collateralState.Denom)

	// check before and after match
	require.Equal(t, collateralState, readCState)
	require.True(t, found)
}
