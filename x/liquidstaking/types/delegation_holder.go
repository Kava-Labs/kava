package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewDelegationHolder returns a new DelegationHolder
func NewDelegationHolder(validator sdk.ValAddress) DelegationHolder {
	return DelegationHolder{
		Validator: validator,
	}
}

// Validate DelegationHolder
func (d DelegationHolder) Validate() error {
	if d.Validator.Empty() {
		return fmt.Errorf("validator cannot be empty")
	}

	return nil
}

// DelegationHolders is a slice of DelegationHolder
type DelegationHolders []DelegationHolder

// Validate validates DelegationHolders
func (ds DelegationHolders) Validate() error {
	delegationHolderValidatorMap := make(map[string]DelegationHolder)
	for _, d := range ds {
		if err := d.Validate(); err != nil {
			return err
		}

		// duplicate validator addresses are invalid
		dupVal, ok := delegationHolderValidatorMap[d.Validator.String()]
		if ok {
			return fmt.Errorf("duplicate validator: %s\n%s", d, dupVal)
		}
		delegationHolderValidatorMap[d.Validator.String()] = d
	}
	return nil
}
