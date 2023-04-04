package v0_15

import (
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName The name that will be used throughout the module
	ModuleName = "issuance"
)

// GenesisState is the state that must be provided at genesis for the issuance module
type GenesisState struct {
	Params   Params        `json:"params" yaml:"params"`
	Supplies AssetSupplies `json:"supplies" yaml:"supplies"`
}

// Params governance parameters for the issuance module
type Params struct {
	Assets Assets `json:"assets" yaml:"assets"`
}

// Assets slice of Asset
type Assets []Asset

// Asset type for assets in the issuance module
type Asset struct {
	Owner            sdk.AccAddress   `json:"owner" yaml:"owner"`
	Denom            string           `json:"denom" yaml:"denom"`
	BlockedAddresses []sdk.AccAddress `json:"blocked_addresses" yaml:"blocked_addresses"`
	Paused           bool             `json:"paused" yaml:"paused"`
	Blockable        bool             `json:"blockable" yaml:"blockable"`
	RateLimit        RateLimit        `json:"rate_limit" yaml:"rate_limit"`
}

// RateLimit parameters for rate-limiting the supply of an issued asset
type RateLimit struct {
	Active     bool          `json:"active" yaml:"active"`
	Limit      sdkmath.Int   `json:"limit" yaml:"limit"`
	TimePeriod time.Duration `json:"time_period" yaml:"time_period"`
}

// AssetSupplies is a slice of AssetSupply
type AssetSupplies []AssetSupply

// AssetSupply contains information about an asset's rate-limited supply (the total supply of the asset is tracked in the top-level supply module)
type AssetSupply struct {
	CurrentSupply sdk.Coin      `json:"current_supply"  yaml:"current_supply"`
	TimeElapsed   time.Duration `json:"time_elapsed" yaml:"time_elapsed"`
}
