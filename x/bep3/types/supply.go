package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// AssetSupply contains information about an asset's supply
type AssetSupply struct {
	IncomingSupply sdk.Coin `json:"incoming_supply"  yaml:"incoming_supply"`
	OutgoingSupply sdk.Coin `json:"outgoing_supply"  yaml:"outgoing_supply"`
	CurrentSupply  sdk.Coin `json:"current_supply"  yaml:"current_supply"`
	SupplyLimit    sdk.Coin `json:"supply_limit"  yaml:"supply_limit"`
}

// NewAssetSupply initializes a new AssetSupply
func NewAssetSupply(incomingSupply, outgoingSupply, currentSupply, supplyLimit sdk.Coin) AssetSupply {
	return AssetSupply{
		IncomingSupply: incomingSupply,
		OutgoingSupply: outgoingSupply,
		CurrentSupply:  currentSupply,
		SupplyLimit:    supplyLimit,
	}
}

// Validate performs a basic validation of an asset supply fields.
func (a AssetSupply) Validate() error {
	if !a.IncomingSupply.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "incoming supply %s", a.IncomingSupply)
	}
	if !a.OutgoingSupply.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "outgoing supply %s", a.OutgoingSupply)
	}
	if !a.CurrentSupply.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "current supply %s", a.CurrentSupply)
	}
	if !a.SupplyLimit.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "supply limit %s", a.SupplyLimit)
	}
	denom := a.CurrentSupply.Denom
	if (a.IncomingSupply.Denom != denom) ||
		(a.OutgoingSupply.Denom != denom) ||
		(a.SupplyLimit.Denom != denom) {
		return fmt.Errorf("asset supply denoms do not match %s %s %s %s", a.CurrentSupply.Denom, a.IncomingSupply.Denom, a.OutgoingSupply.Denom, a.SupplyLimit.Denom)
	}
	return nil
}

// Equal returns if two asset supplies are equal
func (a AssetSupply) Equal(b AssetSupply) bool {
	return (a.SupplyLimit.IsEqual(b.SupplyLimit) &&
		a.IncomingSupply.IsEqual(b.IncomingSupply) &&
		a.CurrentSupply.IsEqual(b.CurrentSupply) &&
		a.OutgoingSupply.IsEqual(b.OutgoingSupply))
}

// String implements stringer
func (a AssetSupply) String() string {
	return fmt.Sprintf(`
	asset supply:
		Incoming supply:    %s
		Outgoing supply:    %s
		Current supply:     %s
		Supply limit:       %s
		`,
		a.IncomingSupply, a.OutgoingSupply, a.CurrentSupply, a.SupplyLimit)
}

// AssetSupplies is a slice of AssetSupply
type AssetSupplies []AssetSupply
