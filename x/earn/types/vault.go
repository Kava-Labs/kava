package types

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewVaultRecord returns a new VaultRecord with 0 supply.
func NewVaultRecord(vaultDenom string, amount sdk.Dec) VaultRecord {
	return VaultRecord{
		TotalShares: NewVaultShare(vaultDenom, amount),
	}
}

// Validate returns an error if a VaultRecord is invalid.
func (vr *VaultRecord) Validate() error {
	return vr.TotalShares.Validate()
}

// VaultRecords is a slice of VaultRecord.
type VaultRecords []VaultRecord

// Validate returns an error if a slice of VaultRecords is invalid.
func (vrs VaultRecords) Validate() error {
	denoms := make(map[string]bool)

	for _, vr := range vrs {
		if err := vr.Validate(); err != nil {
			return err
		}

		if denoms[vr.TotalShares.Denom] {
			return fmt.Errorf("duplicate vault denom %s", vr.TotalShares.Denom)
		}

		denoms[vr.TotalShares.Denom] = true
	}

	return nil
}

// NewVaultShareRecord returns a new VaultShareRecord with the provided supplied
// coins.
func NewVaultShareRecord(depositor sdk.AccAddress, shares VaultShares) VaultShareRecord {
	return VaultShareRecord{
		Depositor: depositor,
		Shares:    shares,
	}
}

// Validate returns an error if an VaultShareRecord is invalid.
func (vsr VaultShareRecord) Validate() error {
	if vsr.Depositor.Empty() {
		return fmt.Errorf("depositor is empty")
	}

	if err := vsr.Shares.Validate(); err != nil {
		return fmt.Errorf("invalid vault share record shares: %w", err)
	}

	return nil
}

// VaultShareRecords is a slice of VaultShareRecord.
type VaultShareRecords []VaultShareRecord

// Validate returns an error if a slice of VaultRecords is invalid.
func (vsrs VaultShareRecords) Validate() error {
	addrs := make(map[string]bool)

	for _, vr := range vsrs {
		if err := vr.Validate(); err != nil {
			return err
		}

		if _, found := addrs[vr.Depositor.String()]; found {
			return fmt.Errorf("duplicate address %s", vr.Depositor.String())
		}

		addrs[vr.Depositor.String()] = true
	}

	return nil
}

// NewAllowedVault returns a new AllowedVault with the given values.
func NewAllowedVault(
	denom string,
	strategyTypes StrategyTypes,
	isPrivateVault bool,
	allowedDepositors []sdk.AccAddress,
) AllowedVault {
	return AllowedVault{
		Denom:             denom,
		Strategies:        strategyTypes,
		IsPrivateVault:    isPrivateVault,
		AllowedDepositors: allowedDepositors,
	}
}

// Validate returns an error if the AllowedVault is invalid
func (a *AllowedVault) Validate() error {
	if err := sdk.ValidateDenom(a.Denom); err != nil {
		return errorsmod.Wrap(ErrInvalidVaultDenom, err.Error())
	}

	// Private -> 1+ allowed depositors
	// Non-private -> 0 allowed depositors
	if a.IsPrivateVault && len(a.AllowedDepositors) == 0 {
		return fmt.Errorf("private vaults require non-empty AllowedDepositors")
	}

	if !a.IsPrivateVault && len(a.AllowedDepositors) > 0 {
		return fmt.Errorf("non-private vaults cannot have any AllowedDepositors")
	}

	return a.Strategies.Validate()
}

// IsStrategyAllowed returns true if the given strategy type is allowed for the
// vault.
func (a *AllowedVault) IsStrategyAllowed(strategy StrategyType) bool {
	for _, s := range a.Strategies {
		if s == strategy {
			return true
		}
	}

	return false
}

// IsAccountAllowed returns true if the given account is allowed to deposit into
// the vault.
func (a *AllowedVault) IsAccountAllowed(account sdk.AccAddress) bool {
	// Anyone can deposit to non-private vaults
	if !a.IsPrivateVault {
		return true
	}

	for _, addr := range a.AllowedDepositors {
		if addr.Equals(account) {
			return true
		}
	}

	return false
}

// AllowedVaults is a slice of AllowedVault.
type AllowedVaults []AllowedVault

// Validate returns an error if the AllowedVaults is invalid.
func (a AllowedVaults) Validate() error {
	denoms := make(map[string]bool)

	for _, v := range a {
		if err := v.Validate(); err != nil {
			return err
		}

		if denoms[v.Denom] {
			return fmt.Errorf("duplicate vault denom %s", v.Denom)
		}

		denoms[v.Denom] = true
	}

	return nil
}
