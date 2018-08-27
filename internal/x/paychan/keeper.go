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

// ============================================== Main Business Logic


// Create a new payment channel and lock up sender funds.
func (k Keeper) CreateChannel(ctx sdk.Context, sender sdk.Address, receiver sdk.Address, coins sdk.Coins) (sdk.Tags, sdk.Error) {
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

	// TODO add to tags

	return tags, err
}



func (k Keeper) InitCloseChannelBySender(update Update) {
	// This is roughly the default path for non unidirectional channels

	// TODO Validate update - e.g. check signed by sender

	q := k.getSubmittedUpdateQueue(ctx)
	if q.Contains(update.ChannelID) {
		// Someone has previously tried to update channel
		existingSUpdate := k.getSubmittedUpdate(ctx, update.ChannelID)
		k.addToSubmittedUpdateQueue(ctx, k.applyNewUpdate(existingSUpdate, update))
	} else {
		// No one has tried to update channel.
		submittedUpdate := SubmittedUpdate{
			Update: update
			executionTime: ctx.BlockHeight()+ChannelDisputeTime //TODO check what exactly BlockHeight refers to
		}
		k.addToSubmittedUpdateQueue(ctx, submittedUpdate)
	}
}

func (k Keeper) CloseChannelByReceiver(update Update) () {
	// TODO Validate update

	// Check if there is an update in the queue already
	q := k.getSubmittedUpdateQueue(ctx)
	if q.Contains(update.ChannelID) {
		// Someone has previously tried to update channel but receiver has final say
		k.removeFromSubmittedUpdateQueue(ctx, update.ChannelID)
	}
	
	k.closeChannel(ctx, update)
}

// Main function that compare updates against each other.
// Pure function
func (k Keeper) applyNewUpdate(existingSUpdate, proposedUpdate) SubmittedUpdate {
	var returnUpdate SubmittedUpdate

	if existingSUpdate.sequence > proposedUpdate.sequence {
		// update accepted
		returnUpdate = SubmittedUpdate{
			Update: proposedUpdate
			ExecutionTime: existingSUpdate.ExecutionTime
		}
	} else {
		// update rejected
		returnUpdate = existingSUpdate
	}
	return returnUpdate
}

func (k Keeper) closeChannel(ctx sdk.Context, update Update) {
	channel := k.getChannel(ctx, update.ChannelID)

	// Add coins to sender and receiver
	for address, coins := range update.CoinsUpdate {
		// TODO check somewhere if coins are not negative?
		k.ck.AddCoins(ctx, address, coins)
	}
	
	k.deleteChannel(ctx, update.ChannelID)
}



// =========================================== QUEUE


func (k Keeper) addToSubmittedUpdatesQueue(ctx sdk.Context, sUpdate SubmittedUpdate) {
	// always overwrite prexisting values - leave paychan logic to higher levels
	// get current queue
	q := k.getSubmittedUpdateQueue(ctx)
	// append ID to queue
	if q.Contains(sUpdate.ChannelID)! {
		q = append(q, sUpdate.ChannelID)
	}
	// set queue
	k.setSubmittedUpdateQueue(ctx, q)
	// store submittedUpdate
	k.setSubmittedUpdate(ctx, sUpdate)
}
func (k Keeper) removeFromSubmittdUpdatesQueue(ctx sdk.Context, channelID) {
	// get current queue
	q := k.getSubmittedUpdateQueue(ctx)
	// remove id
	q.RemoveMatchingElements(channelID)
	// set queue
	k.setSubmittedUpdateQueue(ctx, q)
	// delete submittedUpdate
	k.deleteSubmittedUpdate(ctx, channelID)
}

