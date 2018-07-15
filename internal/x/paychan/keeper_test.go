package paychan

import (
	"testing"
	//"github.com/stretchr/testify/assert"

	abci "github.com/tendermint/abci/types"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
)

// GetPaychan
//  - gets a paychan if it exists, and not if it doesn't
// setPaychan
//  - sets a paychan
// CreatePaychan
//  - creates a paychan under normal conditions
// ClosePaychan
//  - closes a paychan under normal conditions
// GetPaychans
// paychanKey

func setupMultiStore() (sdk.MultiStore, *sdk.KVStoreKey, *sdk.KVStoreKey) {
	// create db
	db := dbm.NewMemDB()
	// create keys
	authKey := sdk.NewKVStoreKey("authkey")
	paychanKey := sdk.NewKVStoreKey("paychankey")
	// create new multistore around db
	ms := store.NewCommitMultiStore(db) // DB handle plus store key maps
	// register separate stores in the multistore
	ms.MountStoreWithDB(authKey, sdk.StoreTypeIAVL, db) // sets store key map
	ms.MountStoreWithDB(paychanKey, sdk.StoreTypeIAVL, db)
	ms.LoadLatestVersion()

	return ms, authKey, paychanKey
}

func setupCodec() *wire.Codec {
	cdc := wire.NewCodec()
	auth.RegisterBaseAccount(cdc)
	// TODO might need to register paychan struct
	return cdc
}

func TestKeeper(t *testing.T) {
	// Setup

	// create multistore and key
	ms, authKey, paychanKey := setupMultiStore()

	// create and initialise codec(s)
	cdc := setupCodec()

	// create context
	ctx := sdk.NewContext(ms, abci.Header{}, false, nil, log.NewNopLogger())

	// create accountMapper
	accountMapper := auth.NewAccountMapper(cdc, authKey, &auth.BaseAccount{})

	// create coinkeeper
	coinKeeper := bank.NewKeeper(accountMapper)

	// create keeper
	paychanKeeper := NewKeeper(cdc, paychanKey, coinKeeper)

	// Test no paychans exist
	_, exists := paychanKeeper.GetPaychan(ctx, sdk.Address{}, sdk.Address{}, 0)
	if exists {
		t.Error("payment channel found when none exist")
	}

	// Test paychan can be set and get
	p := Paychan{
		Sender:   sdk.Address([]byte("senderAddress")),
		Receiver: sdk.Address([]byte("receiverAddress")),
		Id:       0,
		Balance:  sdk.Coins{{"KVA", 100}},
	}
	paychanKeeper.setPaychan(ctx, p)

	_, exists = paychanKeeper.GetPaychan(ctx, p.Sender, p.Receiver, p.Id)
	if !exists {
		t.Error("payment channel not found")
	}

	// Test create paychan under normal conditions
	senderAddress := sdk.Address([]byte("senderAddress"))
	senderFunds := sdk.Coins{{"KVA", 100}}
	receiverAddress := sdk.Address([]byte("receiverAddress"))
	balance := sdk.Coins{{"KVA", 10}}

	coinKeeper.SetCoins(ctx, senderAddress, senderFunds)

	_, err := paychanKeeper.CreatePaychan(ctx, senderAddress, receiverAddress, balance)
	if err != nil {
		t.Error("unexpected error created payment channel", err)
	}

	p, exists = paychanKeeper.GetPaychan(ctx, senderAddress, receiverAddress, 0)
	if !exists {
		t.Error("payment channel missing")
	}
	if !p.Balance.IsEqual(balance) {
		t.Error("payment channel balance incorrect", p.Balance, balance)
	}
	expectedNewSenderFunds := senderFunds.Minus(balance)
	if !coinKeeper.GetCoins(ctx, senderAddress).IsEqual(expectedNewSenderFunds) {
		t.Error("sender has incorrect balance after paychan creation")
	}

	// Test close paychan under normal conditions
	senderFunds = coinKeeper.GetCoins(ctx, senderAddress)
	receiverAmount := sdk.Coins{{"KVA", 9}}
	_, err = paychanKeeper.ClosePaychan(ctx, senderAddress, receiverAddress, 0, receiverAmount)
	if err != nil {
		t.Error("unexpected error closing payment channel", err)
	}
	// paychan shouldn't exist
	_, exists = paychanKeeper.GetPaychan(ctx, senderAddress, receiverAddress, 0)
	if exists {
		t.Error("payment channel should not exist")
	}
	// sender's funds should have increased
	expectedNewSenderFunds = senderFunds.Plus(balance.Minus(receiverAmount))
	if !coinKeeper.GetCoins(ctx, senderAddress).IsEqual(expectedNewSenderFunds) {
		t.Error("sender has incorrect balance after paychan creation", expectedNewSenderFunds)
	}
	// receiver's funds should have increased
	expectedNewReceiverFunds := receiverAmount // started at zero
	if !coinKeeper.GetCoins(ctx, receiverAddress).IsEqual(expectedNewReceiverFunds) {
		t.Error("receiver has incorrect balance after paychan creation")
	}

}
