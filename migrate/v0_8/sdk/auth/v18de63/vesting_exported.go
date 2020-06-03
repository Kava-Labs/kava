package v18de63

// Note: interfaces have had methods removed as they're not needed for unmarshalling genesis.json
// This allows account types to be copy and pasted into this package without all their methods.

// VestingAccount defines an account type that vests coins via a vesting schedule.
type VestingAccount interface {
	Account
}