func (k Keeper) getSubmittedUpdatesQueue(ctx sdk.Context) (Queue, bool) {
	// load from DB
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(k.getSubmittedUpdatesQueueKey())

	var q Queue
	if bz == nil {
		return q, false
	}
	// unmarshal
	k.cdc.MustUnmarshalBinary(bz, &q)
	// return
	return q, true
}
func (k Keeper) setSubmittedUpdatesQueue(ctx sdk.Context, q Queue) {
	store := ctx.KVStore(k.storeKey)
	// marshal
	bz := k.cdc.MustMarshalBinary(q)
	// write to db
	key := k.getSubmittedUpdatesQueueKey()
	store.Set(key, bz)
}
func (k Keeper) getSubmittedUpdatesQueueKey() []byte {
	return []byte("submittedUpdatesQueue")
}

// ============= SUBMITTED UPDATES
// These are keyed by the IDs of thei associated Channels
// This section deals with only setting and getting

func (k Keeper) getSubmittedUpdate(ctx sdk.Context, channelID ChannelID) (SubmittedUpdate, bool) {
	
	// load from DB
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(k.getSubmittedUpdateKey(channelID))

	var sUpdate SubmittedUpdate
	if bz == nil {
		return sUpdate, false
	}
	// unmarshal
	k.cdc.MustUnmarshalBinary(bz, &sUpdate)
	// return
	return sUpdate, true
}

// Store payment channel struct in blockchain store.
func (k Keeper) setSubmittedUpdate(ctx sdk.Context, sUpdate SubmittedUpdate) {
	store := ctx.KVStore(k.storeKey)
	// marshal
	bz := k.cdc.MustMarshalBinary(sUpdate) // panics if something goes wrong
	// write to db
	key := k.getSubmittedUpdateKey(sUpdate.channelID)
	store.Set(key, bz) // panics if something goes wrong
}

func (k Keeper) deleteSubmittedUpdate(ctx sdk.Context, channelID ) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(k.getSubmittedUpdateKey(channelID))
	// TODO does this have return values? What happens when key doesn't exist?
}
func (k Keeper) getSubmittedUpdateKey(channelID ChannelID) []byte {
	return []byte(fmt.Sprintf("submittedUpdate:%d", channelID))
}


// ========================================== CHANNELS

// Reteive a payment channel struct from the blockchain store.
func (k Keeper) getChannel(ctx sdk.Context, channelID ChannelID) (Channel, bool) {
	
	// load from DB
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(k.getChannelKey(channelID))

	var channel Channel
	if bz == nil {
		return channel, false
	}
	// unmarshal
	k.cdc.MustUnmarshalBinary(bz, &channel)
	// return
	return channel, true
}

// Store payment channel struct in blockchain store.
func (k Keeper) setChannel(ctx sdk.Context, channel Channel) {
	store := ctx.KVStore(k.storeKey)
	// marshal
	bz := k.cdc.MustMarshalBinary(channel) // panics if something goes wrong
	// write to db
	key := sdk.getChannelKey(channel.ID)
	store.Set(key, bz) // panics if something goes wrong
}

func (k Keeper) deleteChannel(ctx sdk.Context, channelID ) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(k.getChannelKey(channelID))
	// TODO does this have return values? What happens when key doesn't exist?
}

func (k Keeper) getNewChannelID(ctx sdk.Context) (int64, error) {
	// get last channel ID
	store := k.KVStore(k.storeKey)
	bz := store.Get(k.getLastChannelIDKey())
	if bz == nil {
		return nil, // TODO throw some error (assumes this has been initialized elsewhere) or just set to zero here
	}
	var lastID ChannelID
	k.cdc.MustUnmarshalBinary(bz, &lastID)
	// increment to create new one
	newID := lastID+1
	bz = k.cdc.MustMarshalBinary(newID)
	// set last channel id again
	store.Set(k.getLastChannelIDKey(), bz)
	// return
	return newID
}

func (k Keeper) getChannelKey(channelID ChannelID) []byte {
	return []bytes(fmt.Sprintf("channel:%d", channelID))
}
func (k Keeper) getLastChannelIDKey() []byte {
	return []bytes("lastChannelID")
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
