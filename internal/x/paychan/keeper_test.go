package paychan

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	//"github.com/stretchr/testify/require"
	//"github.com/tendermint/tendermint/crypto"
	"testing"
)

func TestKeeper(t *testing.T) {
	t.Run("CreateChannel", func(t *testing.T) {

		// TODO test for receiver account not existing (OK) and sender not existing (not ok)

		accountSeeds := []string{"senderSeed", "receiverSeed"}
		const (
			senderAccountIndex   int = 0
			receiverAccountIndex int = 1
		)
		_, addrs, _, _ := createTestGenAccounts(accountSeeds, sdk.Coins{}) // pure function

		testCases := []struct {
			name                string
			sender              sdk.AccAddress
			receiver            sdk.AccAddress
			coins               sdk.Coins
			shouldCreateChannel bool
			shouldError         bool
		}{
			{
				"HappyPath",
				addrs[senderAccountIndex],
				addrs[receiverAccountIndex],
				sdk.Coins{sdk.NewCoin("KVA", 10)},
				true,
				false,
			},
			{
				"NilAddress",
				sdk.AccAddress{},
				sdk.AccAddress{},
				sdk.Coins{sdk.NewCoin("KVA", 10)},
				false,
				true,
			},
			{
				"NilCoins",
				addrs[senderAccountIndex],
				addrs[receiverAccountIndex],
				sdk.Coins{},
				false,
				true,
			},
			{
				"NegativeCoins",
				addrs[senderAccountIndex],
				addrs[receiverAccountIndex],
				sdk.Coins{sdk.NewCoin("KVA", -57)},
				false,
				true,
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				////// SETUP
				// create basic mock app
				ctx, coinKeeper, channelKeeper, addrs, _, _, genAccFunding := createMockApp(accountSeeds)
				//
				////// ACTION
				_, err := channelKeeper.CreateChannel(ctx, testCase.sender, testCase.receiver, testCase.coins)

				//
				////// CHECK RESULTS
				// Check error
				if testCase.shouldError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
				// Check if channel exists and is correct
				channelID := ChannelID(0) // should be 0 as first channel
				createdChan, found := channelKeeper.getChannel(ctx, channelID)

				if testCase.shouldCreateChannel {
					expectedChan := Channel{
						ID:           channelID,
						Participants: [2]sdk.AccAddress{testCase.sender, testCase.receiver},
						Coins:        testCase.coins,
					}

					// channel exists and correct
					assert.True(t, found)
					assert.Equal(t, expectedChan, createdChan)
					// check coins deducted from sender
					assert.Equal(t, genAccFunding.Minus(testCase.coins), coinKeeper.GetCoins(ctx, testCase.sender))
					// check no coins deducted from receiver
					assert.Equal(t, genAccFunding, coinKeeper.GetCoins(ctx, testCase.receiver))
					// check next global channelID incremented
					assert.Equal(t, ChannelID(1), channelKeeper.getNewChannelID(ctx))
				} else {
					// channel doesn't exist
					assert.False(t, found)
					assert.Equal(t, Channel{}, createdChan)
					// check no coins deducted from sender
					assert.Equal(t, genAccFunding, coinKeeper.GetCoins(ctx, addrs[senderAccountIndex]))
					// check no coins deducted from receiver
					assert.Equal(t, genAccFunding, coinKeeper.GetCoins(ctx, addrs[receiverAccountIndex]))
					// check next global channelID not incremented
					assert.Equal(t, ChannelID(0), channelKeeper.getNewChannelID(ctx))
				}
			})
		}
	})

	t.Run("CloseChannelByReceiver", func(t *testing.T) {
		// TODO convert to table driven and add more test cases
		//		channel exists or not (assume channels correct)
		//		various Updates
		//		submittedUpdates existing or not (assume they are valid)

		// SETUP
		accountSeeds := []string{"senderSeed", "receiverSeed"}
		const (
			senderAccountIndex   int = 0
			receiverAccountIndex int = 1
		)
		ctx, coinKeeper, channelKeeper, addrs, pubKeys, privKeys, genAccFunding := createMockApp(accountSeeds)

		coins := sdk.Coins{sdk.NewCoin("KVA", 10)}

		// create new channel
		channelID := ChannelID(0) // should be 0 as first channel
		channel := Channel{
			ID:           channelID,
			Participants: [2]sdk.AccAddress{addrs[senderAccountIndex], addrs[receiverAccountIndex]},
			Coins:        coins,
		}
		channelKeeper.setChannel(ctx, channel)

		// create closing update
		payout := Payout{sdk.Coins{sdk.NewCoin("KVA", 3)}, sdk.Coins{sdk.NewCoin("KVA", 7)}}
		update := Update{
			ChannelID: channelID,
			Payout:    payout,
			// empty sig
		}
		cryptoSig, _ := privKeys[senderAccountIndex].Sign(update.GetSignBytes())
		update.Sigs = [1]UpdateSignature{UpdateSignature{
			PubKey:          pubKeys[senderAccountIndex],
			CryptoSignature: cryptoSig,
		}}

		// Set empty submittedUpdatesQueue TODO work out proper genesis initialisation
		channelKeeper.setSubmittedUpdatesQueue(ctx, SubmittedUpdatesQueue{})

		// ACTION
		_, err := channelKeeper.CloseChannelByReceiver(ctx, update)

		// CHECK RESULTS
		// no error
		assert.NoError(t, err)
		// coins paid out
		senderPayout := payout[senderAccountIndex]
		assert.Equal(t, genAccFunding.Plus(senderPayout), coinKeeper.GetCoins(ctx, addrs[senderAccountIndex]))
		receiverPayout := payout[receiverAccountIndex]
		assert.Equal(t, genAccFunding.Plus(receiverPayout), coinKeeper.GetCoins(ctx, addrs[receiverAccountIndex]))
		// channel deleted
		_, found := channelKeeper.getChannel(ctx, channelID)
		assert.False(t, found)

	})

	t.Run("InitCloseChannelBySender", func(t *testing.T) {

		// TODO do some documentation here
		// Ideally this should mock calls to ctx.store.Get/Set - test the side effects without being dependent on implementatino details
		// TODO test correct behaviour when a submittedUpdate already exists

		accountSeeds := []string{"senderSeed", "receiverSeed", "notInChannelSeed"}
		const (
			senderAccountIndex   int = 0
			receiverAccountIndex int = 1
			otherAccountIndex    int = 2
		)
		chanID := ChannelID(0)

		type testUpdate struct { // A parameterised version of an Update for use in specifying test cases.
			channelID          ChannelID // channelID of submitted update
			payout             Payout    // payout of submitted update
			pubKeyAccountIndex int       // pubkey of signature of submitted update
			sigAccountIndex    int       // crypto signature of signature of submitted update
		}
		testCases := []struct {
			name                    string
			setupChannel            bool
			updateToSubmit          testUpdate
			expectedSubmittedUpdate string // "empty" or "sameAsSubmitted"
			shouldError             bool
		}{
			{
				"HappyPath",
				true,
				testUpdate{chanID, Payout{sdk.Coins{sdk.NewCoin("KVA", 3)}, sdk.Coins{sdk.NewCoin("KVA", 7)}}, senderAccountIndex, senderAccountIndex},
				"sameAsSubmited",
				false,
			},
			{
				"NoChannel",
				false,
				testUpdate{chanID, Payout{sdk.Coins{sdk.NewCoin("KVA", 3)}, sdk.Coins{sdk.NewCoin("KVA", 7)}}, senderAccountIndex, senderAccountIndex},
				"empty",
				true,
			},
			{
				"NoCoins",
				true,
				testUpdate{chanID, Payout{sdk.Coins{}}, senderAccountIndex, senderAccountIndex},
				"empty",
				true,
			},
			{
				"NegativeCoins",
				true,
				testUpdate{chanID, Payout{sdk.Coins{sdk.NewCoin("KVA", -5)}, sdk.Coins{sdk.NewCoin("KVA", 15)}}, senderAccountIndex, senderAccountIndex},
				"empty",
				true,
			},
			{
				"TooManyCoins",
				true,
				testUpdate{chanID, Payout{sdk.Coins{sdk.NewCoin("KVA", 100)}, sdk.Coins{sdk.NewCoin("KVA", 7)}}, senderAccountIndex, senderAccountIndex},
				"empty",
				true,
			},
			{
				"WrongSignature",
				true,
				testUpdate{chanID, Payout{sdk.Coins{sdk.NewCoin("KVA", 3)}, sdk.Coins{sdk.NewCoin("KVA", 7)}}, senderAccountIndex, otherAccountIndex},
				"empty",
				true,
			},
			{
				"WrongPubKey",
				true,
				testUpdate{chanID, Payout{sdk.Coins{sdk.NewCoin("KVA", 3)}, sdk.Coins{sdk.NewCoin("KVA", 7)}}, otherAccountIndex, senderAccountIndex},
				"empty",
				true,
			},
			{
				"ReceiverSigned",
				true,
				testUpdate{chanID, Payout{sdk.Coins{sdk.NewCoin("KVA", 3)}, sdk.Coins{sdk.NewCoin("KVA", 7)}}, receiverAccountIndex, receiverAccountIndex},
				"empty",
				true,
			},
		}
		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {

				// SETUP
				ctx, _, channelKeeper, addrs, pubKeys, privKeys, _ := createMockApp(accountSeeds)
				// Set empty submittedUpdatesQueue TODO work out proper genesis initialisation
				channelKeeper.setSubmittedUpdatesQueue(ctx, SubmittedUpdatesQueue{})
				// create new channel
				if testCase.setupChannel {
					channel := Channel{
						ID:           chanID, // should be 0 as first channel
						Participants: [2]sdk.AccAddress{addrs[senderAccountIndex], addrs[receiverAccountIndex]},
						Coins:        sdk.Coins{sdk.NewCoin("KVA", 10)},
					}
					channelKeeper.setChannel(ctx, channel)
				}

				// create update
				// basic values
				updateToSubmit := Update{
					ChannelID: testCase.updateToSubmit.channelID,
					Payout:    testCase.updateToSubmit.payout,
					// empty sig
				}
				// create update's signature
				cryptoSig, _ := privKeys[testCase.updateToSubmit.sigAccountIndex].Sign(updateToSubmit.GetSignBytes())
				updateToSubmit.Sigs = [1]UpdateSignature{UpdateSignature{
					PubKey:          pubKeys[testCase.updateToSubmit.pubKeyAccountIndex],
					CryptoSignature: cryptoSig,
				}}

				// ACTION
				_, err := channelKeeper.InitCloseChannelBySender(ctx, updateToSubmit)

				// CHECK RESULTS
				// Check error
				if testCase.shouldError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
				// Check submittedUpdate
				su, found := channelKeeper.getSubmittedUpdate(ctx, chanID)
				switch testCase.expectedSubmittedUpdate {
				case "empty":
					assert.False(t, found)
					assert.Zero(t, su)
				case "sameAsSubmitted":
					assert.True(t, found)
					expectedSU := SubmittedUpdate{updateToSubmit, ChannelDisputeTime}
					assert.Equal(t, expectedSU, su)
				}

			})
		}

	})

}

/*
func privAndAddr() (crypto.PrivKey, sdk.AccAddress) {
	priv := ed25519.GenPrivKey()
	addr := sdk.AccAddress(priv.PubKey().Address())
	return priv, addr

sig, err := priv.Sign(signBytes)
*/

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
