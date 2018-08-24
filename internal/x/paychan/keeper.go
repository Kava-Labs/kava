package paychan

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/bank"
)

// Keeper of the paychan store
// Handles validation internally. Does not rely on calling code to do validation.
// Aim to keep public methods safe, private ones not necessaily.
type Keeper struct {
	storeKey   sdk.StoreKey
	cdc        *wire.Codec // needed to serialize objects before putting them in the store
	coinKeeper bank.Keeper

	// TODO investigate codespace
	//codespace sdk.CodespaceType
}

// Called when creating new app.
func NewKeeper(cdc *wire.Codec, key sdk.StoreKey, ck bank.Keeper) Keeper {
	keeper := Keeper{
		storeKey:   key,
		cdc:        cdc,
		coinKeeper: ck,
		//codespace:  codespace,
	}
	return keeper
}

// bunch of business logic ...
/*
// Reteive a payment channel struct from the blockchain store.
// They are indexed by a concatenation of sender address, receiver address, and an integer.
func (k Keeper) GetPaychan(ctx sdk.Context, sender sdk.Address, receiver sdk.Address, id int64) (Paychan, bool) {
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
func (k Keeper) setPaychan(ctx sdk.Context, pych Paychan) {
	store := ctx.KVStore(k.storeKey)
	// marshal
	bz := k.cdc.MustMarshalBinary(pych) // panics if something goes wrong
	// write to db
	pychKey := paychanKey(pych.Sender, pych.Receiver, pych.Id)
	store.Set(pychKey, bz) // panics if something goes wrong
}
*/

// Create a new payment channel and lock up sender funds.
func (k Keeper) CreatePaychan(ctx sdk.Context, sender sdk.Address, receiver sdk.Address, coins sdk.Coins) (sdk.Tags, sdk.Error) {
	// TODO do validation and maybe move somewhere nicer
	/*
		// args present
		if len(sender) == 0 {
			return nil, sdk.ErrInvalidAddress(sender.String())
		}
		if len(receiver) == 0 {
			return nil, sdk.ErrInvalidAddress(receiver.String())
		}
		if len(amount) == 0 {
			return nil, sdk.ErrInvalidCoins(amount.String())
		}
		// Check if coins are sorted, non zero, positive
		if !amount.IsValid() {
			return nil, sdk.ErrInvalidCoins(amount.String())
		}
		if !amount.IsPositive() {
			return nil, sdk.ErrInvalidCoins(amount.String())
		}
		// sender should exist already as they had to sign.
		// receiver address exists. am is the account mapper in the coin keeper.
		// TODO automatically create account if not present?
		// TODO remove as account mapper not available to this pkg
		//if k.coinKeeper.am.GetAccount(ctx, receiver) == nil {
		//	return nil, sdk.ErrUnknownAddress(receiver.String())
		//}

		// sender has enough coins - done in Subtract method
		// TODO check if sender and receiver different?
	*/

	// Calculate next id
	id := k.getNewChannelID(ctx)
	// subtract coins from sender
	_, tags, err := k.coinKeeper.SubtractCoins(ctx, sender, coins)
	if err != nil {
		return nil, err
	}
	// create new Paychan struct
	channel := Channel{
		ID: id
		Participants:	[2]sdk.AccAddress{sender, receiver},
		Coins:  coins,
	}
	// save to db
	k.setChannel(ctx, channel)

	// TODO create tags
	//tags := sdk.NewTags()
	return tags, err
}

/* This is how gov manages creating unique IDs. Needs to be deterministic - can't use UUID
func (keeper Keeper) getNewChannelID(ctx sdk.Context) (channelID int64, err sdk.Error) {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(KeyNextProposalID)
	if bz == nil {
		return -1, ErrInvalidGenesis(keeper.codespace, "InitialProposalID never set")
	}
	keeper.cdc.MustUnmarshalBinary(bz, &proposalID)
	bz = keeper.cdc.MustMarshalBinary(proposalID + 1)
	store.Set(KeyNextProposalID, bz)
	return proposalID, nil
*/

func (k Keeper) ChannelCloseByReceiver() () {
	// Validate inputs
	// k.closeChannel
}

func (k Keeper) InitChannelCloseBySender() () {
	// Validate inputs
	// Create SubmittedUpdate from Update and add to queue
}

func (k Keeper) closeChannel() () {
	// Remove corresponding SubmittedUpdate from queue (if it exist)
	// Add coins to sender and receiver
	// Delete Channel
}
/*
// Close a payment channel and distribute funds to participants.
func (k Keeper) ClosePaychan(ctx sdk.Context, sender sdk.Address, receiver sdk.Address, id int64, receiverAmount sdk.Coins) (sdk.Tags, sdk.Error) {
	if len(sender) == 0 {
		return nil, sdk.ErrInvalidAddress(sender.String())
	}
	if len(receiver) == 0 {
		return nil, sdk.ErrInvalidAddress(receiver.String())
	}
	if len(receiverAmount) == 0 {
		return nil, sdk.ErrInvalidCoins(receiverAmount.String())
	}
	// check id ≥ 0
	if id < 0 {
		return nil, sdk.ErrInvalidAddress(strconv.Itoa(int(id))) // TODO implement custom errors
	}

	// Check if coins are sorted, non zero, non negative
	if !receiverAmount.IsValid() {
		return nil, sdk.ErrInvalidCoins(receiverAmount.String())
	}
	if !receiverAmount.IsPositive() {
		return nil, sdk.ErrInvalidCoins(receiverAmount.String())
	}

	store := ctx.KVStore(k.storeKey)

	pych, exists := k.GetPaychan(ctx, sender, receiver, id)
	if !exists {
		return nil, sdk.ErrUnknownAddress("paychan not found") // TODO implement custom errors
	}
	// compute coin distribution
	senderAmount := pych.Balance.Minus(receiverAmount) // Minus sdk.Coins method
	// check that receiverAmt not greater than paychan balance
	if !senderAmount.IsNotNegative() {
		return nil, sdk.ErrInsufficientFunds(pych.Balance.String())
	}
	// add coins to sender
	// creating account if it doesn't exist
	k.coinKeeper.AddCoins(ctx, sender, senderAmount)
	// add coins to receiver
	k.coinKeeper.AddCoins(ctx, receiver, receiverAmount)

	// delete paychan from db
	pychKey := paychanKey(pych.Sender, pych.Receiver, pych.Id)
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
func paychanKey(sender sdk.Address, receiver sdk.Address, id int64) []byte {

	//sdk.Address is just a slice of bytes under a different name
	//convert id to string then to byte slice
	idAsBytes := []byte(strconv.Itoa(int(id)))
	// concat sender and receiver and integer ID
	key := append(sender.Bytes(), receiver.Bytes()...)
	key = append(key, idAsBytes...)
	return key
}

// Get all paychans between a given sender and receiver.
func (k Keeper) GetPaychans(sender sdk.Address, receiver sdk.Address) []Paychan {
	var paychans []Paychan
	// TODO Implement this
	return paychans
}

// maybe getAllPaychans(sender sdk.address) []Paychan
*/
