package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// AssetSupply contains information about an asset's supply
type AssetSupply struct {
	Denom          string   `json:"denom"  yaml:"denom"`
	IncomingSupply sdk.Coin `json:"incoming_supply"  yaml:"incoming_supply"`
	OutgoingSupply sdk.Coin `json:"outgoing_supply"  yaml:"outgoing_supply"`
	CurrentSupply  sdk.Coin `json:"current_supply"  yaml:"current_supply"`
	SupplyLimit    sdk.Coin `json:"supply_limit"  yaml:"supply_limit"`
}

// NewAssetSupply initializes a new AssetSupply
func NewAssetSupply(denom string, incomingSupply, outgoingSupply, currentSupply, supplyLimit sdk.Coin) AssetSupply {
	return AssetSupply{
		Denom:          denom,
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
	if !a.Limit.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "limit %s", a.Limit)
	}
	return sdk.ValidateDenom(a.Denom)
}

// String implements stringer
func (a AssetSupply) String() string {
	return fmt.Sprintf(`
	%s supply:
		Incoming supply:    %s
		Outgoing supply:    %s
		Current supply:     %s
		Supply limit:       %s
		`,
		a.Denom, a.IncomingSupply, a.OutgoingSupply, a.CurrentSupply, a.SupplyLimit)
}

// AssetSupplies is a slice of AssetSupply
type AssetSupplies []AssetSupply
