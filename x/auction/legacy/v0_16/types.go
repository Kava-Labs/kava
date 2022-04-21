package v0_16

import (
	types "github.com/cosmos/cosmos-sdk/codec/types"
)

// GenesisAuctions is a slice of genesis auctions.
type GenesisAuctions []GenesisAuction

// InterfaceRegistry returns a registry of interfaces for
// all concrete v0_16 auction types
func InterfaceRegistry() types.InterfaceRegistry {
	registry := types.NewInterfaceRegistry()
	registry.RegisterInterface(
		"kava.auction.v1beta1.Auction",
		(*Auction)(nil),
		&SurplusAuction{},
		&DebtAuction{},
		&CollateralAuction{},
	)

	registry.RegisterInterface(
		"kava.auction.v1beta1.GenesisAuction",
		(*GenesisAuction)(nil),
		&SurplusAuction{},
		&DebtAuction{},
		&CollateralAuction{},
	)

	return registry
}
