package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewDelegationHolder returns a new DelegationHolder
func NewDelegationHolder(macc sdk.AccAddress, validator sdk.ValAddress) DelegationHolder {
	return DelegationHolder{
		ModuleAccount: macc,
		Validator:     validator,
	}
}

// Validate DelegationHolder
func (d DelegationHolder) Validate() error {
	if d.ModuleAccount.Empty() {
		return fmt.Errorf("module account cannot be empty")
	}
	if d.Validator.Empty() {
		return fmt.Errorf("validator cannot be empty")
	}

	return nil
}

// DelegationHolders is a slice of DelegationHolder
type DelegationHolders []DelegationHolder

// Validate validates DelegationHolders
func (ds DelegationHolders) Validate() error {
	delegationHolderMaccMap := make(map[string]DelegationHolder)
	delegationHolderValidatorMap := make(map[string]DelegationHolder)
	for _, d := range ds {
		if err := d.Validate(); err != nil {
			return err
		}

		// duplicate module account addresses are invalid
		dupMacc, ok := delegationHolderMaccMap[d.ModuleAccount.String()]
		if ok {
			return fmt.Errorf("duplicate module account: %s\n%s", d, dupMacc)
		}
		delegationHolderMaccMap[d.ModuleAccount.String()] = d

		// duplicate validator addresses are invalid
		dupVal, ok := delegationHolderValidatorMap[d.Validator.String()]
		if ok {
			return fmt.Errorf("duplicate validator: %s\n%s", d, dupVal)
		}
		delegationHolderValidatorMap[d.Validator.String()] = d
	}
	return nil
}
