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

type VaultRecords []VaultRecord

type VaultShareRecords []VaultShareRecord

// NewVaultShareRecord returns a new VaultShareRecord with the provided supplied
// coins.
func NewVaultShareRecord(depositor sdk.AccAddress, supplied ...sdk.Coin) VaultShareRecord {
	return VaultShareRecord{
		Depositor:      depositor,
		AmountSupplied: sdk.NewCoins(supplied...),
	}
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
