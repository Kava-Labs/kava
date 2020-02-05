package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/params"
	// "github.com/kava-labs/kava/x/bep3/types"
)

// Parameter keys
var (
	KeyMinimumLockTime = []byte("MinimumLockTime")
	KeyChainParams     = []byte("Chains")
	// TODO: validate this time as reasonable
	DefaultHTLTStartingID                = uint64(1)
	DefaultRelayer                       = sdk.AccAddress{}
	DefaultMinimumLockTime time.Duration = 24 * time.Hour
	DefaultChainParams                   = ChainParams{}
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
	ChainID         string         `json:"chain_id" yaml:"chain_id"` // blockchain ID
	Deputy          sdk.AccAddress `json:"deputy" yaml:"deputy"`
	MaximumLockTime int            `json:"maximum_lock_time" yaml:"maximum_lock_time"`
	SupportedAssets AssetParams    `json:"chain_assets" yaml:"chain_assets"` // list of supported assets
}

// String implements fmt.Stringer
func (cp ChainParam) String() string {
	return fmt.Sprintf(`Chain:
	Chain ID: %s
	Deputy: %s
	Maximum Lock Time: %d
	Supported assets: %s`,
		cp.ChainID, cp.Deputy, cp.MaximumLockTime, cp.SupportedAssets)
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
	Symbol string  `json:"symbol" yaml:"symbol"`
	CoinID string  `json:"coin_id" yaml:"coin_id"` // internationally recognized coin ID
	Limit  sdk.Int `json:"limit" yaml:"limit"`     // asset limit
	Active bool    `json:"active" yaml:"active"`
}

// String implements fmt.Stringer
func (ap AssetParam) String() string {
	return fmt.Sprintf(`Asset:
	Symbol: %s
	Coin ID: %s
	Limit: %s
	Active: %t`,
		ap.Symbol, ap.CoinID, ap.Limit, ap.Active)
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
	// if p.Relayer.Empty() {
	// 	return sdk.ErrInternal("relayer address cannot be empty")
	// }
	// if p.MinimumLockTime <= 0 {
	// 	return sdk.ErrInternal("minimum lock time must be greater than 0")
	// }
	return nil
}
