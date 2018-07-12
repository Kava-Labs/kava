package paychan

import (
	"strconv"
	sdk "github.com/cosmos/cosmos-sdk/types"          
)

// Paychan Type
// Used to represent paychan in keeper module and to serialize.
// probably want to convert this to a general purpose "state"
struct Paychan {
	sender sdk.Address
	receiver sdk.Address
	id integer
	balance sdk.Coins
}


// Message Types

// Message implement the sdk.Msg interface:

// type Msg interface {

// 	// Return the message type.
// 	// Must be alphanumeric or empty.
// 	Type() string

// 	// Get the canonical byte representation of the Msg.
// 	GetSignBytes() []byte

// 	// ValidateBasic does a simple validation check that
// 	// doesn't require access to any other information.
// 	ValidateBasic() Error

// 	// Signers returns the addrs of signers that must sign.
// 	// CONTRACT: All signatures must be present to be valid.
// 	// CONTRACT: Returns addrs in some deterministic order.
// 	GetSigners() []Address
// }


// A message to create a payment channel.
type MsgCreate struct {
	// maybe just wrap a paychan struct
	sender   sdk.Address
	receiver sdk.Address
	amount   sdk.Coins
}

// Create a new message.
// Called in client code when constructing transaction from cli args to send to the network.
// maybe just a placeholder for more advanced future functionality?
// func (msg CreatMsg) NewMsgCreate(sender sdk.Address, receiver sdk.Address, amount sdk.Coins) MsgCreate {
// 	return MsgCreate{
// 		sender
// 		receiver
// 		amount
// 	}
// }

func (msg MsgCreate) Type() string { return "paychan" }

func (msg MsgCreate) GetSignBytes() []byte {
	// TODO create msgCdc in wire.go
	b, err := msgCdc.MarshalJSON(struct {
		SenderAddr   string    `json:"sender_addr"`
		ReceiverAddr string    `json:"receiver_addr"`
		Amount       sdk.Coins `json:"amount"`
	}{
		SenderAddr:   sdk.MustBech32ifyAcc(msg.sender),
		ReceiverAddr: sdk.MustBech32ifyAcc(msg.receiver),
		Amount:       msg.amount,
	})
	if err != nil {
		panic(err)
	}
	return b
}

func (msg MsgCreate) ValidateBasic() sdk.Error {
	// Validate msg as an optimisation to avoid all validation going to keeper. It's run before the sigs are checked by the auth module.
	// Validate without external information (such as account balance)

	// check if all fields present / not 0 valued
	if len(msg.sender) == 0 {
		return sdk.ErrInvalidAddress(msg.sender.String())
	}
	if len(msg.receiver) == 0 {
		return sdk.ErrInvalidAddress(msg.receiver.String())
	}
	if len(msg.amount) == 0 {
		return sdk.ErrInvalidCoins(msg.amount.String())
	}
	// Check if coins are sorted, non zero, non negative
	if !msg.amount.IsValid() {
		return sdk.ErrInvalidCoins(msg.amount.String())
	}
	if !msg.amount.IsPositive() {
		return sdk.ErrInvalidCoins(msg.amount.String())
	}
	// TODO check if Address valid?
}

func (msg MsgCreate) GetSigners() []sdk.Address {
	// Only sender must sign to create a paychan
	return []sdk.Address{msg.sender}
}




// A message to close a payment channel.
type MsgClose struct {
	// have to include sender and receiver in msg explicitly (rather than just universal paychanID)
	//  this gives ability to verify signatures with no external information
	sender         sdk.Address
	receiver       sdk.Address
	id             integer
	receiverAmount sdk.Coins // amount the receiver should get - sender amount implicit with paychan balance
}

// func (msg MsgClose) NewMsgClose(sender sdk.Address, receiver sdk.Address, id integer, receiverAmount sdk.Coins) MsgClose {
// 	return MsgClose{
// 		sender
// 		receiver
// 		id
// 		receiverAmount
// 	}
// }

func (msg MsgClose) Type() string { return "paychan" }

func (msg MsgClose) GetSignBytes() []byte {
	// TODO create msgCdc in wire.go
	b, err := msgCdc.MarshalJSON(struct {
		SenderAddr     string    `json:"sender_addr"`
		ReceiverAddr   string    `json:"receiver_addr"`
		Id             integer   `json:"id"`
		ReceiverAmount sdk.Coins `json:"receiver_amount"`
	}{
		SenderAddr:   sdk.MustBech32ifyAcc(msg.sender),
		ReceiverAddr: sdk.MustBech32ifyAcc(msg.receiver),
		Id:           msg.id
		Amount:       msg.receiverAmount,
	})
	if err != nil {
		panic(err)
	}
	return b
}

func (msg MsgClose) ValidateBasic() sdk.Error {
	// check if all fields present / not 0 valued
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
	// TODO check if Address valid?
}

func (msg MsgClose) GetSigners() []sdk.Address {
	// Both sender and receiver must sign in order to close a channel
	retutn []sdk.Address{sender, receiver}
}

