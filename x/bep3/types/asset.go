package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
