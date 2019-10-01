package types

import (
	"bytes"

	"github.com/cosmos/cosmos-sdk/x/auth/exported"
)

// GenesisState - all auth state that must be provided at genesis
type GenesisState struct {
	Accounts exported.GenesisAccounts `json:"accounts" yaml:"accounts"`
}

// NewGenesisState - Create a new genesis state
func NewGenesisState(accounts exported.GenesisAccounts) GenesisState {
	return GenesisState{
		Accounts: accounts,
	}
}

// DefaultGenesisState - Return a default genesis state
func DefaultGenesisState() GenesisState {
	return NewGenesisState(exported.GenesisAccounts{})
}

// Equal checks whether two gov GenesisState structs are equivalent
func (data GenesisState) Equal(data2 GenesisState) bool {
	b1 := ModuleCdc.MustMarshalBinaryBare(data)
	b2 := ModuleCdc.MustMarshalBinaryBare(data2)
	return bytes.Equal(b1, b2)
}

// IsEmpty returns true if a GenesisState is empty
func (data GenesisState) IsEmpty() bool {
	return data.Equal(GenesisState{})
}

// ValidateGenesis returns nil because accounts are validated by auth
func ValidateGenesis(data GenesisState) error {
	return nil
}
