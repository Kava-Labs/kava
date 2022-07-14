package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewVaultRecord returns a new VaultRecord with 0 supply.
func NewVaultRecord(vaultDenom string) VaultRecord {
	return VaultRecord{
		Denom:       vaultDenom,
		TotalSupply: sdk.NewCoin(vaultDenom, sdk.ZeroInt()),
	}
}

type VaultRecords []VaultRecord

type VaultShareRecords []VaultShareRecord

// NewVaultShareRecord returns a new VaultShareRecord with 0 supply.
func NewVaultShareRecord(depositor sdk.AccAddress, vaultDenom string) VaultShareRecord {
	return VaultShareRecord{
		Depositor:      depositor,
		AmountSupplied: sdk.NewCoin(vaultDenom, sdk.ZeroInt()),
	}
}

type AllowedVaults []AllowedVault

func (a AllowedVaults) Validate() error {
	for _, v := range a {
		if err := v.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (a *AllowedVault) Validate() error {
	if a.Denom == "" {
		return ErrInvalidVaultDenom
	}

	if a.VaultStrategy == STRATEGY_TYPE_UNKNOWN {
		return ErrInvalidVaultStrategy
	}

	return nil
}
