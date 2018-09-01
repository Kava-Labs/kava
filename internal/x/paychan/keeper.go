package paychan

import (
	"bytes"
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

	// Check addresses valid (Technicaly don't need to check sender address is valid as SubtractCoins does that)
	if len(sender) == 0 {
		return nil, sdk.ErrInvalidAddress(sender.String())
	}
	if len(receiver) == 0 {
		return nil, sdk.ErrInvalidAddress(receiver.String())
	}
	// check coins are sorted and positive (disallow channels with zero balance)
	if !coins.IsValid() {
		return nil, sdk.ErrInvalidCoins(coins.String())
	}
	if !coins.IsPositive() {
		return nil, sdk.ErrInvalidCoins(coins.String())
	}

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

	err := k.validateUpdate(ctx, update)
	if err != nil {
		return nil, err
	}

	q, found := k.getSubmittedUpdatesQueue(ctx)
	if !found {
		panic("SubmittedUpdatesQueue not found.") // TODO nicer custom errors
	}
	if q.Contains(update.ChannelID) {
		// Someone has previously tried to update channel
		// In bidirectional channels the new update is compared against existing and replaces it if it has a higher sequence number.

		// existingSUpdate, found := k.getSubmittedUpdate(ctx, update.ChannelID)
		// if !found {
		// 	panic("can't find element in queue that should exist")
		// }
		// k.addToSubmittedUpdatesQueue(ctx, k.applyNewUpdate(existingSUpdate, update))

		// However in unidirectional case, only the sender can close a channel this way. No clear need for them to be able to submit an update replacing a previous one they sent, so don't allow it.
		// TODO tags
		// TODO custom errors return sdk.EmptyTags(), sdk.NewError("Sender can't submit an update for channel if one has already been submitted.")
		sdk.ErrInternal("Sender can't submit an update for channel if one has already been submitted.")
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

	err := k.validateUpdate(ctx, update)
	if err != nil {
		return nil, err
	}

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
// Not needed in unidirectional case.
// func (k Keeper) applyNewUpdate(existingSUpdate SubmittedUpdate, proposedUpdate Update) SubmittedUpdate {
// 	var returnUpdate SubmittedUpdate

// 	if existingSUpdate.Sequence > proposedUpdate.Sequence {
// 		// update accepted
// 		returnUpdate = SubmittedUpdate{
// 			Update:        proposedUpdate,
// 			ExecutionTime: existingSUpdate.ExecutionTime, // FIXME any new update proposal should be subject to full dispute period from submission
// 		}
// 	} else {
// 		// update rejected
// 		returnUpdate = existingSUpdate
// 	}
// 	return returnUpdate
// }

func (k Keeper) validateUpdate(ctx sdk.Context, update Update) sdk.Error {
	// Check that channel exists
	channel, found := k.getChannel(ctx, update.ChannelID)
	if !found {
		return sdk.ErrInternal("Channel doesn't exist")
	}
	// Check the num of payout participants match channel participants
	if len(update.Payout) != len(channel.Participants) {
		return sdk.ErrInternal("Payout doesn't match number of channel participants")
	}
	// Check each coins are valid
	for _, coins := range update.Payout {
		if !coins.IsValid() {
			return sdk.ErrInternal("Payout coins aren't formatted correctly")
		}
	}
	// Check payout coins are each not negative (can be zero though)
	if !update.Payout.IsNotNegative() {
		return sdk.ErrInternal("Payout cannot be negative")
	}
	// Check payout sums to match channel.Coins
	if !channel.Coins.IsEqual(update.Payout.Sum()) {
		return sdk.ErrInternal("Payout amount doesn't match channel amount")
	}
	// Check sender signature is OK
	if !k.verifySignatures(ctx, channel, update) {
		return sdk.ErrInternal("Signature on update not valid")
	}
	return nil
}

// unsafe close channel - doesn't check if update matches existing channel TODO make safer?
func (k Keeper) closeChannel(ctx sdk.Context, update Update) (sdk.Tags, sdk.Error) {
	var err sdk.Error
	var tags sdk.Tags

	channel, _ := k.getChannel(ctx, update.ChannelID)
	// TODO check channel exists and participants matches update payout length

	// Add coins to sender and receiver
	// TODO check for possible errors first to avoid coins being half paid out?
	for i, coins := range update.Payout {
		// TODO check somewhere if coins are not negative?
		_, tags, err = k.coinKeeper.AddCoins(ctx, channel.Participants[i], coins)
		if err != nil {
			panic(err)
		}
	}

	k.deleteChannel(ctx, update.ChannelID)

	return tags, nil
}

func (k Keeper) verifySignatures(ctx sdk.Context, channel Channel, update Update) bool {
	// In non unidirectional channels there will be more than one signature to check

	signBytes := update.GetSignBytes()

	address := channel.Participants[0]
	pubKey := update.Sigs[0].PubKey
	cryptoSig := update.Sigs[0].CryptoSignature

	// Check public key submitted with update signature matches the account address
	valid := bytes.Equal(pubKey.Address(), address) &&
		// Check the signature is correct
		pubKey.VerifyBytes(signBytes, cryptoSig)
	return valid

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
// These are keyed by the IDs of their associated Channels
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
