package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/x/params"
)

// Parameter keys
var (
	KeyMinLockTime     = []byte("MinLockTime")
	KeyMaxLockTime     = []byte("MaxLockTime")
	KeySupportedAssets = []byte("SupportedAssets")

	AbsoluteMaximumLockTime int64 = 10000
	DefaultMinLockTime      int64 = 10
	DefaultMaxLockTime      int64 = 1000
	DefaultSupportedAssets        = AssetParams{AssetParam{Denom: "kava", CoinID: "459", Limit: 1, Active: false}}
)

// Params governance parameters for bep3 module
type Params struct {
	MinLockTime     int64       `json:"min_lock_time" yaml:"min_lock_time"`       // HTLT minimum lock time
	MaxLockTime     int64       `json:"max_lock_time" yaml:"max_lock_time"`       // HTLT maximum lock time
	SupportedAssets AssetParams `json:"supported_assets" yaml:"supported_assets"` // supported assets
}

// String implements fmt.Stringer
func (p Params) String() string {
	return fmt.Sprintf(`Params:
	Min lock time: %d,
	Max lock time: %d,
	Supported assets: %s`,
		p.MinLockTime, p.MaxLockTime, p.SupportedAssets)
}

// NewParams returns a new params object
func NewParams(minLockTime int64, maxLockTime int64, supportedAssets AssetParams) Params {
	return Params{
		MinLockTime:     minLockTime,
		MaxLockTime:     maxLockTime,
		SupportedAssets: supportedAssets,
	}
}

// DefaultParams returns default params for bep3 module
func DefaultParams() Params {
	return NewParams(DefaultMinLockTime, DefaultMaxLockTime, DefaultSupportedAssets)
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
	Limit: %d
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
		{Key: KeyMinLockTime, Value: &p.MinLockTime},
		{Key: KeyMaxLockTime, Value: &p.MaxLockTime},
		{Key: KeySupportedAssets, Value: &p.SupportedAssets},
	}
}

// Validate ensure that params have valid values
func (p Params) Validate() error {
	if p.MinLockTime <= 0 {
		return fmt.Errorf("minimum lock time must be greater than 0")
	}
	if p.MaxLockTime <= 0 {
		return fmt.Errorf("minimum lock time must be greater than 0")
	}
	if p.MinLockTime >= p.MaxLockTime {
		return fmt.Errorf("maximum lock time must be greater than minimum lock time")
	}
	if p.MaxLockTime > AbsoluteMaximumLockTime {
		return fmt.Errorf(fmt.Sprintf("maximum lock time cannot be longer than %d", AbsoluteMaximumLockTime))
	}
	coinIDs := make(map[string]bool)
	for _, asset := range p.SupportedAssets {
		if len(asset.Denom) == 0 {
			return fmt.Errorf("asset denom cannot be empty")
		}
		if len(asset.CoinID) == 0 {
			return fmt.Errorf(fmt.Sprintf("asset %s cannot have an empty coin id", asset.Denom))
		}
		if coinIDs[asset.CoinID] {
			return fmt.Errorf(fmt.Sprintf("asset %s cannot have duplicate coin id %s", asset.Denom, asset.CoinID))
		}
		coinIDs[asset.CoinID] = true
		if asset.Limit <= 0 {
			return fmt.Errorf(fmt.Sprintf("asset %s must have limit greater than 0", asset.Denom))
		}
	}

	return nil
}
