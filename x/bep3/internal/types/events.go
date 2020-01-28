package types

// bep3 module event types
const (
	// TODO: Create your event types
	// EventType<Action>    		= "action"
	EventTypeCreateHtlt = "create_htlt"

	// TODO: Create keys fo your events, the values will be derivided from the msg
	// AttributeKeyAddress  		= "address"
	AttributeKeyHtltID     = "htlt_id"
	AttributeKeyFrom       = "htlt_from"
	AttributeKeyTo         = "htlt_to"
	AttributeKeyCoinDenom  = "coin_denom"
	AttributeKeyCoinAmount = "coin_amount"
	// TODO: Some events may not have values for that reason you want to emit that something happened.
	// AttributeValueDoubleSign = "double_sign"

	AttributeValueCategory = ModuleName
)
