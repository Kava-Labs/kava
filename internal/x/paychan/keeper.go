package paychan

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
)

// keeper of the paychan store
type Keeper struct {
	storeKey sdk.StoreKey
	cdc        *wire.Codec // needed to serialize objects before putting them in the store
	coinKeeper bank.Keeper

	// codespace
	//codespace sdk.CodespaceType // ??
}

// Called when creating new app.
//func NewKeeper(cdc *wire.Codec, key sdk.StoreKey, ck bank.Keeper, codespace sdk.CodespaceType) Keeper {
func NewKeeper(cdc *wire.Codec, key sdk.StoreKey, ck bank.Keeper) Keeper {
	keeper := Keeper{
		storeKey: key,
		cdc:        cdc,
		coinKeeper: ck,
		//codespace:  codespace,
	}
	return keeper
}

// bunch of business logic ...


// Reteive a payment channel struct from the blockchain store.
// They are indexed by a concatenation of sender address, receiver address, and an integer.
func (keeper Keeper) GetPaychan(ctx sdk.Context, sender sdk.Address, receiver sdk.Address, id integer) (Paychan, bool) {
	// Return error as second argument instead of bool?
	var pych Paychan
	// load from DB
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(paychanKey(sender, receiver, id))
	if bz == nil {
		return pych, false
	}
	// unmarshal
	k.cdc.MustUnmarshalBinary(bz, &pych)
	// return
	return pych, true
}

// Store payment channel struct in blockchain store.
func (keeper Keeper) setPaychan(pych Paychan) sdk.Error {
	store := ctx.KVStore(k.storeKey)
	// marshal
	bz := k.cdc.MustMarshalBinary(pych)
	// write to db
	pychKey := paychanKey(pych.sender, pych.receiver, pych.id)
	store.Set(pychKey, bz)
	// TODO handler errors
}

// Create a new payment channel and lock up sender funds.
func (keeer Keeper) CreatePaychan(sender sdk.Address, receiver sdkAddress, amt sdk.Coins) (sdk.Tags, sdk.Error) {
	// Calculate next id (num existing paychans plus 1)
	id := len(keeper.GetPaychans(sender, receiver)) + 1
	// subtract coins from sender
	k.coinKeeper.SubtractCoins(ctx, sender, amt)
	// create new Paychan struct (create ID)
	pych := Paychan{sender,
		receiver,
		id,
		balance: amt}
	// save to db
	err := k.setPaychan(pych)


	// TODO validation
	// sender has enough coins - done in Subtract method
	// receiver address exists?
	// paychan doesn't exist already

	tags := sdk.NewTags()
	return tags, err
}

// Close a payment channel and distribute funds to participants.
func (keeper Keeper) ClosePaychan(sender sdk.Address, receiver sdk.Address, id integer, receiverAmt sdk.Coins) (sdk.Tags, sdk.Error) {
	pych := GetPaychan(ctx, sender, receiver, id)
	// compute coin distribution
	senderAmt = pych.balance.Minus(receiverAmt) // Minus sdk.Coins method
	// add coins to sender
	k.coinKeeper.AddCoins(ctx, sender, senderAmt)
	// add coins to receiver
	k.coinKeeper.AddCoins(ctx, receiver, receiverAmt)
	// delete paychan from db
	pychKey := paychanKey(pych.sender, pych.receiver, pych.id)
	store.Delete(pychKey)


	// TODO validation
	// paychan exists
	// output coins are less than paychan balance
	// sender and receiver addresses exist?

	//sdk.NewTags(
	//	"action", []byte("channel closure"),
	//	"receiver", receiver.Bytes(),
	//	"sender", sender.Bytes(),
	//	"id", ??)
	tags := sdk.NewTags()
	return tags, nil
}

// Creates a key to reference a paychan in the blockchain store.
func paychanKey(sender sdk.Address, receiver sdk.Address, id integer) []byte {
	
	//sdk.Address is just a slice of bytes under a different name
	//convert id to string then to byte slice
	idAsBytes := []byte(strconv.Itoa(id))
	// concat sender and receiver and integer ID
	return append(sender.Bytes(), receiver.Bytes()..., idAsBytes...)
}

// Get all paychans between a given sender and receiver.
func (keeper Keeper) GetPaychans(sender sdk.Address, receiver sdk.Address) []Paychan {
	var paychans []Paychan
	// TODO Implement this
	return paychans
}
// maybe getAllPaychans(sender sdk.address) []Paychan
