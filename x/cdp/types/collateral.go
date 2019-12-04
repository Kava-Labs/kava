package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CollateralState stores global information tied to a particular collateral type.
type CollateralState struct {
	Denom     string  `json:"denom" yaml:"denom"`           // Type of collateral
	TotalDebt sdk.Coins `json:"total_debt" yaml:"total_debt"` // total debt collateralized by a this coin type
	AccumulatedFees sdk.Coins // Ignoring fees for now
}