package auction_test

import (
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/auction"
)

var _, testAddrs = app.GeneratePrivKeyAddressPairs(2)
var testTime = time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
var testAuction = auction.NewCollateralAuction(
	"seller",
	c("lotdenom", 10),
	testTime,
	c("biddenom", 1000),
	auction.WeightedAddresses{Addresses: testAddrs, Weights: []sdk.Int{sdk.OneInt(), sdk.OneInt()}},
	c("debt", 1000),
).WithID(3).(auction.GenesisAuction)

func TestInitGenesis(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		// setup keepers
		tApp := app.NewTestApp()
		keeper := tApp.GetAuctionKeeper()
		ctx := tApp.NewContext(true, abci.Header{})
		// setup module account
		bankKeeper := tApp.GetBankKeeper()
		moduleAcc := accountKeeper.GetModuleAccount(ctx, auction.ModuleName)
		require.NoError(t, moduleAcc.SetCoins(testAuction.GetModuleAccountCoins()))
		accountKeeper.SetModuleAccount(ctx, moduleAcc)

		// create genesis
		gs := auction.NewGenesisState(
			10,
			auction.DefaultParams(),
			auction.GenesisAuctions{testAuction},
		)

		// run init
		require.NotPanics(t, func() {
			auction.InitGenesis(ctx, keeper, gs)
		})

		// check state is as expected
		actualID, err := keeper.GetNextAuctionID(ctx)
		require.NoError(t, err)
		require.Equal(t, gs.NextAuctionID, actualID)

		require.Equal(t, gs.Params, keeper.GetParams(ctx))

		// TODO is there a nicer way of comparing state?
		sort.Slice(gs.Auctions, func(i, j int) bool {
			return gs.Auctions[i].GetID() > gs.Auctions[j].GetID()
		})
		i := 0
		keeper.IterateAuctions(ctx, func(a auction.Auction) bool {
			require.Equal(t, gs.Auctions[i], a)
			i++
			return false
		})
	})
	t.Run("invalid (invalid nextAuctionID)", func(t *testing.T) {
		// setup keepers
		tApp := app.NewTestApp()
		ctx := tApp.NewContext(true, abci.Header{})

		// create invalid genesis
		gs := auction.NewGenesisState(
			0, // next id < testAuction ID
			auction.DefaultParams(),
			auction.GenesisAuctions{testAuction},
		)

		// check init fails
		require.Panics(t, func() {
			auction.InitGenesis(ctx, tApp.GetAuctionKeeper(), gs)
		})
	})
	t.Run("invalid (missing mod account coins)", func(t *testing.T) {
		// setup keepers
		tApp := app.NewTestApp()
		ctx := tApp.NewContext(true, abci.Header{})

		// create invalid genesis
		gs := auction.NewGenesisState(
			10,
			auction.DefaultParams(),
			auction.GenesisAuctions{testAuction},
		)
		// invalid as there is no module account setup

		// check init fails
		require.Panics(t, func() {
			auction.InitGenesis(ctx, tApp.GetAuctionKeeper(), gs)
		})
	})
}

func TestExportGenesis(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		// setup state
		tApp := app.NewTestApp()
		ctx := tApp.NewContext(true, abci.Header{})
		tApp.InitializeFromGenesisStates()

		// export
		gs := auction.ExportGenesis(ctx, tApp.GetAuctionKeeper())

		// check state matches
		require.Equal(t, auction.DefaultGenesisState(), gs)
	})
	t.Run("one auction", func(t *testing.T) {
		// setup state
		tApp := app.NewTestApp()
		ctx := tApp.NewContext(true, abci.Header{})
		tApp.InitializeFromGenesisStates()
		tApp.GetAuctionKeeper().SetAuction(ctx, testAuction)

		// export
		gs := auction.ExportGenesis(ctx, tApp.GetAuctionKeeper())

		// check state matches
		expectedGenesisState := auction.DefaultGenesisState()
		expectedGenesisState.Auctions = append(expectedGenesisState.Auctions, testAuction)
		require.Equal(t, expectedGenesisState, gs)
	})
}
