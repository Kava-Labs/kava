package types

// Transactions objects must fulfill the Tx
type Tx interface {
	// Gets the all the transaction's messages.
	// GetMsgs() []Msg

	// ValidateBasic does a simple and lightweight validation check that doesn't
	// require access to any other information.
	// ValidateBasic() Error
}
