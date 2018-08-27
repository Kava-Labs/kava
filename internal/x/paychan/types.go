package paychan

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"strconv"
)

/*  CHANNEL TYPES  */

// Used to represent a channel in the keeper module.
// Participants is limited to two as currently these are unidirectional channels.
// Last participant is designated as receiver.
type Channel struct {
	ID       ChannelID
	Participants [2]sdk.AccAddress
	Coins  sdk.Coins
}

type ChannelID int64 // TODO should this be positive only

// The data that is passed between participants as payments, and submitted to the blockchain to close a channel.
type Update struct {
	ChannelID int64
	CoinsUpdate map[sdk.AccAddress]sdk.Coins
	Sequence int64
	sig // TODO type, only sender needs to sign
}

// An update that has been submitted to the blockchain, but not yet acted on.
type SubmittedUpdate {
	Update
	executionTime int64 // BlockHeight
}

type SubmittedUpdateQueue []ChannelID
// Check if value is in queue
func (suq SubmittedChannelID) Contains(channelID ChannelID) bool {
	found := false
	for _, id := range(suq) {
		if id == channelID {
			found = true
			break
		}
	}
	return found
}
// Remove all values from queue that match argument
func (suq SubmittedUpdateQueue) RemoveMatchingElements(channelID ChannelID) {
	newSUQ := SubmittedUpdateQueue{}

	for _, id := range(suq) {
		if id != channelID {
			newSUQ = append(newSUQ, id)
		}
	}
	suq = newSUQ
}

var ChannelDisputeTime = 2000 // measured in blocks

/*  MESSAGE TYPES  */
/*
Message implement the sdk.Msg interface:
type Msg interface {

	// Return the message type.
	// Must be alphanumeric or empty.
	Type() string

	// Get the canonical byte representation of the Msg.
	GetSignBytes() []byte

	// ValidateBasic does a simple validation check that
	// doesn't require access to any other information.
	ValidateBasic() Error

	// Signers returns the addrs of signers that must sign.
	// CONTRACT: All signatures must be present to be valid.
	// CONTRACT: Returns addrs in some deterministic order.
	GetSigners() []Address
}
*/

// A message to create a payment channel.
type MsgCreate struct {
	Participants	[2]sdk.AccAddress
	Coins	sdk.Coins
}

//Create a new message.
/*
Called in client code when constructing transaction from cli args to send to the network.
maybe just a placeholder for more advanced future functionality?
func (msg CreatMsg) NewMsgCreate(sender sdk.Address, receiver sdk.Address, amount sdk.Coins) MsgCreate {
	return MsgCreate{
		sender
		receiver
		amount
	}
}
*/

func (msg MsgCreate) Type() string { return "paychan" }

func (msg MsgCreate) GetSignBytes() []byte {
	// TODO create msgCdc in wire.go
	bz, err := msgCdc.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return MustSortJSON(bz)
}

func (msg MsgCreate) ValidateBasic() sdk.Error {
	// Validate msg as an optimisation to avoid all validation going to keeper. It's run before the sigs are checked by the auth module.
	// Validate without external information (such as account balance)

	//TODO implement

	/*
	// check if all fields present / not 0 valued
	if len(msg.Sender) == 0 {
		return sdk.ErrInvalidAddress(msg.Sender.String())
	}
	if len(msg.Receiver) == 0 {
		return sdk.ErrInvalidAddress(msg.Receiver.String())
	}
	if len(msg.Amount) == 0 {
		return sdk.ErrInvalidCoins(msg.Amount.String())
	}
	// Check if coins are sorted, non zero, non negative
	if !msg.Amount.IsValid() {
		return sdk.ErrInvalidCoins(msg.Amount.String())
	}
	if !msg.Amount.IsPositive() {
		return sdk.ErrInvalidCoins(msg.Amount.String())
	}
	// TODO check if Address valid?
	*/
	return nil
}

func (msg MsgCreate) GetSigners() []sdk.Address {
	// Only sender must sign to create a paychan
	return []sdk.AccAddress{msg.Participants[0]} // select sender address
}

// A message to close a payment channel.
type MsgSubmitUpdate struct {
	Update
	submitter sdk.AccAddress
}

// func (msg MsgSubmitUpdate) NewMsgSubmitUpdate(update Update) MsgSubmitUpdate {
// 	return MsgSubmitUpdate{
// 		update
// 	}
// }

func (msg MsgSubmitUpdate) Type() string { return "paychan" }

func (msg MsgSubmitUpdate) GetSignBytes() []byte {
	// TODO create msgCdc in wire.go
	bz, err := msgCdc.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return MustSortJSON(bz)
}

func (msg MsgSubmitUpdate) ValidateBasic() sdk.Error {

	// TODO implement
	/*
	// check if all fields present / not 0 valued
	if len(msg.Sender) == 0 {
		return sdk.ErrInvalidAddress(msg.Sender.String())
	}
	if len(msg.Receiver) == 0 {
		return sdk.ErrInvalidAddress(msg.Receiver.String())
	}
	if len(msg.ReceiverAmount) == 0 {
		return sdk.ErrInvalidCoins(msg.ReceiverAmount.String())
	}
	// check id ≥ 0
	if msg.Id < 0 {
		return sdk.ErrInvalidAddress(strconv.Itoa(int(msg.Id))) // TODO implement custom errors
	}
	// Check if coins are sorted, non zero, non negative
	if !msg.ReceiverAmount.IsValid() {
		return sdk.ErrInvalidCoins(msg.ReceiverAmount.String())
	}
	if !msg.ReceiverAmount.IsPositive() {
		return sdk.ErrInvalidCoins(msg.ReceiverAmount.String())
	}
	// TODO check if Address valid?
	*/
	return nil
}

func (msg MsgSubmitUpdate) GetSigners() []sdk.Address {
	// Signing not strictly necessary as signatures contained within the channel update.
	// TODO add signature by submitting address
	return []sdk.Address{}
}
