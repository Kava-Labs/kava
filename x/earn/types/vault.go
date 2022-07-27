package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
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
	if err := sdk.ValidateDenom(vr.Denom); err != nil {
		return sdkerrors.Wrap(ErrInvalidVaultDenom, err.Error())
	}

	if vr.TotalSupply.Denom != vr.Denom {
		return fmt.Errorf(
			"total supply denom %v does not match vault record denom %v",
			vr.TotalSupply.Denom,
			vr.Denom,
		)
	}

	if err := vr.TotalSupply.Validate(); err != nil {
		return fmt.Errorf("vault total supply is invalid: %w", err)
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
	if err := sdk.ValidateDenom(a.Denom); err != nil {
		return sdkerrors.Wrap(ErrInvalidVaultDenom, err.Error())
	}

	if a.VaultStrategy == STRATEGY_TYPE_UNSPECIFIED {
		return ErrInvalidVaultStrategy
	}

	return nil
}

type VaultShares []VaultShare

func (v VaultShares) AmountOf(denom string) sdk.Int {
	for _, s := range v {
		if s.Denom == denom {
			return s.Amount
		}
	}

	return sdk.ZeroInt()
}

func (v VaultShares) IsZero() bool {
	for _, s := range v {
		// If any amount is non-zero, false
		if !s.Amount.IsZero() {
			return false
		}
	}

	return true
}
