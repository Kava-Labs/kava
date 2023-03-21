package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// NewAPY returns a new instance of APY
func NewAPY(collateralType string, apy sdk.Dec) Apy {
	return Apy{
		CollateralType: collateralType,
		Apy:            apy,
	}
}

// APYs is a slice of APY
type APYs []Apy
