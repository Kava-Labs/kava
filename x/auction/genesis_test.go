package auction_test

import (
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/auction"
	"github.com/kava-labs/kava/x/auction/types"
)

var (
	_, testAddrs = app.GeneratePrivKeyAddressPairs(2)
	testTime     = time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	testAuction  = types.NewCollateralAuction(
		"seller",
		c("lotdenom", 10),
		testTime,
		c("biddenom", 1000),
		types.WeightedAddresses{Addresses: testAddrs, Weights: []sdkmath.Int{sdk.OneInt(), sdk.OneInt()}},
		c("debt", 1000),
	).WithID(3).(types.GenesisAuction)
)

func TestInitGenesis(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		// setup keepers
		tApp := app.NewTestApp()
		ctx := tApp.NewContext(true, tmproto.Header{Height: 1})

		// setup module account
		tApp.FundModuleAccount(ctx, types.ModuleName, testAuction.GetModuleAccountCoins())

		// set up auction genesis state with module account
		auctionGS, err := types.NewGenesisState(
			10,
			types.DefaultParams(),
			[]types.GenesisAuction{testAuction},
		)
		require.NoError(t, err)

		// run init
		keeper := tApp.GetAuctionKeeper()
		require.NotPanics(t, func() {
			auction.InitGenesis(ctx, keeper, tApp.GetBankKeeper(), tApp.GetAccountKeeper(), auctionGS)
		})

		// check state is as expected
		actualID, err := keeper.GetNextAuctionID(ctx)
		require.NoError(t, err)
		require.Equal(t, auctionGS.NextAuctionId, actualID)

		require.Equal(t, auctionGS.Params, keeper.GetParams(ctx))

		genesisAuctions, err := types.UnpackGenesisAuctions(auctionGS.Auctions)
		if err != nil {
			panic(err)
		}

		sort.Slice(genesisAuctions, func(i, j int) bool {
			return genesisAuctions[i].GetID() > genesisAuctions[j].GetID()
		})
		i := 0
		keeper.IterateAuctions(ctx, func(a types.Auction) bool {
			require.Equal(t, genesisAuctions[i], a)
			i++
			return false
		})
	})
	t.Run("invalid (invalid nextAuctionID)", func(t *testing.T) {
		// setup keepers
		tApp := app.NewTestApp()
		ctx := tApp.NewContext(true, tmproto.Header{Height: 1})

		// setup module account
		tApp.FundModuleAccount(ctx, types.ModuleName, testAuction.GetModuleAccountCoins())

		// create invalid genesis
		auctionGS, err := types.NewGenesisState(
			0, // next id < testAuction ID
			types.DefaultParams(),
			[]types.GenesisAuction{testAuction},
		)
		require.NoError(t, err)

		// check init fails
		require.Panics(t, func() {
			auction.InitGenesis(ctx, tApp.GetAuctionKeeper(), tApp.GetBankKeeper(), tApp.GetAccountKeeper(), auctionGS)
		})
	})
	t.Run("invalid (missing mod account coins)", func(t *testing.T) {
		// setup keepers
		tApp := app.NewTestApp()
		ctx := tApp.NewContext(true, tmproto.Header{Height: 1})

		// invalid as there is no module account setup

		// create invalid genesis
		auctionGS, err := types.NewGenesisState(
			10,
			types.DefaultParams(),
			[]types.GenesisAuction{testAuction},
		)
		require.NoError(t, err)

		// check init fails
		require.Panics(t, func() {
			auction.InitGenesis(ctx, tApp.GetAuctionKeeper(), tApp.GetBankKeeper(), tApp.GetAccountKeeper(), auctionGS)
		})
	})
}

func TestExportGenesis(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		tApp := app.NewTestApp()
		ctx := tApp.NewContext(true, tmproto.Header{Height: 1})

		expectedGenesisState := types.DefaultGenesisState()
		auction.InitGenesis(
			ctx,
			tApp.GetAuctionKeeper(),
			tApp.GetBankKeeper(),
			tApp.GetAccountKeeper(),
			expectedGenesisState,
		)

		gs := auction.ExportGenesis(ctx, tApp.GetAuctionKeeper())
		require.Equal(t, expectedGenesisState, gs)
	})
	t.Run("one auction", func(t *testing.T) {
		tApp := app.NewTestApp()
		ctx := tApp.NewContext(true, tmproto.Header{Height: 1})

		expectedGenesisState, err := types.NewGenesisState(
			10,
			types.DefaultParams(),
			[]types.GenesisAuction{testAuction},
		)
		require.NoError(t, err)

		// fund module account to pass InitGenesis validation
		tApp.FundModuleAccount(ctx, types.ModuleName, testAuction.GetModuleAccountCoins())
		auction.InitGenesis(
			ctx,
			tApp.GetAuctionKeeper(),
			tApp.GetBankKeeper(),
			tApp.GetAccountKeeper(),
			expectedGenesisState,
		)

		gs := auction.ExportGenesis(ctx, tApp.GetAuctionKeeper())
		require.Equal(t, expectedGenesisState, gs)
	})
}
