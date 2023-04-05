package types

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewGenesisState returns a new genesis state object for the module.
func NewGenesisState(accounts []Account, params Params) *GenesisState {
	return &GenesisState{
		Accounts: accounts,
		Params:   params,
	}
}

// DefaultGenesisState returns the default genesis state for the module.
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(
		[]Account{},
		DefaultParams(),
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

	if err := gs.Params.Validate(); err != nil {
		return err
	}

	return nil
}

func NewAccount(addr sdk.AccAddress, balance sdkmath.Int) *Account {
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
