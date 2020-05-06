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
	Limit          sdk.Coin `json:"limit"  yaml:"limit"`
}

// NewAssetSupply initializes a new AssetSupply
func NewAssetSupply(denom string, incomingSupply, outgoingSupply, currentSupply, limit sdk.Coin) AssetSupply {
	return AssetSupply{
		Denom:          denom,
		IncomingSupply: incomingSupply,
		OutgoingSupply: outgoingSupply,
		CurrentSupply:  currentSupply,
		Limit:          limit,
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
	return fmt.Sprintf("Asset Supply"+
		"\n    Denom:              %s"+
		"\n    Incoming supply:    %s"+
		"\n    Outgoing supply:    %s"+
		"\n    Current supply:     %s"+
		"\n    Limit:       %s"+
		a.Denom, a.IncomingSupply, a.OutgoingSupply, a.CurrentSupply, a.Limit)
}

// AssetSupplies is a slice of AssetSupply
type AssetSupplies []AssetSupply

// String implements stringer
func (supplies AssetSupplies) String() string {
	out := ""
	for _, supply := range supplies {
		out += supply.String() + "\n"
	}
	return out
}
