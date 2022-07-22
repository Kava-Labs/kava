package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewVaultRecord returns a new VaultRecord with 0 supply.
func NewVaultRecord(vaultDenom string) VaultRecord {
	return VaultRecord{
		Denom:       vaultDenom,
		TotalSupply: sdk.NewCoin(vaultDenom, sdk.ZeroInt()),
	}
}

// Validate returns an error if a VaultRecord is invalid.
func (vr *VaultRecord) Validate() error {
	if vr.Denom == "" {
		return ErrInvalidVaultDenom
	}

	if vr.TotalSupply.Denom != vr.Denom {
		return fmt.Errorf("total supply denom does not match vault record denom: %w", ErrInvalidVaultTotalSupply)
	}

	if vr.TotalSupply.IsNegative() {
		return fmt.Errorf("vault total supply is negative: %w", ErrInvalidVaultTotalSupply)
	}

	return nil
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

		if denoms[vr.Denom] {
			return fmt.Errorf("duplicate vault denom %s", vr.Denom)
		}

		denoms[vr.Denom] = true
	}

	return nil
}

// NewVaultShareRecord returns a new VaultShareRecord with the provided supplied
// coins.
func NewVaultShareRecord(depositor sdk.AccAddress, supplied ...sdk.Coin) VaultShareRecord {
	return VaultShareRecord{
		Depositor:      depositor,
		AmountSupplied: sdk.NewCoins(supplied...),
	}
}

// Validate returns an error if an VaultShareRecord is invalid.
func (vsr VaultShareRecord) Validate() error {
	if vsr.Depositor.Empty() {
		return fmt.Errorf("depositor is empty")
	}

	if err := vsr.AmountSupplied.Validate(); err != nil {
		return fmt.Errorf("invalid vault share record amount supplied: %w", err)
	}

	return nil
}

// VaultShareRecords is a slice of VaultShareRecord.
type VaultShareRecords []VaultShareRecord

// Validate returns an error if a slice of VaultRecords is invalid.
func (vsrs VaultShareRecords) Validate() error {
	denoms := make(map[string]bool)

	for _, vr := range vsrs {
		if err := vr.Validate(); err != nil {
			return err
		}

		if denoms[vr.AmountSupplied.Denom] {
			return fmt.Errorf("duplicate vault denom %s", vr.AmountSupplied.Denom)
		}

		denoms[vr.AmountSupplied.Denom] = true
	}

	return nil
}

// NewAllowedVaults returns a new AllowedVaults with the given denom and strategy type.
func NewAllowedVault(denom string, strategyType StrategyType) AllowedVault {
	return AllowedVault{
		Denom:         denom,
		VaultStrategy: strategyType,
	}
}

type AllowedVaults []AllowedVault

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

func (a *AllowedVault) Validate() error {
	if a.Denom == "" {
		return ErrInvalidVaultDenom
	}

	if a.VaultStrategy == STRATEGY_TYPE_UNSPECIFIED {
		return ErrInvalidVaultStrategy
	}

	return nil
}
