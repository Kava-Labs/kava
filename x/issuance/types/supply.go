package types

import (
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

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
		return errorsmod.Wrapf(sdkerrors.ErrInvalidCoins, "outgoing supply %s", a.CurrentSupply)
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
