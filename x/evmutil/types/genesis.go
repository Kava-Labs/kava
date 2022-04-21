package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewGenesisState returns a new genesis state object for the module.
func NewGenesisState(accounts []Account) *GenesisState {
	return &GenesisState{
		Accounts: accounts,
	}
}

// DefaultGenesisState returns the default genesis state for the module.
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(
		[]Account{},
	)
}

// Validate performs basic validation of genesis data.
func (gs GenesisState) Validate() error {
	seenAccounts := make(map[string]bool)
	for _, account := range gs.Accounts {
		if seenAccounts[account.Address.String()] {
			return fmt.Errorf("duplicate account for address %s", account.Address)
		}

		if err := account.Validate(); err != nil {
			return err
		}

		seenAccounts[account.Address.String()] = true
	}
	return nil
}

func NewAccount(addr sdk.AccAddress, balance sdk.Int) *Account {
	return &Account{
		Address: addr,
		Balance: balance,
	}
}

func (b Account) Validate() error {
	if b.Address.Empty() {
		return fmt.Errorf("address cannot be empty")
	}
	if b.Balance.IsNegative() {
		return fmt.Errorf("balance amount cannot be negative; amount: %d", b.Balance)
	}
	return nil
}
