package paychan

import (
	sdk "github.com/cosmos/cosmos-sdk/types"          
)

////////

// probably want to convert this to a general purpose "state"
struct Paychan {
	sender sdk.Address
	receiver sdk.Address
	id integer
	balance sdk.Coins
}


/////////////

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

/////////////// CreatePayChan
// find a less confusing name
type MsgCreate struct {
	// maybe just wrap a paychan struct
	sender sdk.Address
	receiver sdk.Address
	amount sdk.Coins
}

func (msg CreatMsg) NewMsgCreate() MsgCreate {
	return MsgCreate{ }
}

func (msg MsgCreate) Type() string { return "paychan" }

func (msg MsgCreate) GetSigners() []sdk.Address {
	// sender
	//return []sdk.Address{msg.sender}
}

func (msg MsgCreate) GetSignBytes() []byte {
	
}

func (msg MsgCreate) ValidateBasic() sdk.Error {
	// verify msg as much as possible without using external information (such as account balance)
	// are all fields present
	// are all fields valid
	// maybe check if sender and receiver is different
}

/////////////////
type MsgClose struct {
	// have to include sender and receiver in msg explicitly (rather than just universal paychanID)
	//  this gives ability to verify signatures with no external information
	sender sdk.Address
	receiver sdk.Address
	id integer
	receiverAmount sdk.Coins // amount the receiver should get - sender amount implicit with paychan balance
}

func (msg MsgClose) NewMsgClose( args... ) MsgClose {
	return MsgClose{ args... }
}

func (msg MsgClose) Type() string { return "paychan" }

func (msg MsgClose) GetSigners() []sdk.Address {
	// sender and receiver
}

func (msg MsgClose) GetSignBytes() []byte {
	
}

func (msg MsgClose) ValidateBasic() sdk.Error {
	return msg.IBCPacket.ValidateBasic()
}

