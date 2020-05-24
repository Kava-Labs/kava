package v18de63

// Note: interfaces have had methods removed as they're not needed for unmarshalling genesis.json
// This allows account types to be copy and pasted into this package without all their methods.

// Account is an interface used to store coins at a given address within state.
// It presumes a notion of sequence numbers for replay protection,
// a notion of account numbers for replay protection for previously pruned accounts,
// and a pubkey for authentication purposes.
//
// Many complex conditions can be used in the concrete struct which implements Account.
type Account interface {
	// GetAddress() sdk.AccAddress
	// SetAddress(sdk.AccAddress) error // errors if already set.

	// GetPubKey() crypto.PubKey // can return nil.
	// SetPubKey(crypto.PubKey) error

	// GetAccountNumber() uint64
	// SetAccountNumber(uint64) error

	// GetSequence() uint64
	// SetSequence(uint64) error

	// GetCoins() sdk.Coins
	// SetCoins(sdk.Coins) error

	// // Calculates the amount of coins that can be sent to other accounts given
	// // the current time.
	// SpendableCoins(blockTime time.Time) sdk.Coins

	// // Ensure that account implements stringer
	// String() string
}

// GenesisAccounts defines a slice of GenesisAccount objects
type GenesisAccounts []GenesisAccount

// GenesisAccount defines a genesis account that embeds an Account with validation capabilities.
type GenesisAccount interface {
	Account
	// Validate() error
}
