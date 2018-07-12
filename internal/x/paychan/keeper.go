package paychan

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
)

// keeper of the paychan store
// Handles validation internally. Does not rely on calling code to do validation.
// Aim to keep public methids safe, private ones not necessaily.
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
	store.Set(pychKey, bz) // panics if something goes wrong
}

// Create a new payment channel and lock up sender funds.
func (keeer Keeper) CreatePaychan(ctx sdk.Context, sender sdk.Address, receiver sdkAddress, amt sdk.Coins) (sdk.Tags, sdk.Error) {
	// TODO move validation somewhere nicer
	// args present
	if len(sender) == 0 {
		return sdk.ErrInvalidAddress(sender.String())
	}
	if len(receiver) == 0 {
		return sdk.ErrInvalidAddress(receiver.String())
	}
	if len(amount) == 0 {
		return sdk.ErrInvalidCoins(amount.String())
	}
	// Check if coins are sorted, non zero, positive
	if !amount.IsValid() {
		return sdk.ErrInvalidCoins(amount.String())
	}
	if !amount.IsPositive() {
		return sdk.ErrInvalidCoins(amount.String())
	}
	// sender should exist already as they had to sign.
	// receiver address exists. am is the account mapper in the coin keeper.
	// TODO automatically create account if not present?
	if k.coinKepper.am.GetAccount(ctx, receiver) == nil {
		return sdk.ErrUnknownAddress(receiver.String())
	}
	// sender has enough coins - done in Subtract method
	// TODO check if sender and receiver different?
	

	// Calculate next id (num existing paychans plus 1)
	id := len(keeper.GetPaychans(sender, receiver)) + 1 // TODO check for overflow?
	// subtract coins from sender
	coins, tags, err := k.coinKeeper.SubtractCoins(ctx, sender, amt)
	if err != nil {
		return nil, err
	}
	// create new Paychan struct
	pych := Paychan{sender,
					receiver,
					id,
					balance: amt}
	// save to db
	k.setPaychan(pych)
	

	// TODO create tags
	tags := sdk.NewTags()
	return tags, err
}

// Close a payment channel and distribute funds to participants.
func (keeper Keeper) ClosePaychan(sender sdk.Address, receiver sdk.Address, id integer, receiverAmt sdk.Coins) (sdk.Tags, sdk.Error) {
	if len(msg.sender) == 0 {
		return sdk.ErrInvalidAddress(msg.sender.String())
	}
	if len(msg.receiver) == 0 {
		return sdk.ErrInvalidAddress(msg.receiver.String())
	}
	if len(msg.receiverAmount) == 0 {
		return sdk.ErrInvalidCoins(msg.receiverAmount.String())
	}
	// check id ≥ 0
	if msg.id < 0 {
		return sdk.ErrInvalidAddress(strconv.Itoa(id)) // TODO implement custom errors
	}

	// Check if coins are sorted, non zero, non negative
	if !msg.receiverAmount.IsValid() {
		return sdk.ErrInvalidCoins(msg.receiverAmount.String())
	}
	if !msg.receiverAmount.IsPositive() {
		return sdk.ErrInvalidCoins(msg.receiverAmount.String())
	}


	store := ctx.KVStore(k.storeKey)

	pych, exists := GetPaychan(ctx, sender, receiver, id)
	if !exists {
		return nil, sdk.ErrUnknownAddress() // TODO implement custom errors
	}
	// compute coin distribution
	senderAmt = pych.balance.Minus(receiverAmt) // Minus sdk.Coins method
	// check that receiverAmt not greater than paychan balance
	if !senderAmt.IsNotNegative() {
		return nil, sdk.ErrInsufficientFunds(pych.balance.String())
	}
	// add coins to sender
	// creating account if it doesn't exist
	k.coinKeeper.AddCoins(ctx, sender, senderAmt)
	// add coins to receiver
	k.coinKeeper.AddCoins(ctx, receiver, receiverAmt)

	// delete paychan from db
	pychKey := paychanKey(pych.sender, pych.receiver, pych.id)
	store.Delete(pychKey)

	// TODO create tags
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
