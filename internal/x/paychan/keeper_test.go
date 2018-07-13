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
		sender:   sdk.Address([]byte("senderAddress")),
		receiver: sdk.Address([]byte("receiverAddress")),
		id:       0,
		balance:  sdk.Coins{{"KVA", 100}},
	}
	paychanKeeper.setPaychan(ctx, p)

	_, exists = paychanKeeper.GetPaychan(ctx, p.sender, p.receiver, p.id)
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

	p, _ = paychanKeeper.GetPaychan(ctx, senderAddress, receiverAddress, 0)
	if !p.balance.IsEqual(balance) {
		t.Error("payment channel balance incorrect", p.balance, balance)
	}
	expectedNewSenderFunds := senderFunds.Minus(balance)
	if !coinKeeper.GetCoins(ctx, senderAddress).IsEqual(expectedNewSenderFunds) {
		t.Error("sender has incorrect balance after paychan creation")
	}

}

// example from x/bank

//func TestKeeper(t *testing.T) {
// ms, authKey := setupMultiStore()

// cdc := wire.NewCodec()
// auth.RegisterBaseAccount(cdc)

// ctx := sdk.NewContext(ms, abci.Header{}, false, nil, log.NewNopLogger())
// accountMapper := auth.NewAccountMapper(cdc, authKey, &auth.BaseAccount{})
// coinKeeper := NewKeeper(accountMapper)

// addr := sdk.Address([]byte("addr1"))
// addr2 := sdk.Address([]byte("addr2"))
// addr3 := sdk.Address([]byte("addr3"))
// acc := accountMapper.NewAccountWithAddress(ctx, addr)

// // Test GetCoins/SetCoins
// accountMapper.SetAccount(ctx, acc)
// assert.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{}))

// coinKeeper.SetCoins(ctx, addr, sdk.Coins{{"foocoin", 10}})
// assert.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{{"foocoin", 10}}))

// // Test HasCoins
// assert.True(t, coinKeeper.HasCoins(ctx, addr, sdk.Coins{{"foocoin", 10}}))
// assert.True(t, coinKeeper.HasCoins(ctx, addr, sdk.Coins{{"foocoin", 5}}))
// assert.False(t, coinKeeper.HasCoins(ctx, addr, sdk.Coins{{"foocoin", 15}}))
// assert.False(t, coinKeeper.HasCoins(ctx, addr, sdk.Coins{{"barcoin", 5}}))

// // Test AddCoins
// coinKeeper.AddCoins(ctx, addr, sdk.Coins{{"foocoin", 15}})
// assert.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{{"foocoin", 25}}))

// coinKeeper.AddCoins(ctx, addr, sdk.Coins{{"barcoin", 15}})
// assert.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{{"barcoin", 15}, {"foocoin", 25}}))

// // Test SubtractCoins
// coinKeeper.SubtractCoins(ctx, addr, sdk.Coins{{"foocoin", 10}})
// coinKeeper.SubtractCoins(ctx, addr, sdk.Coins{{"barcoin", 5}})
// assert.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{{"barcoin", 10}, {"foocoin", 15}}))

// coinKeeper.SubtractCoins(ctx, addr, sdk.Coins{{"barcoin", 11}})
// assert.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{{"barcoin", 10}, {"foocoin", 15}}))

// coinKeeper.SubtractCoins(ctx, addr, sdk.Coins{{"barcoin", 10}})
// assert.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{{"foocoin", 15}}))
// assert.False(t, coinKeeper.HasCoins(ctx, addr, sdk.Coins{{"barcoin", 1}}))

// // Test SendCoins
// coinKeeper.SendCoins(ctx, addr, addr2, sdk.Coins{{"foocoin", 5}})
// assert.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{{"foocoin", 10}}))
// assert.True(t, coinKeeper.GetCoins(ctx, addr2).IsEqual(sdk.Coins{{"foocoin", 5}}))

// _, err2 := coinKeeper.SendCoins(ctx, addr, addr2, sdk.Coins{{"foocoin", 50}})
// assert.Implements(t, (*sdk.Error)(nil), err2)
// assert.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{{"foocoin", 10}}))
// assert.True(t, coinKeeper.GetCoins(ctx, addr2).IsEqual(sdk.Coins{{"foocoin", 5}}))

// coinKeeper.AddCoins(ctx, addr, sdk.Coins{{"barcoin", 30}})
// coinKeeper.SendCoins(ctx, addr, addr2, sdk.Coins{{"barcoin", 10}, {"foocoin", 5}})
// assert.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{{"barcoin", 20}, {"foocoin", 5}}))
// assert.True(t, coinKeeper.GetCoins(ctx, addr2).IsEqual(sdk.Coins{{"barcoin", 10}, {"foocoin", 10}}))

// // Test InputOutputCoins
// input1 := NewInput(addr2, sdk.Coins{{"foocoin", 2}})
// output1 := NewOutput(addr, sdk.Coins{{"foocoin", 2}})
// coinKeeper.InputOutputCoins(ctx, []Input{input1}, []Output{output1})
// assert.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{{"barcoin", 20}, {"foocoin", 7}}))
// assert.True(t, coinKeeper.GetCoins(ctx, addr2).IsEqual(sdk.Coins{{"barcoin", 10}, {"foocoin", 8}}))

// inputs := []Input{
// 	NewInput(addr, sdk.Coins{{"foocoin", 3}}),
// 	NewInput(addr2, sdk.Coins{{"barcoin", 3}, {"foocoin", 2}}),
// }

// outputs := []Output{
// 	NewOutput(addr, sdk.Coins{{"barcoin", 1}}),
// 	NewOutput(addr3, sdk.Coins{{"barcoin", 2}, {"foocoin", 5}}),
// }
// coinKeeper.InputOutputCoins(ctx, inputs, outputs)
// assert.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{{"barcoin", 21}, {"foocoin", 4}}))
// assert.True(t, coinKeeper.GetCoins(ctx, addr2).IsEqual(sdk.Coins{{"barcoin", 7}, {"foocoin", 6}}))
// assert.True(t, coinKeeper.GetCoins(ctx, addr3).IsEqual(sdk.Coins{{"barcoin", 2}, {"foocoin", 5}}))

//}
