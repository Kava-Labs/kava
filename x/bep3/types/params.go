package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// Parameter keys
var (
	KeyBnbDeputyAddress = []byte("BnbDeputyAddress")
	KeyMinBlockLock     = []byte("MinBlockLock")
	KeyMaxBlockLock     = []byte("MaxBlockLock")
	KeySupportedAssets  = []byte("SupportedAssets")

	AbsoluteMaximumBlockLock int64 = 10000
	AbsoluteMinimumBlockLock int64 = 50
	DefaultMinBlockLock      int64 = 80
	DefaultMaxBlockLock      int64 = 600
	DefaultSupportedAssets         = AssetParams{AssetParam{Denom: "kava", CoinID: "459", Limit: sdk.NewInt(1), Active: false}}
)

// Params governance parameters for bep3 module
type Params struct {
	BnbDeputyAddress sdk.AccAddress `json:"bnb_deputy_address" yaml:"bnb_deputy_address"` // Bnbchain deputy address
	MinBlockLock     int64          `json:"min_block_lock" yaml:"min_block_lock"`         // AtomicSwap minimum block lock
	MaxBlockLock     int64          `json:"max_block_lock" yaml:"max_block_lock"`         // AtomicSwap maximum block lock
	SupportedAssets  AssetParams    `json:"supported_assets" yaml:"supported_assets"`     // supported assets
}

// String implements fmt.Stringer
func (p Params) String() string {
	return fmt.Sprintf(`Params:
	Bnbchain deputy address: %s,
	Min block lock: %d,
	Max block lock: %d,
	Supported assets: %s`,
		p.BnbDeputyAddress.String(), p.MinBlockLock, p.MaxBlockLock, p.SupportedAssets)
}

// NewParams returns a new params object
func NewParams(bnbDeputyAddress sdk.AccAddress, minBlockLock, maxBlockLock int64, supportedAssets AssetParams) Params {
	return Params{
		BnbDeputyAddress: bnbDeputyAddress,
		MinBlockLock:     minBlockLock,
		MaxBlockLock:     maxBlockLock,
		SupportedAssets:  supportedAssets,
	}
}

// DefaultParams returns default params for bep3 module
func DefaultParams() Params {
	defaultBnbDeputyAddress, _ := sdk.AccAddressFromBech32("kava1xy7hrjy9r0algz9w3gzm8u6mrpq97kwta747gj")
	return NewParams(defaultBnbDeputyAddress, DefaultMinBlockLock, DefaultMaxBlockLock, DefaultSupportedAssets)
}

// AssetParam governance parameters for each asset within a supported chain
type AssetParam struct {
	Denom  string  `json:"denom" yaml:"denom"`     // name of the asset
	CoinID string  `json:"coin_id" yaml:"coin_id"` // internationally recognized coin ID
	Limit  sdk.Int `json:"limit" yaml:"limit"`     // asset supply limit
	Active bool    `json:"active" yaml:"active"`   // denotes if asset is available or paused
}

// String implements fmt.Stringer
func (ap AssetParam) String() string {
	return fmt.Sprintf(`Asset:
	Denom: %s
	Coin ID: %s
	Limit: %s
	Active: %t`,
		ap.Denom, ap.CoinID, ap.Limit.String(), ap.Active)
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
		{Key: KeyBnbDeputyAddress, Value: &p.BnbDeputyAddress},
		{Key: KeyMinBlockLock, Value: &p.MinBlockLock},
		{Key: KeyMaxBlockLock, Value: &p.MaxBlockLock},
		{Key: KeySupportedAssets, Value: &p.SupportedAssets},
	}
}

// Validate ensure that params have valid values
func (p Params) Validate() error {
	if p.MinBlockLock < AbsoluteMinimumBlockLock {
		return fmt.Errorf(fmt.Sprintf("minimum block lock cannot be less than %d", AbsoluteMinimumBlockLock))
	}
	if p.MinBlockLock >= p.MaxBlockLock {
		return fmt.Errorf("maximum block lock must be greater than minimum block lock")
	}
	if p.MaxBlockLock > AbsoluteMaximumBlockLock {
		return fmt.Errorf(fmt.Sprintf("maximum block lock cannot be greater than %d", AbsoluteMaximumBlockLock))
	}

	coinIDs := make(map[string]bool)
	coinDenoms := make(map[string]bool)
	for _, asset := range p.SupportedAssets {
		if len(asset.Denom) == 0 {
			return fmt.Errorf("asset denom cannot be empty")
		}
		if len(asset.CoinID) == 0 {
			return fmt.Errorf(fmt.Sprintf("asset %s cannot have an empty coin id", asset.Denom))
		}
		if !asset.Limit.IsPositive() {
			return fmt.Errorf(fmt.Sprintf("asset %s must have a positive supply limit", asset.Denom))
		}
		if coinDenoms[asset.Denom] {
			return fmt.Errorf(fmt.Sprintf("asset %s cannot have duplicate denom", asset.Denom))
		}
		coinDenoms[asset.Denom] = true
		if coinIDs[asset.CoinID] {
			return fmt.Errorf(fmt.Sprintf("asset %s cannot have duplicate coin id %s", asset.Denom, asset.CoinID))
		}
		coinIDs[asset.CoinID] = true
	}
	return nil
}
