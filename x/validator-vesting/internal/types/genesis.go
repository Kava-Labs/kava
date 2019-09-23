package types

import (
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
