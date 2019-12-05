package cdp_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/cdp"
	"github.com/kava-labs/kava/x/pricefeed"
)

func TestApp_CreateModifyDeleteCDP(t *testing.T) {
	// Setup
	tApp := app.NewTestApp()
	privKeys, addrs := app.GeneratePrivKeyAddressPairs(1)
	testAddr := addrs[0]
	testPrivKey := privKeys[0]
	tApp.InitializeFromGenesisStates(
		tApp.NewAuthGenStateFromAccounts(addrs, []sdk.Coins{cs(c("xrp", 100))}),
	)
	// check balance
	ctx := tApp.NewContext(false, abci.Header{})
	tApp.CheckBalance(t, ctx, testAddr, cs(c("xrp", 100)))

	// setup cdp keeper
	keeper := tApp.GetCDPKeeper()
	params := cdp.CdpParams{
		GlobalDebtLimit: sdk.NewInt(100000),
		CollateralParams: []cdp.CollateralParams{
			{
				Denom:            "xrp",
				LiquidationRatio: sdk.MustNewDecFromStr("1.5"),
				DebtLimit:        sdk.NewInt(10000),
			},
		},
		StableDenoms: []string{"usdx"},
	}
	keeper.SetParams(ctx, params)
	keeper.SetGlobalDebt(ctx, sdk.NewInt(0))
	// setup pricefeed
	pfKeeper := tApp.GetPriceFeedKeeper()
	ap := pricefeed.AssetParams{
		Assets: []pricefeed.Asset{pricefeed.Asset{AssetCode: "xrp", Description: ""}},
	}
	pfKeeper.SetAssetParams(ctx, ap)
	pfKeeper.SetPrice(
		ctx, sdk.AccAddress{}, "xrp",
		sdk.MustNewDecFromStr("1.00"),
		sdk.NewInt(10))
	pfKeeper.SetCurrentPrices(ctx)
	tApp.EndBlock(abci.RequestEndBlock{})
	tApp.Commit()

	// Create CDP
	msgs := []sdk.Msg{cdp.NewMsgCreateOrModifyCDP(testAddr, "xrp", i(10), i(5))}
	simapp.SignCheckDeliver(t, tApp.Codec(), tApp.BaseApp, abci.Header{Height: tApp.LastBlockHeight() + 1}, msgs, []uint64{0}, []uint64{0}, true, true, testPrivKey)

	// check balance
	ctx = tApp.NewContext(true, abci.Header{})
	tApp.CheckBalance(t, ctx, testAddr, cs(c("usdx", 5), c("xrp", 90)))

	// Modify CDP
	msgs = []sdk.Msg{cdp.NewMsgCreateOrModifyCDP(testAddr, "xrp", i(40), i(5))}
	simapp.SignCheckDeliver(t, tApp.Codec(), tApp.BaseApp, abci.Header{Height: tApp.LastBlockHeight() + 1}, msgs, []uint64{0}, []uint64{1}, true, true, testPrivKey)

	// check balance
	ctx = tApp.NewContext(true, abci.Header{})
	tApp.CheckBalance(t, ctx, testAddr, cs(c("usdx", 10), c("xrp", 50)))

	// Delete CDP
	msgs = []sdk.Msg{cdp.NewMsgCreateOrModifyCDP(testAddr, "xrp", i(-50), i(-10))}
	simapp.SignCheckDeliver(t, tApp.Codec(), tApp.BaseApp, abci.Header{Height: tApp.LastBlockHeight() + 1}, msgs, []uint64{0}, []uint64{2}, true, true, testPrivKey)

	// check balance
	ctx = tApp.NewContext(true, abci.Header{})
	tApp.CheckBalance(t, ctx, testAddr, cs(c("xrp", 100)))
}
