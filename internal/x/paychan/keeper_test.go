package paychan

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	//"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"
	"testing"
)

func TestKeeper(t *testing.T) {

	t.Run("CreateChannel", func(t *testing.T) {

		//
		////// SETUP
		// create basic mock app
		ctx, coinKeeper, channelKeeper, addrs, genAccFunding := createMockApp()

		sender := addrs[0]
		receiver := addrs[1]
		coins := sdk.Coins{sdk.NewCoin("KVA", 10)}

		//
		////// ACTION
		_, err := channelKeeper.CreateChannel(ctx, sender, receiver, coins)

		//
		////// CHECK RESULTS
		assert.Nil(t, err)
		// channel exists with correct attributes
		channelID := ChannelID(0) // should be 0 as first channel
		expectedChan := Channel{
			ID:           channelID,
			Participants: [2]sdk.AccAddress{sender, receiver},
			Coins:        coins,
		}
		createdChan, _ := channelKeeper.getChannel(ctx, channelID)
		assert.Equal(t, expectedChan, createdChan)
		// check coins deducted from sender
		assert.Equal(t, genAccFunding.Minus(coins), coinKeeper.GetCoins(ctx, sender))
		// check no coins deducted from receiver
		assert.Equal(t, genAccFunding, coinKeeper.GetCoins(ctx, receiver))
		// check next chan id
		assert.Equal(t, ChannelID(1), channelKeeper.getNewChannelID(ctx))
	})

	t.Run("ReceiverCloseChannel", func(t *testing.T) {
		// SETUP
		ctx, coinKeeper, channelKeeper, addrs, genAccFunding := createMockApp()

		sender := addrs[0]
		receiver := addrs[1]
		coins := sdk.Coins{sdk.NewCoin("KVA", 10)}

		// create new channel
		channelID := ChannelID(0) // should be 0 as first channel
		channel := Channel{
			ID:           channelID,
			Participants: [2]sdk.AccAddress{sender, receiver},
			Coins:        coins,
		}
		channelKeeper.setChannel(ctx, channel)

		// create closing update
		payouts := Payouts{
			{sender, sdk.Coins{sdk.NewCoin("KVA", 3)}},
			{receiver, sdk.Coins{sdk.NewCoin("KVA", 7)}},
		}
		update := Update{
			ChannelID: channelID,
			Payouts:   payouts,
			Sigs:      [1]crypto.Signature{},
		}
		// Set empty submittedUpdatesQueue TODO work out proper genesis initialisation
		channelKeeper.setSubmittedUpdatesQueue(ctx, SubmittedUpdatesQueue{})

		// ACTION
		_, err := channelKeeper.CloseChannelByReceiver(ctx, update)

		// CHECK RESULTS
		// no error
		assert.Nil(t, err)
		// coins paid out
		senderPayout, _ := payouts.Get(sender)
		assert.Equal(t, genAccFunding.Plus(senderPayout), coinKeeper.GetCoins(ctx, sender))
		receiverPayout, _ := payouts.Get(receiver)
		assert.Equal(t, genAccFunding.Plus(receiverPayout), coinKeeper.GetCoins(ctx, receiver))
		// channel deleted
		_, found := channelKeeper.getChannel(ctx, channelID)
		assert.False(t, found)

	})

	t.Run("SenderInitCloseChannel", func(t *testing.T) {
		// SETUP
		ctx, _, channelKeeper, addrs, _ := createMockApp()

		sender := addrs[0]
		receiver := addrs[1]
		coins := sdk.Coins{sdk.NewCoin("KVA", 10)}

		// create new channel
		channelID := ChannelID(0) // should be 0 as first channel
		channel := Channel{
			ID:           channelID,
			Participants: [2]sdk.AccAddress{sender, receiver},
			Coins:        coins,
		}
		channelKeeper.setChannel(ctx, channel)

		// create closing update
		payouts := Payouts{
			{sender, sdk.Coins{sdk.NewCoin("KVA", 3)}},
			{receiver, sdk.Coins{sdk.NewCoin("KVA", 7)}},
		}
		update := Update{
			ChannelID: channelID,
			Payouts:   payouts,
			Sigs:      [1]crypto.Signature{},
		}
		// Set empty submittedUpdatesQueue TODO work out proper genesis initialisation
		channelKeeper.setSubmittedUpdatesQueue(ctx, SubmittedUpdatesQueue{})

		// ACTION
		_, err := channelKeeper.InitCloseChannelBySender(ctx, update)

		// CHECK RESULTS
		// no error
		assert.Nil(t, err)
		// submittedupdate in queue and correct
		suq, found := channelKeeper.getSubmittedUpdatesQueue(ctx)
		assert.True(t, found)
		assert.True(t, suq.Contains(channelID))

		su, found := channelKeeper.getSubmittedUpdate(ctx, channelID)
		assert.True(t, found)
		expectedSubmittedUpdate := SubmittedUpdate{
			Update:        update,
			ExecutionTime: ChannelDisputeTime,
		}
		assert.Equal(t, expectedSubmittedUpdate, su)
		// TODO check channel is still in db and coins haven't changed?
	})

}

/*

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
*/
