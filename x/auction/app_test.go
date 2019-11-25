package auction

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/mock"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
)

// TestApp contans several basic integration tests of creating an auction, placing a bid, and the auction closing.

func TestApp_ForwardAuction(t *testing.T) {
	// Setup
	mapp, keeper, addresses, privKeys := setUpMockApp()
	seller := addresses[0]
	//sellerKey := privKeys[0]
	buyer := addresses[1]
	buyerKey := privKeys[1]

	// Create a block where an auction is started (lot: 20 t1, initialBid: 0 t2)
	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx := mapp.BaseApp.NewContext(false, header)                                                          // make sure first arg is false, otherwise no db writes
	keeper.StartForwardAuction(ctx, seller, sdk.NewInt64Coin("token1", 20), sdk.NewInt64Coin("token2", 0)) // lot, initialBid
	mapp.EndBlock(abci.RequestEndBlock{})
	mapp.Commit()

	// Check seller's coins have decreased
	mock.CheckBalance(t, mapp, seller, sdk.NewCoins(sdk.NewInt64Coin("token1", 80), sdk.NewInt64Coin("token2", 100)))

	// Deliver a block that contains a PlaceBid tx (bid: 10 t2, lot: same as starting)
	msgs := []sdk.Msg{NewMsgPlaceBid(0, buyer, sdk.NewInt64Coin("token2", 10), sdk.NewInt64Coin("token1", 20))} // bid, lot
	header = abci.Header{Height: mapp.LastBlockHeight() + 1}
	mock.SignCheckDeliver(t, mapp.Cdc, mapp.BaseApp, header, msgs, []uint64{1}, []uint64{0}, true, true, buyerKey) // account number for the buyer account is 1

	// Check buyer's coins have decreased
	mock.CheckBalance(t, mapp, buyer, sdk.NewCoins(sdk.NewInt64Coin("token1", 100), sdk.NewInt64Coin("token2", 90)))
	// Check seller's coins have increased
	mock.CheckBalance(t, mapp, seller, sdk.NewCoins(sdk.NewInt64Coin("token1", 80), sdk.NewInt64Coin("token2", 110)))

	// Deliver empty blocks until the auction should be closed (bid placed on block 3)
	// TODO is there a way of skipping ahead? This takes a while and prints a lot.
	for h := mapp.LastBlockHeight() + 1; h < int64(DefaultMaxBidDuration)+4; h++ {
		mapp.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: h}})
		mapp.EndBlock(abci.RequestEndBlock{Height: h})
		mapp.Commit()
	}
	// Check buyer's coins increased
	mock.CheckBalance(t, mapp, buyer, sdk.NewCoins(sdk.NewInt64Coin("token1", 120), sdk.NewInt64Coin("token2", 90)))
}

