package types

type VaultRecords []VaultRecord

type VaultShareRecords []VaultShareRecord

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
