package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// AssetSupply contains information about an asset's supply
type AssetSupply struct {
	IncomingSupply           sdk.Coin      `json:"incoming_supply"  yaml:"incoming_supply"`
	OutgoingSupply           sdk.Coin      `json:"outgoing_supply"  yaml:"outgoing_supply"`
	CurrentSupply            sdk.Coin      `json:"current_supply"  yaml:"current_supply"`
	TimeLimitedCurrentSupply sdk.Coin      `json:"time_limited_current_supply" yaml:"time_limited_current_supply"`
	TimeElapsed              time.Duration `json:"time_elapsed" yaml:"time_elapsed"`
}

// NewAssetSupply initializes a new AssetSupply
func NewAssetSupply(incomingSupply, outgoingSupply, currentSupply, timeLimitedSupply sdk.Coin, timeElapsed time.Duration) AssetSupply {
	return AssetSupply{
		IncomingSupply:           incomingSupply,
		OutgoingSupply:           outgoingSupply,
		CurrentSupply:            currentSupply,
		TimeLimitedCurrentSupply: timeLimitedSupply,
		TimeElapsed:              timeElapsed,
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
	if !a.TimeLimitedCurrentSupply.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "time-limited current supply %s", a.CurrentSupply)
	}
	denom := a.CurrentSupply.Denom
	if (a.IncomingSupply.Denom != denom) ||
		(a.OutgoingSupply.Denom != denom) ||
		(a.TimeLimitedCurrentSupply.Denom != denom) {
		return fmt.Errorf("asset supply denoms do not match %s %s %s %s", a.CurrentSupply.Denom, a.IncomingSupply.Denom, a.OutgoingSupply.Denom, a.TimeLimitedCurrentSupply.Denom)
	}
	return nil
}

// Equal returns if two asset supplies are equal
func (a AssetSupply) Equal(b AssetSupply) bool {
	return (a.IncomingSupply.IsEqual(b.IncomingSupply) &&
		a.CurrentSupply.IsEqual(b.CurrentSupply) &&
		a.OutgoingSupply.IsEqual(b.OutgoingSupply) &&
		a.TimeLimitedCurrentSupply.IsEqual(b.TimeLimitedCurrentSupply) &&
		a.TimeElapsed == b.TimeElapsed)
}

// String implements stringer
func (a AssetSupply) String() string {
	return fmt.Sprintf(`
	asset supply:
		Incoming supply:    %s
		Outgoing supply:    %s
		Current supply:     %s
		Time-limited current cupply: %s
		Time elapsed: %s
		`,
		a.IncomingSupply, a.OutgoingSupply, a.CurrentSupply, a.TimeLimitedCurrentSupply, a.TimeElapsed)
}

// AssetSupplies is a slice of AssetSupply
type AssetSupplies []AssetSupply
