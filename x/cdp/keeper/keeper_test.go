package keeper

import (
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/kava-labs/kava/x/cdp/types"
	"github.com/kava-labs/kava/x/pricefeed"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"
)

// How could one reduce the number of params in the test cases. Create a table driven test for each of the 4 add/withdraw collateral/debt?

// These are more like app level tests - I think this is a symptom of having 'ModifyCDP' do a lot. Could be easier for testing purposes to break it down.
func TestKeeper_ModifyCDP(t *testing.T) {
	_, addrs := mock.GeneratePrivKeyAddressPairs(1)
	ownerAddr := addrs[0]

	type state struct { // TODO this allows invalid state to be set up, should it?
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
			mapp, keeper, _, _ := setUpMockAppWithoutGenesis()
			// initialize cdp owner account with coins
			genAcc := auth.BaseAccount{
				Address: ownerAddr,
				Coins:   tc.priorState.OwnerCoins,
			}
			mock.SetGenesis(mapp, []authexported.Account{&genAcc})
			// create a new context
			header := abci.Header{Height: mapp.LastBlockHeight() + 1, Time: tmtime.Now()}
			mapp.BeginBlock(abci.RequestBeginBlock{Header: header})
			ctx := mapp.BaseApp.NewContext(false, header)
			keeper.SetParams(ctx, defaultParamsSingle())
			// setup store state
			ap := pricefeed.Params{
				Markets: []pricefeed.Market{
					pricefeed.Market{
						MarketID: "xrp", BaseAsset: "xrp",
						QuoteAsset: "usd", Oracles: pricefeed.Oracles{}, Active: true},
				},
			}
			keeper.pricefeedKeeper.SetParams(ctx, ap)
			_, err := keeper.pricefeedKeeper.SetPrice(
				ctx, ownerAddr, "xrp",
				sdk.MustNewDecFromStr(tc.price),
				header.Time.Add(time.Hour*1))
			if err != nil {
				t.Log("test context height", ctx.BlockHeight())
				t.Log(err)
				t.Log(tc.name)
			}
			err = keeper.pricefeedKeeper.SetCurrentPrices(ctx, "xrp")
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
				keeper.setCollateralState(ctx, tc.priorState.CollateralState)
			}

			// call func under test
			err = keeper.ModifyCDP(ctx, tc.args.owner, tc.args.collateralDenom, tc.args.changeInCollateral, tc.args.changeInDebt)
			mapp.EndBlock(abci.RequestEndBlock{})
			mapp.Commit()

			// get new state for verification
			actualCDP, found := keeper.GetCDP(ctx, tc.args.owner, tc.args.collateralDenom)

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
			mock.CheckBalance(t, mapp, ownerAddr, tc.expectedState.OwnerCoins)
		})
	}
}

// TODO change to table driven test to test more test cases
func TestKeeper_PartialSeizeCDP(t *testing.T) {
	// Setup
	const collateral = "xrp"
	mapp, keeper, _, _ := setUpMockAppWithoutGenesis()
	genAccs, addrs, _, _ := mock.CreateGenAccounts(1, cs(c(collateral, 100)))
	testAddr := addrs[0]
	mock.SetGenesis(mapp, genAccs)
	// setup pricefeed
	header := abci.Header{Height: mapp.LastBlockHeight() + 1, Time: tmtime.Now()}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx := mapp.BaseApp.NewContext(false, header)
	keeper.SetParams(ctx, defaultParamsSingle())
	ap := pricefeed.Params{
		Markets: []pricefeed.Market{
			pricefeed.Market{
				MarketID: "xrp", BaseAsset: "xrp",
				QuoteAsset: "usd", Oracles: pricefeed.Oracles{}, Active: true},
		},
	}
	keeper.pricefeedKeeper.SetParams(ctx, ap)
	keeper.pricefeedKeeper.SetPrice(
		ctx, sdk.AccAddress{}, collateral,
		sdk.MustNewDecFromStr("1.00"),
		header.Time.Add(time.Hour*1))
	keeper.pricefeedKeeper.SetCurrentPrices(ctx, collateral)
	// Create CDP
	keeper.SetGlobalDebt(ctx, i(0))
	err := keeper.ModifyCDP(ctx, testAddr, collateral, i(10), i(5))
	require.NoError(t, err)
	// Reduce price
	keeper.pricefeedKeeper.SetPrice(
		ctx, sdk.AccAddress{}, collateral,
		sdk.MustNewDecFromStr("0.90"),
		header.Time.Add(time.Hour*1))
	keeper.pricefeedKeeper.SetCurrentPrices(ctx, collateral)

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
	mapp, keeper, _, _ := setUpMockAppWithoutGenesis()
	mock.SetGenesis(mapp, []authexported.Account(nil))
	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx := mapp.BaseApp.NewContext(false, header)
	keeper.SetParams(ctx, defaultParamsMulti())
	// setup CDPs
	_, addrs := mock.GeneratePrivKeyAddressPairs(2)
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
	keeper.deleteCDP(ctx, cdps[0])
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
	mapp, keeper, _, _ := setUpMockAppWithoutGenesis()
	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx := mapp.BaseApp.NewContext(false, header)
	keeper.SetParams(ctx, defaultParamsSingle())
	_, addrs := mock.GeneratePrivKeyAddressPairs(1)
	cdp := types.CDP{addrs[0], "xrp", i(412), i(56)}

	// write and read from store
	keeper.SetCDP(ctx, cdp)
	readCDP, found := keeper.GetCDP(ctx, cdp.Owner, cdp.CollateralDenom)

	// check before and after match
	require.True(t, found)
	require.Equal(t, cdp, readCDP)

	// delete auction
	keeper.deleteCDP(ctx, cdp)

	// check auction does not exist
	_, found = keeper.GetCDP(ctx, cdp.Owner, cdp.CollateralDenom)
	require.False(t, found)
}
func TestKeeper_GetSetGDebt(t *testing.T) {
	// setup keeper, create GDebt
	mapp, keeper, _, _ := setUpMockAppWithoutGenesis()
	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx := mapp.BaseApp.NewContext(false, header)
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
	mapp, keeper, _, _ := setUpMockAppWithoutGenesis()
	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx := mapp.BaseApp.NewContext(false, header)
	keeper.SetParams(ctx, defaultParamsSingle())
	collateralState := types.CollateralState{"xrp", i(15400)}

	// write and read from store
	keeper.setCollateralState(ctx, collateralState)
	readCState, found := keeper.GetCollateralState(ctx, collateralState.Denom)

	// check before and after match
	require.Equal(t, collateralState, readCState)
	require.True(t, found)
}
