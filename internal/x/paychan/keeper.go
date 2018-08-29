package paychan

import (
	"fmt"
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
func (k Keeper) CreateChannel(ctx sdk.Context, sender sdk.AccAddress, receiver sdk.AccAddress, coins sdk.Coins) (sdk.Tags, sdk.Error) {
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

	// subtract coins from sender
	_, tags, err := k.coinKeeper.SubtractCoins(ctx, sender, coins)
	if err != nil {
		return nil, err
	}
	// Calculate next id
	id := k.getNewChannelID(ctx)
	// create new Paychan struct
	channel := Channel{
		ID:           id,
		Participants: [2]sdk.AccAddress{sender, receiver},
		Coins:        coins,
	}
	// save to db
	k.setChannel(ctx, channel)

	// TODO add to tags

	return tags, err
}

func (k Keeper) InitCloseChannelBySender(ctx sdk.Context, update Update) (sdk.Tags, sdk.Error) {
	// This is roughly the default path for non unidirectional channels

	// TODO Validate update - e.g. check signed by sender

	q, found := k.getSubmittedUpdatesQueue(ctx)
	if !found {
		panic("SubmittedUpdatesQueue not found.") // TODO nicer custom errors
	}
	if q.Contains(update.ChannelID) {
		// Someone has previously tried to update channel
		existingSUpdate, found := k.getSubmittedUpdate(ctx, update.ChannelID)
		if !found {
			panic("can't find element in queue that should exist")
		}
		k.addToSubmittedUpdatesQueue(ctx, k.applyNewUpdate(existingSUpdate, update))
	} else {
		// No one has tried to update channel
		submittedUpdate := SubmittedUpdate{
			Update:        update,
			ExecutionTime: ctx.BlockHeight() + ChannelDisputeTime, //TODO check what exactly BlockHeight refers to
		}
		k.addToSubmittedUpdatesQueue(ctx, submittedUpdate)
	}

	tags := sdk.EmptyTags() // TODO tags

	return tags, nil
}

func (k Keeper) CloseChannelByReceiver(ctx sdk.Context, update Update) (sdk.Tags, sdk.Error) {
	// TODO Validate update

	// Check if there is an update in the queue already
	q, found := k.getSubmittedUpdatesQueue(ctx)
	if !found {
		panic("SubmittedUpdatesQueue not found.") // TODO nicer custom errors
	}
	if q.Contains(update.ChannelID) {
		// Someone has previously tried to update channel but receiver has final say
		k.removeFromSubmittedUpdatesQueue(ctx, update.ChannelID)
	}

	tags, err := k.closeChannel(ctx, update)

	return tags, err
}

// Main function that compare updates against each other.
// Pure function
func (k Keeper) applyNewUpdate(existingSUpdate SubmittedUpdate, proposedUpdate Update) SubmittedUpdate {
	var returnUpdate SubmittedUpdate

	if existingSUpdate.Sequence > proposedUpdate.Sequence {
		// update accepted
		returnUpdate = SubmittedUpdate{
			Update:        proposedUpdate,
			ExecutionTime: existingSUpdate.ExecutionTime,
		}
	} else {
		// update rejected
		returnUpdate = existingSUpdate
	}
	return returnUpdate
}

// unsafe close channel - doesn't check if update matches existing channel TODO make safer?
func (k Keeper) closeChannel(ctx sdk.Context, update Update) (sdk.Tags, sdk.Error) {
	var err sdk.Error
	var tags sdk.Tags

	// Add coins to sender and receiver
	// TODO check for possible errors first to avoid coins being half paid out?
	for _, payout := range update.Payouts {
		// TODO check somewhere if coins are not negative?
		_, tags, err = k.coinKeeper.AddCoins(ctx, payout.Address, payout.Coins)
		if err != nil {
			panic(err)
		}
	}

	k.deleteChannel(ctx, update.ChannelID)

	return tags, nil
}

// =========================================== QUEUE

func (k Keeper) addToSubmittedUpdatesQueue(ctx sdk.Context, sUpdate SubmittedUpdate) {
	// always overwrite prexisting values - leave paychan logic to higher levels
	// get current queue
	q, found := k.getSubmittedUpdatesQueue(ctx)
	if !found {
		panic("SubmittedUpdatesQueue not found.")
	}
	// append ID to queue
	if !q.Contains(sUpdate.ChannelID) {
		q = append(q, sUpdate.ChannelID)
	}
	// set queue
	k.setSubmittedUpdatesQueue(ctx, q)
	// store submittedUpdate
	k.setSubmittedUpdate(ctx, sUpdate)
}
func (k Keeper) removeFromSubmittedUpdatesQueue(ctx sdk.Context, channelID ChannelID) {
	// get current queue
	q, found := k.getSubmittedUpdatesQueue(ctx)
	if !found {
		panic("SubmittedUpdatesQueue not found.")
	}
	// remove id
	q.RemoveMatchingElements(channelID)
	// set queue
	k.setSubmittedUpdatesQueue(ctx, q)
	// delete submittedUpdate
	k.deleteSubmittedUpdate(ctx, channelID)
}

func (k Keeper) getSubmittedUpdatesQueue(ctx sdk.Context) (SubmittedUpdatesQueue, bool) {
	// load from DB
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(k.getSubmittedUpdatesQueueKey())

	var suq SubmittedUpdatesQueue
	if bz == nil {
		return suq, false // TODO maybe create custom error to pass up here
	}
	// unmarshal
	k.cdc.MustUnmarshalBinary(bz, &suq)
	// return
	return suq, true
}
func (k Keeper) setSubmittedUpdatesQueue(ctx sdk.Context, suq SubmittedUpdatesQueue) {
	store := ctx.KVStore(k.storeKey)
	// marshal
	bz := k.cdc.MustMarshalBinary(suq)
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
	key := k.getSubmittedUpdateKey(sUpdate.ChannelID)
	store.Set(key, bz) // panics if something goes wrong
}

func (k Keeper) deleteSubmittedUpdate(ctx sdk.Context, channelID ChannelID) {
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
	key := k.getChannelKey(channel.ID)
	store.Set(key, bz) // panics if something goes wrong
}

func (k Keeper) deleteChannel(ctx sdk.Context, channelID ChannelID) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(k.getChannelKey(channelID))
	// TODO does this have return values? What happens when key doesn't exist?
}

func (k Keeper) getNewChannelID(ctx sdk.Context) ChannelID {
	// get last channel ID
	var lastID ChannelID
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(k.getLastChannelIDKey())
	if bz == nil {
		lastID = -1 // TODO is just setting to zero if uninitialized ok?
	} else {
		k.cdc.MustUnmarshalBinary(bz, &lastID)
	}
	// increment to create new one
	newID := lastID + 1
	bz = k.cdc.MustMarshalBinary(newID)
	// set last channel id again
	store.Set(k.getLastChannelIDKey(), bz)
	// return
	return newID
}

func (k Keeper) getChannelKey(channelID ChannelID) []byte {
	return []byte(fmt.Sprintf("channel:%d", channelID))
}
func (k Keeper) getLastChannelIDKey() []byte {
	return []byte("lastChannelID")
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