func TestApp_ReverseAuction(t *testing.T) {
	// Setup
	mapp, keeper, addresses, privKeys := setUpMockApp()
	seller := addresses[0]
	sellerKey := privKeys[0]
	buyer := addresses[1]
	//buyerKey := privKeys[1]

	// Create a block where an auction is started
	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx := mapp.BaseApp.NewContext(false, header)
	keeper.StartReverseAuction(ctx, buyer, sdk.NewInt64Coin("token1", 20), sdk.NewInt64Coin("token2", 99)) // buyer, bid, initialLot
	mapp.EndBlock(abci.RequestEndBlock{})
	mapp.Commit()

	// Check buyer's coins have decreased
	mock.CheckBalance(t, mapp, buyer, sdk.NewCoins(sdk.NewInt64Coin("token1", 100), sdk.NewInt64Coin("token2", 1)))

	// Deliver a block that contains a PlaceBid tx
	msgs := []sdk.Msg{NewMsgPlaceBid(0, seller, sdk.NewInt64Coin("token1", 20), sdk.NewInt64Coin("token2", 10))} // bid, lot
	header = abci.Header{Height: mapp.LastBlockHeight() + 1}
	mock.SignCheckDeliver(t, mapp.Cdc, mapp.BaseApp, header, msgs, []uint64{0}, []uint64{0}, true, true, sellerKey)

	// Check seller's coins have decreased
	mock.CheckBalance(t, mapp, seller, sdk.NewCoins(sdk.NewInt64Coin("token1", 80), sdk.NewInt64Coin("token2", 100)))
	// Check buyer's coins have increased
	mock.CheckBalance(t, mapp, buyer, sdk.NewCoins(sdk.NewInt64Coin("token1", 120), sdk.NewInt64Coin("token2", 90)))

	// Deliver empty blocks until the auction should be closed (bid placed on block 3)
	for h := mapp.LastBlockHeight() + 1; h < int64(DefaultMaxBidDuration)+4; h++ {
		mapp.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: h}})
		mapp.EndBlock(abci.RequestEndBlock{Height: h})
		mapp.Commit()
	}

	// Check seller's coins increased
	mock.CheckBalance(t, mapp, seller, sdk.NewCoins(sdk.NewInt64Coin("token1", 80), sdk.NewInt64Coin("token2", 110)))
}
func TestApp_ForwardReverseAuction(t *testing.T) {
	// Setup
	mapp, keeper, addresses, privKeys := setUpMockApp()
	seller := addresses[0]
	//sellerKey := privKeys[0]
	buyer := addresses[1]
	buyerKey := privKeys[1]
	recipient := addresses[2]

	// Create a block where an auction is started
	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx := mapp.BaseApp.NewContext(false, header)
	keeper.StartForwardReverseAuction(ctx, seller, sdk.NewInt64Coin("token1", 20), sdk.NewInt64Coin("token2", 50), recipient) // seller, lot, maxBid, otherPerson
	mapp.EndBlock(abci.RequestEndBlock{})
	mapp.Commit()

	// Check seller's coins have decreased
	mock.CheckBalance(t, mapp, seller, sdk.NewCoins(sdk.NewInt64Coin("token1", 80), sdk.NewInt64Coin("token2", 100)))

	// Deliver a block that contains a PlaceBid tx
	msgs := []sdk.Msg{NewMsgPlaceBid(0, buyer, sdk.NewInt64Coin("token2", 50), sdk.NewInt64Coin("token1", 15))} // bid, lot
	header = abci.Header{Height: mapp.LastBlockHeight() + 1}
	mock.SignCheckDeliver(t, mapp.Cdc, mapp.BaseApp, header, msgs, []uint64{1}, []uint64{0}, true, true, buyerKey)

	// Check bidder's coins have decreased
	mock.CheckBalance(t, mapp, buyer, sdk.NewCoins(sdk.NewInt64Coin("token1", 100), sdk.NewInt64Coin("token2", 50)))
	// Check seller's coins have increased
	mock.CheckBalance(t, mapp, seller, sdk.NewCoins(sdk.NewInt64Coin("token1", 80), sdk.NewInt64Coin("token2", 150)))
	// Check "recipient" has received coins
	mock.CheckBalance(t, mapp, recipient, sdk.NewCoins(sdk.NewInt64Coin("token1", 105), sdk.NewInt64Coin("token2", 100)))

	// Deliver empty blocks until the auction should be closed (bid placed on block 3)
	for h := mapp.LastBlockHeight() + 1; h < int64(DefaultMaxBidDuration)+4; h++ {
		mapp.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: h}})
		mapp.EndBlock(abci.RequestEndBlock{Height: h})
		mapp.Commit()
	}

	// Check buyer's coins increased
	mock.CheckBalance(t, mapp, buyer, sdk.NewCoins(sdk.NewInt64Coin("token1", 115), sdk.NewInt64Coin("token2", 50)))
}

func setUpMockApp() (*mock.App, Keeper, []sdk.AccAddress, []crypto.PrivKey) {
	// Create uninitialized mock app
	mapp := mock.NewApp()

	// Register codecs
	RegisterCodec(mapp.Cdc)

	// Create keepers
	keyAuction := sdk.NewKVStoreKey("auction")
	blacklistedAddrs := make(map[string]bool)
	bankKeeper := bank.NewBaseKeeper(mapp.AccountKeeper, mapp.ParamsKeeper.Subspace(bank.DefaultParamspace), bank.DefaultCodespace, blacklistedAddrs)
	auctionKeeper := NewKeeper(mapp.Cdc, bankKeeper, keyAuction, mapp.ParamsKeeper.Subspace(DefaultParamspace))

	// Register routes
	mapp.Router().AddRoute("auction", NewHandler(auctionKeeper))

	// Add endblocker
	mapp.SetEndBlocker(
		func(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
			EndBlocker(ctx, auctionKeeper)
			return abci.ResponseEndBlock{}
		},
	)
	// Mount and load the stores
	err := mapp.CompleteSetup(keyAuction)
	if err != nil {
		panic("mock app setup failed")
	}

	// Create a bunch (ie 10) of pre-funded accounts to use for tests
	genAccs, addrs, _, privKeys := mock.CreateGenAccounts(10, sdk.NewCoins(sdk.NewInt64Coin("token1", 100), sdk.NewInt64Coin("token2", 100)))
	mock.SetGenesis(mapp, genAccs)

	return mapp, auctionKeeper, addrs, privKeys
}
