package v18de63

// Note: interfaces have had methods removed as they're not needed for unmarshalling genesis.json
// This allows account types to be copy and pasted into this package without all their methods.

// VestingAccount defines an account type that vests coins via a vesting schedule.
type VestingAccount interface {
	Account

	// // Delegation and undelegation accounting that returns the resulting base
	// // coins amount.
	// TrackDelegation(blockTime time.Time, amount sdk.Coins)
	// TrackUndelegation(amount sdk.Coins)

	// GetVestedCoins(blockTime time.Time) sdk.Coins
	// GetVestingCoins(blockTime time.Time) sdk.Coins

	// GetStartTime() int64
	// GetEndTime() int64

	// GetOriginalVesting() sdk.Coins
	// GetDelegatedFree() sdk.Coins
	// GetDelegatedVesting() sdk.Coins
}
