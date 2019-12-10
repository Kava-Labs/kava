package cdp_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/cdp"
)

func TestApp_CreateModifyDeleteCDP(t *testing.T) {
	// Setup
	tApp := app.NewTestApp()
	privKeys, addrs := app.GeneratePrivKeyAddressPairs(1)
	testAddr := addrs[0]
	testPrivKey := privKeys[0]
	tApp.InitializeFromGenesisStates(
		app.NewAuthGenState(addrs, []sdk.Coins{cs(c("xrp", 100))}),
		NewPFGenState("xrp", d("1.00")),
		NewCDPGenState("xrp", d("1.5")),
	)
	// check balance
	ctx := tApp.NewContext(false, abci.Header{})
	tApp.CheckBalance(t, ctx, testAddr, cs(c("xrp", 100)))

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
