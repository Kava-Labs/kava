package types

import (
	sdkmath "cosmossdk.io/math"
)

// NewAPY returns a new instance of APY
func NewAPY(collateralType string, apy sdkmath.LegacyDec) Apy {
	return Apy{
		CollateralType: collateralType,
		Apy:            apy,
	}
}

// APYs is a slice of APY
type APYs []Apy
