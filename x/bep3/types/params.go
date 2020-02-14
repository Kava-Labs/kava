package types

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/x/params"
)

// Parameter keys
var (
	KeyChainParams = []byte("Chains")

	AbsoluteMaximumLockTime               = 720 * time.Hour // 30 days
	DefaultMinLockTime      time.Duration = 12 * time.Hour  // 12 hours
	DefaultMaxLockTime      time.Duration = 168 * time.Hour // 7 days
	DefaultChainParams                    = ChainParams{}
)

// Params governance parameters for bep3 module
type Params struct {
	Chains ChainParams `json:"chains" yaml:"chains"`
}

// String implements fmt.Stringer
func (p Params) String() string {
	return fmt.Sprintf(`Params:
	Chains: %s`,
		p.Chains)
}

// NewParams returns a new params object
func NewParams(chains ChainParams) Params {
	return Params{
		Chains: chains,
	}
}

// DefaultParams returns default params for bep3 module
func DefaultParams() Params {
	return NewParams(DefaultChainParams)
}

// ChainParam governance parameters for each chain within the bep3 module
type ChainParam struct {
	ChainID         string      `json:"chain_id" yaml:"chain_id"`                 // international blockchain identifier
	MinLockTime     int64       `json:"min_lock_time" yaml:"min_lock_time"`       // HTLT minimum lock time
	MaxLockTime     int64       `json:"max_lock_time" yaml:"max_lock_time"`       // HTLT maximum lock time
	SupportedAssets AssetParams `json:"supported_assets" yaml:"supported_assets"` // supported assets
}

// String implements fmt.Stringer
func (cp ChainParam) String() string {
	return fmt.Sprintf(`Chain:
	Chain ID: %s
	Minimum Lock Time: %d
	Maximum Lock Time: %d
	Supported assets: %s`,
		cp.ChainID, cp.MinLockTime, cp.MaxLockTime, cp.SupportedAssets)
}

// ChainParams array of ChainParam
type ChainParams []ChainParam

// String implements fmt.Stringer
func (cps ChainParams) String() string {
	out := "Chain Params\n"
	for _, cp := range cps {
		out += fmt.Sprintf("%s\n", cp)
	}
	return out
}

// AssetParam governance parameters for each asset within a supported chain
type AssetParam struct {
	Denom  string `json:"denom" yaml:"denom"`     // name of the asster
	CoinID string `json:"coin_id" yaml:"coin_id"` // internationally recognized coin ID
	Limit  int64  `json:"limit" yaml:"limit"`     // asset supply limit
	Active bool   `json:"active" yaml:"active"`   // denotes if asset is available or paused
}

// String implements fmt.Stringer
func (ap AssetParam) String() string {
	return fmt.Sprintf(`Asset:
	Denom: %s
	Coin ID: %s
	Limit: %s
	Active: %t`,
		ap.Denom, ap.CoinID, ap.Limit, ap.Active)
}

// AssetParams array of AssetParam
type AssetParams []AssetParam

// String implements fmt.Stringer
func (aps AssetParams) String() string {
	out := "Asset Params\n"
	for _, ap := range aps {
		out += fmt.Sprintf("%s\n", ap)
	}
	return out
}

// ParamKeyTable Key declaration for parameters
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// pairs of bep3 module's parameters.
// nolint
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		{Key: KeyChainParams, Value: &p.Chains},
	}
}

// Validate ensure that params have valid values
func (p Params) Validate() error {
	chainIDs := make(map[string]bool)
	for _, chain := range p.Chains {
		if len(chain.ChainID) == 0 {
			return fmt.Errorf("chain id cannot be empty")
		}
		if chainIDs[chain.ChainID] {
			return fmt.Errorf(fmt.Sprintf("cannot have duplicate chain id %s", chain.ChainID))
		}
		chainIDs[chain.ChainID] = true
		if chain.MinLockTime <= 0 {
			return fmt.Errorf("minimum lock time must be greater than 0")
		}
		if chain.MinLockTime >= chain.MaxLockTime {
			return fmt.Errorf("maximum lock time must be greater than minimum lock time")
		}
		if time.Duration(chain.MaxLockTime) > AbsoluteMaximumLockTime {
			return fmt.Errorf(fmt.Sprintf("maximum lock time cannot be longer than %d", AbsoluteMaximumLockTime))
		}
		coinIDs := make(map[string]bool)
		for _, asset := range chain.SupportedAssets {
			if len(asset.Denom) == 0 {
				return fmt.Errorf("asset denom cannot be empty")
			}
			if len(asset.CoinID) == 0 {
				return fmt.Errorf(fmt.Sprintf("asset %s cannot have an empty coin id", asset.Denom))
			}
			if coinIDs[asset.CoinID] {
				return fmt.Errorf(fmt.Sprintf("asset %s on chain %s cannot have duplicate coin id %s", asset.Denom, chain.ChainID, asset.CoinID))
			}
			coinIDs[asset.CoinID] = true
			if asset.Limit <= 0 {
				return fmt.Errorf(fmt.Sprintf("asset %s must have limit greater than 0", asset.Denom))
			}
		}
	}
	return nil
}
