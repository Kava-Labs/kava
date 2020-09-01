package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// AssetSupply contains information about an asset's rate-limited supply (the total supply of the asset is tracked in the top-level supply module)
type AssetSupply struct {
	CurrentSupply sdk.Coin      `json:"current_supply"  yaml:"current_supply"`
	TimeElapsed   time.Duration `json:"time_elapsed" yaml:"time_elapsed"`
}

// NewAssetSupply initializes a new AssetSupply
func NewAssetSupply(currentSupply sdk.Coin, timeElapsed time.Duration) AssetSupply {
	return AssetSupply{
		CurrentSupply: currentSupply,
		TimeElapsed:   timeElapsed,
	}
}

// Validate performs a basic validation of an asset supply fields.
func (a AssetSupply) Validate() error {
	if !a.CurrentSupply.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "outgoing supply %s", a.CurrentSupply)
	}
	return nil
}

// String implements stringer
func (a AssetSupply) String() string {
	return fmt.Sprintf(`
	asset supply:
		Current supply:     %s
		Time elapsed: %s
		`,
		a.CurrentSupply, a.TimeElapsed)
}

// GetDenom getter method for the denom of the asset supply
func (a AssetSupply) GetDenom() string {
	return a.CurrentSupply.Denom
}

// AssetSupplies is a slice of AssetSupply
type AssetSupplies []AssetSupply

// TODO copy over supply tests from bep3
