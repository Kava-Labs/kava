package v0_16

import (
	v015issuance "github.com/kava-labs/kava/x/issuance/legacy/v0_15"
	v016issuance "github.com/kava-labs/kava/x/issuance/types"
)

func migrateParams(params v015issuance.Params) v016issuance.Params {
	assets := make([]v016issuance.Asset, len(params.Assets))
	for i, asset := range params.Assets {
		blockedAddresses := make([]string, len(asset.BlockedAddresses))
		for i, addr := range asset.BlockedAddresses {
			blockedAddresses[i] = addr.String()
		}
		assets[i] = v016issuance.Asset{
			Owner:            asset.Owner.String(),
			Denom:            asset.Denom,
			BlockedAddresses: blockedAddresses,
			Paused:           asset.Paused,
			Blockable:        asset.Blockable,
			RateLimit: v016issuance.RateLimit{
				Active:     asset.RateLimit.Active,
				Limit:      asset.RateLimit.Limit,
				TimePeriod: asset.RateLimit.TimePeriod,
			},
		}
	}
	return v016issuance.Params{Assets: assets}
}

func migrateSupplies(oldSupplies v015issuance.AssetSupplies) []v016issuance.AssetSupply {
	supplies := make([]v016issuance.AssetSupply, len(oldSupplies))
	for i, supply := range oldSupplies {
		supplies[i] = v016issuance.AssetSupply{
			CurrentSupply: supply.CurrentSupply,
			TimeElapsed:   supply.TimeElapsed,
		}
	}
	return supplies
}

// Migrate converts v0.15 issuance state and returns it in v0.16 format
func Migrate(oldState v015issuance.GenesisState) *v016issuance.GenesisState {
	return &v016issuance.GenesisState{
		Params:   migrateParams(oldState.Params),
		Supplies: migrateSupplies(oldState.Supplies),
	}
}
